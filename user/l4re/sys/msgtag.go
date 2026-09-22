// L4Re user space support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package sys

// MsgTag is the raw L4 message tag (l4_msgtag_t.raw).
//
// Layout (fiasco/src/abi/l4_types.cpp, L4_msg_tag):
//
//	bits  0..5   words
//	bits  6..11  items
//	bits 12..15  flags
//	bits 16..63  label (signed)
type MsgTag uint64

// MakeMsgTag builds a message tag from its components. The label is
// sign-extended into the upper bits, matching the kernel's signed
// arithmetic shift in L4_msg_tag::proto().
//
//go:nosplit
func MakeMsgTag(label int64, words, items uint, flags uint64) MsgTag {
	return MsgTag(uint64(label<<16) | uint64(words&0x3f) | uint64(items&0x3f)<<6 | (flags & 0xf000))
}

// Words returns the number of untyped message words.
//
//go:nosplit
func (t MsgTag) Words() uint { return uint(t & 0x3f) }

// Items returns the number of typed message items.
//
//go:nosplit
func (t MsgTag) Items() uint { return uint(t>>6) & 0x3f }

// Flags returns the flag bits.
//
//go:nosplit
func (t MsgTag) Flags() uint64 { return uint64(t) & 0xf000 }

// Label returns the signed label (protocol on request, result on reply).
//
//go:nosplit
func (t MsgTag) Label() int64 { return int64(t) >> 16 }

// HasError reports whether the IPC failed at the kernel level. The detailed
// error code is then found in TCR.error (see IpcError).
//
//go:nosplit
func (t MsgTag) HasError() bool { return uint64(t)&MsgtagError != 0 }

// Errno is an L4 error code (negative on failure), usable as Go error.
type Errno int64

// Error implements the error interface. It is nosplit so that it can be
// used in early diagnostics before the runtime is up.
//
//go:nosplit
func (e Errno) Error() string {
	switch -e {
	case EOK:
		return "ok"
	case EPerm:
		return "operation not permitted"
	case ENoEnt:
		return "no such entity"
	case EIO:
		return "I/O error"
	case ENoMem:
		return "out of memory"
	case EFault:
		return "bad address"
	case EBusy:
		return "busy"
	case EExist:
		return "already exists"
	case ENoDev:
		return "no such device"
	case EInval:
		return "invalid argument"
	case ERange:
		return "out of range"
	case ENoSys:
		return "not implemented"
	case EBadProto:
		return "unsupported protocol"
	case EAddrNotAvl:
		return "address not available"
	case ENoReply:
		return "no reply"
	case EMsgTooShort:
		return "message too short"
	case EMsgTooLong:
		return "message too long"
	case EMsgMissArg:
		return "message has invalid capability"
	}
	if -e >= EIpcLo {
		return "IPC error"
	}
	return "L4 error"
}

// Result decodes a reply tag into an Errno: an IPC level failure is mapped
// to -(EIpcLo + TCR.error), a negative label is returned as is, and success
// yields 0.
//
//go:nosplit
func (t MsgTag) Result() Errno {
	if t.HasError() {
		return Errno(-(EIpcLo + int64(TCR()[0]&IpcErrorMask)))
	}
	if l := t.Label(); l < 0 {
		return Errno(l)
	}
	return 0
}
