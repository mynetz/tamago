// L4Re userspace support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

// L4 IPC constants for set_fs_base.
//
//   L4_THREAD_AMD64_SET_SEGMENT_BASE_OP = 0x12
//   L4_AMD64_SEGMENT_FS                 = 0
//   L4_PROTO_THREAD                     = -12  (Label_thread in fiasco/abi/l4_types.cpp)
//   L4_SYSF_CALL = SEND|RECV            = 0x03
//
// msgtag.raw layout (fiasco/abi/l4_types.cpp `L4_msg_tag`, see proto()/words()):
//   raw = (label << 16) | (flags & 0xf000) | ((items & 0x3f) << 6) | (words & 0x3f)
//
// proto() returns `static_cast<long>(_tag) >> 16` (signed arithmetic shift).
// For label = -12, words = 2, items = 0, flags = 0:
//   ((-12) << 16) | 2  =  0xFFFFFFFFFFF40000 | 2  =  0xFFFFFFFFFFF40002
//
// IMPORTANT: an earlier version of this file had 0xFFFFFFFFFFFC0002 (proto=-4)
// instead of 0xFFFFFFFFFFF40002 (proto=-12). That mismatch caused
// Thread_object::invoke() to take the "do_ipc" branch (because tag.proto() !=
// Label_thread), enqueuing the calling thread as a sender to itself with
// L4_IPC_NEVER, so the task hung forever inside Sender::sender_enqueue with
// the kernel idling. See docs/settls-blocker.md.
#define THREAD_SET_SEG_OP   0x12
#define IPC_CALL_FLAGS      0x03
#define MSGTAG_SET_FSBASE   0xFFFFFFFFFFF40002

// L4_INVALID_CAP = ~0UL << (L4_CAP_SHIFT-1) = ~0UL << 11 = 0xFFFFFFFFFFFFF800.
//
// When IPC is invoked on L4_INVALID_CAP, the kernel resolves it to a
// capability for the *current* thread with full permissions. This is the
// idiom upstream uClibc / musl / libpthread use for `set_fs` — see
// sources/l4re/pkg/l4re-core/libc/musl/libc/ARCH-x86_64/impl-libc-api-arch.c
// and l4sys/include/types.h's documentation of l4_cap_idx_t. It avoids any
// dependency on l4re_env_t.main_thread, whose value is task-relative and
// not the canonical handle for self-thread invocations.
#define L4_INVALID_CAP_PARTIAL 0xFFFFFFFFFFFFF800

// setTLSUser is the L4Re-native body of runtime/goos.SetTLSUser. It is
// invoked from runtime.settls's userspace path (via the
// runtime/goos.SetTLSUser → setTLSUser jump in goos/goos_amd64.s) to
// install the amd64 FS_BASE register for the current thread.
//
// The Go runtime is not yet up at this point: FS_BASE is 0, no Go-level
// state (g, m) is reachable via TLS. Routine must be NOSPLIT and must
// not touch the stack beyond what SYSCALL allows.
//
// Receives DI = base (already biased by +8 per the Go ELF -8(FS)
// convention).
//
// Issues an L4 IPC THREAD_SET_SEGMENT_BASE_OP against L4_INVALID_CAP,
// which Fiasco resolves to "the current thread".
TEXT setTLSUser(SB),NOSPLIT|NOFRAME,$0
	// Stage MR[0]=op, MR[1]=base in the UTCB at gs:0.
	BYTE	$0x65; MOVQ (0), BX	// BX = *(gs:0) = UTCB pointer
	MOVQ	$THREAD_SET_SEG_OP, (BX)
	MOVQ	DI, 8(BX)

	// Issue the L4 IPC: RAX=tag, RDX=L4_INVALID_CAP|IPC_CALL,
	// RSI=0, R8=0 (L4_IPC_NEVER timeout).
	MOVQ	$L4_INVALID_CAP_PARTIAL, DX
	ORQ	$IPC_CALL_FLAGS, DX
	MOVQ	$MSGTAG_SET_FSBASE, AX
	XORQ	SI, SI
	XORQ	R8, R8
	SYSCALL

	// FS_BASE is now installed. Return to runtime.settls's caller.
	RET
