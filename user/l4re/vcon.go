// L4Re userspace support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package l4re_user

import "unsafe"

// Vcon protocol constants. See:
//
//	sources/l4re/pkg/l4re-core/l4sys/include/vcon.h
//	    enum L4_vcon_ops { L4_VCON_WRITE_OP = 0UL, ... };
//
//	sources/l4re/pkg/l4re-core/l4sys/include/types.h
//	    L4_PROTO_LOG       = -13L
//	    L4_MSGTAG_SCHEDULE = 0x2000
//
//	sources/l4re/pkg/l4re-core/l4sys/include/consts.h
//	    L4_SYSF_SEND      0x01
const (
	vconWriteOp     uint64 = 0
	protoLog        int64  = -13
	msgtagSchedule  uint64 = 0x2000
	sysfSend        uint64 = 0x01
	ipcTimeoutNever uint64 = 0
)

// Maximum chunk size per IPC. The amd64 UTCB has 63 mwords of generic
// message data; the Vcon protocol uses MR[0] for the opcode and MR[1] for
// the payload size, leaving room for 61 * 8 = 488 byte chunks. We pick a
// conservative cap.
const vconMaxChunk = 256

// vconWrite issues one or more L4_VCON_WRITE_OP IPCs to the given vcon
// capability index, splitting the buffer into chunks that fit a single IPC.
func vconWrite(cap uint64, buf []byte) {
	for len(buf) > 0 {
		n := len(buf)
		if n > vconMaxChunk {
			n = vconMaxChunk
		}
		vconWriteChunk(cap, buf[:n])
		buf = buf[n:]
	}
}

// vconWriteChunk writes one chunk that fits in a single IPC.
func vconWriteChunk(cap uint64, buf []byte) {
	// Build the message tag (see types.h `l4_msgtag` constructor):
	//
	//   raw = (label << 16) | (words & 0x3f) | ((items & 0x3f) << 6) | (flags & 0xf000)
	//
	// label = L4_PROTO_LOG (-13), so the upper bits are sign-extended ones.
	// We compute the signed shift first then bit-cast to uint64.
	// words = 2 + ceil(len / 8). items = 0. flags = SCHEDULE.
	words := uint64(2 + (len(buf)+7)/8)
	// Two's-complement reinterpret of the signed protocol label shifted
	// into the high bits of the message tag. Routing this through a
	// stack variable forces a runtime cast and avoids the
	// constant-overflow check Go does on uint64(<negative-constant>).
	signedLabel := protoLog
	labelShifted := uint64(signedLabel << 16)
	tag := labelShifted | (words & 0x3f) | (msgtagSchedule & 0xf000)

	// Stage MR[0]=op, MR[1]=size, MR[2..]=buf into a local buffer. Asm
	// copies it into the UTCB and issues the syscall.
	var mr [64]uint64
	mr[0] = vconWriteOp
	mr[1] = uint64(len(buf))

	// Copy payload bytes into mr[2..] starting at byte offset 16.
	mrBytes := (*[512]byte)(unsafe.Pointer(&mr[0]))
	for i, b := range buf {
		mrBytes[16+i] = b
	}

	l4IpcSend(cap, tag, ipcTimeoutNever, &mr)
}
