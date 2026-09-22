// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linksysreg

#include "textflag.h"

// sysReg implements runtime/goos.SysReg for bare metal execution at EL1,
// where the ID_AA64* registers can be read directly.
//
// func sysReg(reg uint) uint64
TEXT sysReg(SB),NOSPLIT,$0-16
	MOVD	reg+0(FP), R1
	MOVD	$0x4030, R2
	CMP	R2, R1
	BEQ	isar0
	MOVD	$0x4031, R2
	CMP	R2, R1
	BEQ	isar1
	MOVD	$0x4020, R2
	CMP	R2, R1
	BEQ	pfr0
	MOVD	$0, R0
	MOVD	R0, ret+8(FP)
	RET
isar0:
	MRS	ID_AA64ISAR0_EL1, R0
	MOVD	R0, ret+8(FP)
	RET
isar1:
	MRS	ID_AA64ISAR1_EL1, R0
	MOVD	R0, ret+8(FP)
	RET
pfr0:
	MRS	ID_AA64PFR0_EL1, R0
	MOVD	R0, ret+8(FP)
	RET
