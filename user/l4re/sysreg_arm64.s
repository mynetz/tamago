// L4Re user space support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

// sysReg implements runtime/goos.SysReg. L4Re tasks run at EL0 where the
// ID_AA64* registers trap, and Fiasco neither emulates the access nor
// exposes the values through the KIP. Report 0, i.e. no optional
// instruction set features; the runtime and standard library then use
// their generic code paths.
//
// func sysReg(reg uint) uint64
TEXT sysReg(SB),NOSPLIT,$0-16
	MOVD	$0, R0
	MOVD	R0, ret+8(FP)
	RET
