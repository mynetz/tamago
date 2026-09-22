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

// SetTLS is provided by the user space overlay in use (setTLS symbol),
// e.g. user/linux (arch_prctl) or user/l4re (L4 thread system call). It is
// not required on bare metal, where the runtime installs FS_BASE itself.
TEXT ·SetTLS(SB),NOSPLIT|NOFRAME,$0
	JMP	setTLS(SB)
