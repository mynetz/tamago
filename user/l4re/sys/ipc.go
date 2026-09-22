// L4Re user space support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package sys

// ipc issues the L4 IPC system call with `dest` (capability selector OR-ed
// with SysfSend/SysfRecv/... flags), message tag `tag`, sender label
// `slabel` and `timeout`. Message data is exchanged through the calling
// thread's UTCB, which the kernel locates itself. It returns the reply
// tag. Implemented in ipc_$GOARCH.s.
//
//go:noescape
func ipc(dest uint64, tag MsgTag, slabel uint64, timeout uint64) MsgTag

// Send performs a one-way send of the message currently staged in the
// UTCB message registers to `cap`.
//
//go:nosplit
func Send(cap uint64, tag MsgTag, timeout uint64) MsgTag {
	return ipc(cap|SysfSend, tag, 0, timeout)
}

// Call performs a synchronous call (send followed by a closed receive)
// to `cap` with the message staged in the UTCB. The reply overwrites the
// message registers; the reply tag is returned.
//
//go:nosplit
func Call(cap uint64, tag MsgTag, timeout uint64) MsgTag {
	return ipc(cap|SysfCall, tag, 0, timeout)
}

// Sleep blocks the calling thread in an open receive on the invalid
// capability, i.e. until the timeout expires. Timeout zero (TimeoutNever)
// sleeps forever.
//
//go:nosplit
func Sleep(timeout uint64) MsgTag {
	return ipc(InvalidCap|SysfRecv, 0, 0, timeout)
}

// kipReadNs calls the kernel-provided clock stub embedded in the KIP at
// KipOffsReadNs and returns the current time in nanoseconds. Implemented in
// kip_$GOARCH.s.
func kipReadNs(kip uintptr) int64

// KipClockNs returns the kernel clock in nanoseconds using the KIP at
// `kip`. It is safe to call without a valid goroutine.
//
//go:nosplit
func KipClockNs(kip uintptr) int64 { return kipReadNs(kip) }

// Timeout encodes an L4 timeout (l4_timeout_t.raw) with the given receive
// timeout in microseconds and no send timeout. The kernel represents
// timeouts as 10-bit mantissa / 5-bit exponent pairs (value = m << e µs);
// the conversion rounds up so that the sleep never returns early.
//
// See l4_timeout_from_us() in l4sys/include/__timeout.h.
//
//go:nosplit
func Timeout(rcvUs uint64) uint64 {
	const mMax, eMax = 0x3ff, 0x1f
	if rcvUs == 0 {
		return 0x0400 // L4_IPC_TIMEOUT_0
	}
	if rcvUs > (1<<41)-1 {
		return 0 // L4_IPC_TIMEOUT_NEVER
	}
	var e uint64
	for v := rcvUs; v >= 1<<10; v >>= 1 {
		e++
	}
	m := (rcvUs + (1<<e - 1)) >> e
	if m > mMax {
		if e < eMax {
			e++
			m >>= 1
		} else {
			m = mMax
		}
	}
	return (e << 10) | m
}
