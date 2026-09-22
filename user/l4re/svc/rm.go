// L4Re user space support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package svc

import "github.com/usbarmory/tamago/user/l4re/sys"

// Region map flags (L4Re::Rm::F, l4re/include/rm).
const (
	RmR   uint32 = 0x4 // readable
	RmW   uint32 = 0x1 // writable
	RmX   uint32 = 0x2 // executable
	RmRW  uint32 = RmR | RmW
	RmRWX uint32 = RmRW | RmX

	RmSearchAddr uint32 = 0x20000  // let the region map pick the address
	RmInArea     uint32 = 0x40000  // search only inside the given area
	RmEagerMap   uint32 = 0x80000  // eagerly map all pages
	RmNoEagerMap uint32 = 0x100000 // never eagerly map
)

// rmOpAttach is the opcode of L4Re::Rm::attach (first entry of Rm::Rpcs).
const rmOpAttach = 0

// Attach maps dataspace `ds` into the task's address space through the
// region map of the initial environment and returns the start address of
// the region. With RmSearchAddr in `flags` the region map chooses an
// address at or above `start`; otherwise `start` is the requested address.
//
// The request is L4Re::Rm::attach (l4re/include/rm), marshalled per the
// L4::Ipc RPC rules (opcode is a 4-byte int, arguments follow with their
// natural alignment, capabilities become map items after the data words):
//
//	byte  0: int      opcode = 0 (attach)
//	byte  8: l4_addr_t start
//	byte 16: ulong     size
//	byte 24: uint32    flags
//	byte 32: uint64    offset in dataspace
//	byte 40: uint8     alignment (log2)
//	byte 48: cap idx   client_cap (Opt, the local dataspace selector)
//	byte 56: ulong     name length (1)
//	byte 64: char      name ("")
//	byte 72: uint64    backing_offset
//	MR[10..11]: map item for `ds` with RW rights
//	tag = (L4RE_PROTO_RM, words=10, items=1)
//
// The reply carries the (possibly searched) start address in MR[0].
//
//go:nosplit
func Attach(start uintptr, size uint64, flags uint32, ds Cap) (uintptr, sys.Errno) {
	if env == nil {
		return 0, sys.Errno(-sys.EFault)
	}

	sys.ClearRcvBuffers()

	mr := sys.MR()
	off := 0
	off = sys.PutU32(mr, off, rmOpAttach)
	off = sys.Align(off, 8)
	mr[off/8] = uint64(start)
	off += 8
	mr[off/8] = size
	off += 8
	off = sys.PutU32(mr, off, flags)
	off = sys.Align(off, 8)
	mr[off/8] = 0 // dataspace offset
	off += 8
	off = sys.PutU8(mr, off, 0) // alignment
	off = sys.Align(off, 8)
	mr[off/8] = uint64(ds) // client_cap
	off += 8
	mr[off/8] = 1 // name length (terminator only)
	off += 8
	off = sys.PutU8(mr, off, 0) // name
	off = sys.Align(off, 8)
	mr[off/8] = 0 // backing_offset
	off += 8

	words := off / 8
	rights := sys.CapFpageR
	if flags&RmW != 0 {
		rights = sys.CapFpageRW
	}
	sys.PutMapItem(mr, words, uint64(ds), rights)

	tag := sys.MakeMsgTag(sys.ProtoRm, uint(words), 1, 0)
	reply := sys.Call(uint64(env.Rm), tag, sys.TimeoutNever)

	if e := reply.Result(); e != 0 {
		return 0, e
	}
	return uintptr(mr[0]), 0
}
