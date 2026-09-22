// L4Re user space support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

// L4 IPC system call, Fiasco arm64 register convention. l4_ipc() in
// l4sys/include/ARCH-arm64/arch/ipc.h calls
//
//	__l4_sys_syscall(tag.raw, slabel, dest | flags, timeout.raw)
//
// which is `svc #0; ret` (l4sys/lib/src/ARCH-arm64/syscall.S), i.e.
//
//	X0 = tag.raw      -> reply tag.raw
//	X1 = sender label -> received label
//	X2 = dest | flags
//	X3 = timeout.raw
//
// The kernel honours the AAPCS64 calling convention: X0-X18 may be
// clobbered, X19-X30 and SP are preserved. The UTCB address is published in
// TPIDRRO_EL0.
//
// func ipc(dest uint64, tag MsgTag, slabel uint64, timeout uint64) MsgTag
TEXT ·ipc(SB),NOSPLIT,$0-40
	MOVD	tag+8(FP), R0
	MOVD	slabel+16(FP), R1
	MOVD	dest+0(FP), R2
	MOVD	timeout+24(FP), R3
	SVC
	MOVD	R0, ret+32(FP)
	RET

// utcbPtr returns TPIDRRO_EL0, the UTCB address of the calling thread
// (l4_utcb_direct() in l4sys/include/ARCH-arm64/arch/utcb.h).
//
// func utcbPtr() uintptr
TEXT ·utcbPtr(SB),NOSPLIT,$0-8
	MRS	TPIDRRO_EL0, R0
	MOVD	R0, ret+0(FP)
	RET

// kipReadNs calls the clock stub the kernel places into the KIP at
// L4_KIP_OFFS_READ_NS (fiasco/src/kern/arm/64/kip-time.S). The stub is a
// leaf function following the C ABI that clobbers only X0-X5.
//
// func kipReadNs(kip uintptr) int64
TEXT ·kipReadNs(SB),NOSPLIT,$16-16
	MOVD	kip+0(FP), R1
	ADD	$0x980, R1, R1
	CALL	(R1)
	MOVD	R0, ret+8(FP)
	RET
