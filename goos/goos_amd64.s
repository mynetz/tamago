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

// SetTLSUser is invoked from runtime.settls's userspace path. The
// overlay-supplied symbol setTLSUser provides the kernel-specific FS_BASE
// install sequence (e.g. an L4 IPC against the thread capability for
// L4Re-native tasks). DI holds the desired FS base value (already biased
// by +8). The implementation must RET to runtime.settls's caller.
TEXT ·SetTLSUser(SB),NOSPLIT|NOFRAME,$0
	JMP	setTLSUser(SB)
