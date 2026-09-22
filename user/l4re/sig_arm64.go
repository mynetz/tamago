// L4Re user space support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package l4re_user

import "github.com/usbarmory/tamago/user/l4re/svc"

// sigillHandler returns the address of the SIGILL handler in sig_arm64.s
// that emulates the runtime's CPU feature register reads.
func sigillHandler() uintptr

// bootSignal is the signal whose handler bridges the runtime's Linux
// assumptions during startup, see sig_arm64.s.
const bootSignal = svc.SIGILL

//go:nosplit
func bootHandler() uintptr { return sigillHandler() }
