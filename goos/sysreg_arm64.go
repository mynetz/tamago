// Custom GOOS support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package goos

// System register identifiers accepted by SysReg, encoded as in the MRS
// instruction (op0:op1:CRn:CRm:op2).
const (
	ID_AA64PFR0_EL1  uint = 0x4020
	ID_AA64ISAR0_EL1 uint = 0x4030
	ID_AA64ISAR1_EL1 uint = 0x4031
)
