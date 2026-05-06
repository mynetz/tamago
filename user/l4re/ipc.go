// L4Re userspace support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package l4re_user

// l4IpcSend issues an L4 IPC send to the given capability index. The mr
// buffer is copied into the UTCB message-register area (offset 0) before
// the syscall. Implemented in ipc_amd64.s.
//
//go:noescape
func l4IpcSend(cap uint64, tag uint64, timeout uint64, mr *[64]uint64)
