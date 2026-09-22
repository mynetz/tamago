// L4Re user space support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

// On amd64 the Go runtime keeps the current goroutine pointer in thread
// local storage (-8(FS)) and, at startup, asks the environment to install
// FS_BASE for the initial thread (runtime.settls -> runtime/goos.SetTLS).
// Under Linux this is arch_prctl(ARCH_SET_FS); under Fiasco the same
// operation is a system call on the thread object:
//
//	L4_THREAD_AMD64_SET_SEGMENT_BASE_OP = 0x12 (l4sys/include/ARCH-amd64/segment.h)
//	L4_AMD64_SEGMENT_FS                 = 0    (bits 4.. of the opcode word)
//	MR[1]                               = base
//	tag                                 = (L4_PROTO_THREAD = -12, words = 2)
//
// invoked with l4_ipc_call on L4_INVALID_CAP, which the kernel resolves to
// the calling thread (the idiom used by musl's TLS setup in L4Re, see
// libc/musl/libc/ARCH-x86_64/impl-libc-api-arch.c). Fiasco dispatches the
// thread protocol synchronously (Thread_object::invoke -> invoke_arch), so
// the call returns immediately.
//
// The tag value is written out in full because the label occupies the
// upper 48 bits as a sign-extended value: (-12 << 16) | 2.
#define THREAD_SET_SEG_BASE_OP	0x12
#define MSGTAG_THREAD_WORDS2	0xFFFFFFFFFFF40002
#define L4_INVALID_CAP		0xFFFFFFFFFFFFF800
#define L4_SYSF_CALL		0x03

// setTLS installs FS_BASE = DI for the current thread. It runs before the
// Go runtime is up (no g, no TLS): NOSPLIT, no Go calls, only SYSCALL
// scratch registers are clobbered.
//
// DI = base, already biased for the -8(FS) convention by runtime.settls.
TEXT setTLS(SB),NOSPLIT|NOFRAME,$0
	BYTE	$0x65; MOVQ (0), BX	// BX = *(gs:0) = UTCB
	MOVQ	$THREAD_SET_SEG_BASE_OP, (BX)
	MOVQ	DI, 8(BX)

	MOVQ	$(L4_INVALID_CAP | L4_SYSF_CALL), DX
	MOVQ	$MSGTAG_THREAD_WORDS2, AX
	XORQ	SI, SI			// sender label
	XORQ	R8, R8			// timeout: never
	SYSCALL
	RET
