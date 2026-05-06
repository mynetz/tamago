// L4Re userspace support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package l4re_user

import "unsafe"

// unsafePtr converts a numeric address into an unsafe.Pointer. Used to read
// fields from L4Re-supplied data structures (e.g. l4re_env_t).
func unsafePtr(addr uint64) unsafe.Pointer {
	return unsafe.Pointer(uintptr(addr))
}
