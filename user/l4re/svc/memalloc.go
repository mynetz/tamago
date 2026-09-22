// L4Re user space support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package svc

import "github.com/usbarmory/tamago/user/l4re/sys"

// Mem_alloc flags (L4Re::Mem_alloc::Mem_alloc_flags).
const (
	MaContinuous uint64 = 0x01 // physically contiguous memory
	MaPinned     uint64 = 0x02 // pinned memory (eagerly allocated)
	MaSuperPages uint64 = 0x04 // superpage backed memory
)

// AllocDataspace requests a dataspace of `size` bytes from the memory
// allocator of the initial environment and returns its capability.
//
// The request is L4Re::Mem_alloc::alloc (l4re/include/impl/mem_alloc_impl.h),
// a factory create call for the L4Re::Dataspace protocol carrying three
// variable arguments:
//
//	MR[0]    = L4RE_PROTO_DATASPACE
//	MR[1..2] = varg mword  size
//	MR[3..4] = varg umword flags
//	MR[5..6] = varg umword align (log2)
//	tag      = (L4_PROTO_FACTORY, words=7)
//	BR       = single-cap receive buffer for the new capability
//
//go:nosplit
func AllocDataspace(size int64, flags uint64, alignLog2 uint64) (Cap, sys.Errno) {
	if env == nil {
		return 0, sys.Errno(-sys.EFault)
	}

	target := AllocCap()
	sys.SetRcvCap(uint64(target))

	const vargU = sys.VargTypeUmword | 8<<16
	const vargS = sys.VargTypeMword | 8<<16

	mr := sys.MR()
	mr[0] = uint64(sys.ProtoDataspace)
	mr[1] = vargS
	mr[2] = uint64(size)
	mr[3] = vargU
	mr[4] = flags
	mr[5] = vargU
	mr[6] = alignLog2

	tag := sys.MakeMsgTag(sys.ProtoFactory, 7, 0, 0)
	reply := sys.Call(uint64(env.MemAlloc), tag, sys.TimeoutNever)

	if e := reply.Result(); e != 0 {
		return 0, e
	}
	return target, 0
}
