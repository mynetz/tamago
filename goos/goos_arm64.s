// Custom GOOS support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

TEXT ·CPUInit(SB),NOSPLIT|NOFRAME,$0
	JMP	cpuinit(SB)

// SysReg is provided by the overlay in use (sysReg symbol): bare metal
// packages (arm64/) read the ID_AA64* registers directly at EL1, user
// space packages obtain them from their kernel or return 0.
TEXT ·SysReg(SB),NOSPLIT|NOFRAME,$0
	JMP	sysReg(SB)
