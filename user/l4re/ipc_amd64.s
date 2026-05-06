// L4Re userspace support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

// L4 IPC send trampoline for the Fiasco amd64 ABI. From
// sources/l4re/pkg/l4re-core/l4sys/include/ARCH-amd64/arch/ipc.h, the
// l4_ipc inline syscall expects:
//
//   RAX = tag.raw
//   RDX = dest | flags
//   RSI = slabel (sender label, unused for plain send)
//   R8  = timeout.raw
//
// UTCB is read by the kernel from the per-thread location addressed by
// gs:0; the message registers live at UTCB offset 0 (see
// sources/l4re/pkg/l4re-core/l4sys/include/ARCH-amd64/arch/utcb.h:
// L4_UTCB_MSG_REGS_OFFSET = 0).
//
// SYSF_SEND = 0x01 (sources/l4re/pkg/l4re-core/l4sys/include/consts.h).
//
// func l4IpcSend(cap uint64, tag uint64, timeout uint64, mr *[64]uint64)
TEXT ·l4IpcSend(SB),NOSPLIT,$0-32
	// Load arguments
	MOVQ	cap+0(FP), DX		// DX = cap
	ORQ	$0x01, DX		// DX |= L4_SYSF_SEND
	MOVQ	tag+8(FP), AX		// AX = tag.raw
	MOVQ	timeout+16(FP), R8
	MOVQ	mr+24(FP), DI		// DI = source mr buffer

	// Copy 64 mwords (512 bytes) from mr buffer into the UTCB at gs:0.
	// We read the UTCB pointer from gs:0 and use it as destination. We
	// avoid clobbering R8, RDX, RAX (used as syscall inputs) and RCX/R11
	// (clobbered by SYSCALL itself).
	BYTE	$0x65; MOVQ (0), BX	// BX = *(gs:0)  -- UTCB pointer

	// Copy MR[0..63] from (DI) to (BX).
	MOVQ	$64, CX
copy_mr:
	MOVQ	(DI), R9
	MOVQ	R9, (BX)
	ADDQ	$8, DI
	ADDQ	$8, BX
	DECQ	CX
	JNE	copy_mr

	// Sender label = 0 for plain send.
	XORQ	SI, SI

	// Issue the L4 IPC syscall. Clobbers RCX, R11, R15 per the L4 ABI.
	SYSCALL

	RET
