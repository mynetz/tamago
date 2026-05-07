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
//   L4_PROTO_THREAD                     = -12
//   L4_SYSF_CALL = SEND|RECV            = 0x03
//
// msgtag.raw layout (types.h `l4_msgtag`):
//   raw = (label<<16) | (words & 0x3f) | ((items & 0x3f)<<6) | (flags & 0xf000)
// For label=-12, words=2, items=0, flags=0:
//   raw = uint64(int64(-12)<<16) | 2 = 0xFFFFFFFFFFFC0002
#define THREAD_SET_SEG_OP   0x12
#define IPC_CALL_FLAGS      0x03
#define MSGTAG_SET_FSBASE   0xFFFFFFFFFFFC0002

// l4re_env_t.main_thread offset (l4_cap_idx_t = 8 bytes per field;
// parent=0x00, rm=0x08, mem_alloc=0x10, log=0x18, main_thread=0x20).
// See sources/l4re/pkg/l4re-core/l4re/include/env.h.
#define ENV_OFF_MAIN_THREAD 0x20

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
// Issues an L4 IPC against l4re_env_t.main_thread with the
// L4_THREAD_AMD64_SET_SEGMENT_BASE_OP opcode for segment FS=0.
TEXT setTLSUser(SB),NOSPLIT|NOFRAME,$0
	// Stage MR[0]=op, MR[1]=base in the UTCB at gs:0.
	BYTE	$0x65; MOVQ (0), BX	// BX = *(gs:0) = UTCB pointer
	MOVQ	$THREAD_SET_SEG_OP, (BX)
	MOVQ	DI, 8(BX)

	// Read main_thread cap from l4re_env_t (saved by cpuinit).
	MOVQ	·l4reEnv(SB), CX
	MOVQ	ENV_OFF_MAIN_THREAD(CX), DX

	// Issue the L4 IPC: RAX=tag, RDX=cap|flags, RSI=0, R8=0 (never).
	ORQ	$IPC_CALL_FLAGS, DX
	MOVQ	$MSGTAG_SET_FSBASE, AX
	XORQ	SI, SI
	XORQ	R8, R8
	SYSCALL

	// FS_BASE is now installed. Return to runtime.settls's caller.
	RET
