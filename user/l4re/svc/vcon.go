// L4Re user space support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package svc

import "github.com/usbarmory/tamago/user/l4re/sys"

// Vcon opcodes (l4sys/include/vcon.h, L4_vcon_ops).
const (
	vconWriteOp uint64 = 0
)

// VconMaxWrite is the maximum payload of a single L4::Vcon::write
// (L4_VCON_WRITE_SIZE = (L4_UTCB_GENERIC_DATA_SIZE - 2) * 8).
const VconMaxWrite = (sys.UtcbGenericDataSize - 2) * 8

// VconWrite writes b to the console capability `vcon` using one or more
// L4::Vcon::write requests (l4_vcon_send):
//
//	MR[0]   = L4_VCON_WRITE_OP
//	MR[1]   = length
//	MR[2..] = payload
//	tag     = (L4_PROTO_LOG, words = 2 + ceil(len/8), flags = SCHEDULE)
//
// The request is a one-way send; the console never replies.
//
//go:nosplit
func VconWrite(vcon Cap, b []byte) {
	for len(b) > 0 {
		n := len(b)
		if n > VconMaxWrite {
			n = VconMaxWrite
		}

		mr := sys.MR()
		mr[0] = vconWriteOp
		mr[1] = uint64(n)
		sys.PutBytes(mr, 16, b[:n])

		words := 2 + (n+7)/8
		tag := sys.MakeMsgTag(sys.ProtoLog, uint(words), 0, sys.MsgtagSchedule)
		sys.Send(uint64(vcon), tag, sys.TimeoutNever)

		b = b[n:]
	}
}

// LogWrite writes b to the console of the initial environment. It is a
// no-op before Init.
//
//go:nosplit
func LogWrite(b []byte) {
	if env == nil {
		return
	}
	VconWrite(env.Log, b)
}

// LogString writes s to the console of the initial environment.
//
//go:nosplit
func LogString(s string) {
	if env == nil {
		return
	}
	// Chunked copy through a fixed buffer avoids a string->[]byte
	// conversion, which would allocate.
	var buf [64]byte
	for len(s) > 0 {
		n := len(s)
		if n > len(buf) {
			n = len(buf)
		}
		for i := 0; i < n; i++ {
			buf[i] = s[i]
		}
		VconWrite(env.Log, buf[:n])
		s = s[n:]
	}
}

// LogHex writes v as a 0x-prefixed hexadecimal number to the console.
//
//go:nosplit
func LogHex(v uint64) {
	var buf [18]byte
	buf[0] = '0'
	buf[1] = 'x'
	for i := 0; i < 16; i++ {
		d := byte(v>>(60-4*uint(i))) & 0xf
		if d < 10 {
			buf[2+i] = '0' + d
		} else {
			buf[2+i] = 'a' + d - 10
		}
	}
	LogWrite(buf[:])
}
