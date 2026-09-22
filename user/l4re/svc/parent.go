// L4Re user space support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package svc

import "github.com/usbarmory/tamago/user/l4re/sys"

// parentOpSignal is the opcode of L4Re::Parent::signal (only entry of
// Parent::Rpcs).
const parentOpSignal = 0

// ParentSignal sends L4Re::Parent::signal(sig, val) to the parent of the
// task. Signal 0 asks the parent (ned or moe) to terminate the task with
// exit code `val`; in that case the call does not return.
//
// Layout: int opcode at byte 0, then two unsigned longs at bytes 8 and 16
// (3 words, protocol L4RE_PROTO_PARENT).
//
//go:nosplit
func ParentSignal(sig, val uint64) sys.Errno {
	if env == nil {
		return sys.Errno(-sys.EFault)
	}

	sys.ClearRcvBuffers()

	mr := sys.MR()
	mr[0] = parentOpSignal
	mr[1] = sig
	mr[2] = val

	tag := sys.MakeMsgTag(sys.ProtoParent, 3, 0, 0)
	reply := sys.Call(uint64(env.Parent), tag, sys.TimeoutNever)

	return reply.Result()
}

// Exit terminates the task with the given exit code by signalling the
// parent and, should that return, parking the thread forever.
//
//go:nosplit
func Exit(code int32) {
	ParentSignal(0, uint64(uint32(code)))
	for {
		sys.Sleep(sys.TimeoutNever)
	}
}
