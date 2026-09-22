// L4Re user space support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package svc

import "github.com/usbarmory/tamago/user/l4re/sys"

// POSIX signal numbers and sigaction flags as defined by L4Re's musl
// (libc/musl/contrib/musl/arch/{x86_64,aarch64}/bits/signal.h), identical
// on both architectures.
const (
	SIGILL  = 4
	SIGSEGV = 11

	SigDfl uintptr = 0 // SIG_DFL
	SigIgn uintptr = 1 // SIG_IGN

	SaSiginfo uint32 = 0x4 // SA_SIGINFO
	SaNodefer uint32 = 0x40000000
)

// itasOpSigaction is the opcode of L4Re::Itas::sigaction (third entry of
// Itas::Rpcs: register_thread, unregister_thread, sigaction, ...).
const itasOpSigaction = 2

// Sigaction installs `handler` for signal `sig` with the calling task's
// in-task server (l4re_itas), which is the pager and exception handler
// of every thread it started. Exceptions of the thread that itas cannot
// resolve are converted to POSIX signals and, if a handler is installed,
// delivered on the faulting thread itself: itas rewrites the thread's
// context to enter a trampoline that calls
//
//	handler(signo int32, info *siginfo_t, uctx *ucontext_t)
//
// with the C calling convention and, on return, restores the interrupted
// context from *uctx (the handler may modify it). The UTCB is saved and
// restored around the handler, so the handler may perform IPC.
//
// The request is L4Re::Itas::sigaction(int signum, const struct
// sigaction *act, struct sigaction *oldact), with `act` passed by value
// (musl layout, 152 bytes: handler @0, sa_mask[16] @8, flags @136,
// restorer @144) after the 4-byte opcode and 4-byte signum. The reply
// returns the previous action, which is ignored here.
//
//go:nosplit
func Sigaction(sig int32, handler uintptr, flags uint32) sys.Errno {
	if env == nil {
		return sys.Errno(-sys.EFault)
	}

	sys.ClearRcvBuffers()

	mr := sys.MR()
	off := 0
	off = sys.PutU32(mr, off, itasOpSigaction)
	off = sys.PutU32(mr, off, uint32(sig))
	// struct sigaction, 8-byte aligned
	act := off / 8
	for i := 0; i < 19; i++ {
		mr[act+i] = 0
	}
	mr[act+0] = uint64(handler) // sa_handler / sa_sigaction
	mr[act+17] = uint64(flags)  // sa_flags (int, low half of the word)
	off += 19 * 8

	tag := sys.MakeMsgTag(sys.ProtoItas, uint(off/8), 0, 0)
	reply := sys.Call(uint64(env.Itas), tag, sys.TimeoutNever)

	return reply.Result()
}
