// L4Re user space support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package l4re_user

import "github.com/usbarmory/tamago/user/l4re/svc"

// sigsegvHandler returns the address of the SIGSEGV handler in sig_amd64.s
// that completes the runtime's TLS bring-up.
func sigsegvHandler() uintptr

// bootSignal is the signal whose handler bridges the runtime's Linux
// assumptions during startup, see sig_amd64.s.
const bootSignal = svc.SIGSEGV

//go:nosplit
func bootHandler() uintptr { return sigsegvHandler() }
