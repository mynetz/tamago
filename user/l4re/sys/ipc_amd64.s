// L4Re user space support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

// L4 IPC system call, Fiasco amd64 register convention
// (l4sys/include/ARCH-amd64/arch/ipc.h, l4_ipc()):
//
//	RAX = tag.raw          -> reply tag.raw
//	RDX = dest | flags     -> clobbered
//	RSI = sender label     -> received label
//	R8  = timeout.raw
//
// The kernel reads and writes the UTCB (message/buffer registers) of the
// calling thread directly; the UTCB address is published in gs:0. SYSCALL
// clobbers RCX and R11, Fiasco additionally clobbers R15 (documented as
// scratch in the L4 ABI), all of which are caller-saved for Go ABI0.
//
// func ipc(dest uint64, tag MsgTag, slabel uint64, timeout uint64) MsgTag
TEXT ·ipc(SB),NOSPLIT,$0-40
	MOVQ	dest+0(FP), DX
	MOVQ	tag+8(FP), AX
	MOVQ	slabel+16(FP), SI
	MOVQ	timeout+24(FP), R8
	SYSCALL
	MOVQ	AX, ret+32(FP)
	RET

// utcbPtr returns *(gs:0), the UTCB address of the calling thread
// (l4_utcb_direct() in l4sys/include/ARCH-amd64/arch/utcb.h).
//
// func utcbPtr() uintptr
TEXT ·utcbPtr(SB),NOSPLIT,$0-8
	BYTE	$0x65; MOVQ (0), AX	// gs-prefixed MOVQ 0, AX
	MOVQ	AX, ret+0(FP)
	RET

// kipReadNs calls the clock stub the kernel places into the KIP at
// L4_KIP_OFFS_READ_NS (fiasco/src/kern/ia32/64/kip-time.S). The stub is a
// leaf function following the C ABI that clobbers only RAX and RDX.
//
// func kipReadNs(kip uintptr) int64
TEXT ·kipReadNs(SB),NOSPLIT,$8-16
	MOVQ	kip+0(FP), AX
	ADDQ	$0x980, AX
	CALL	AX
	MOVQ	AX, ret+8(FP)
	RET
