// L4Re user space support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package sys

import "unsafe"

// utcbPtr returns the address of the calling thread's UTCB. The kernel
// publishes it in an architecture specific register: `gs:0` on amd64,
// TPIDRRO_EL0 on arm64. Implemented in utcb_$GOARCH.s.
func utcbPtr() uintptr

// UTCB returns the address of the calling thread's UTCB.
//
//go:nosplit
func UTCB() uintptr { return utcbPtr() }

// MR returns the 64 message registers of the calling thread's UTCB.
//
//go:nosplit
func MR() *[64]uint64 {
	return (*[64]uint64)(unsafe.Pointer(utcbPtr() + UtcbMsgRegsOffset))
}

// BR returns the buffer register area of the calling thread's UTCB: BR[0]
// is the buffer descriptor register (bdr), BR[1..] the buffer items.
//
//go:nosplit
func BR() *[64]uint64 {
	return (*[64]uint64)(unsafe.Pointer(utcbPtr() + UtcbBufRegsOffset))
}

// TCR returns the thread control registers of the calling thread's UTCB:
// TCR[0] is the IPC error word, TCR[1] the transfer timeout, TCR[2..4]
// are user words.
//
//go:nosplit
func TCR() *[5]uint64 {
	return (*[5]uint64)(unsafe.Pointer(utcbPtr() + UtcbThreadRegsOffset))
}

// IpcError returns the IPC error code of the last failed IPC.
//
//go:nosplit
func IpcError() uint64 { return TCR()[0] & IpcErrorMask }

// SetRcvCap programs the buffer registers to receive a single capability
// into slot `target` (l4_factory_create_start_u pattern): bdr = 0 and one
// buffer item `target | L4_RCV_ITEM_SINGLE_CAP`, followed by a zero
// terminator.
//
//go:nosplit
func SetRcvCap(target uint64) {
	br := BR()
	br[0] = 0
	br[1] = target | RcvItemSingleCap
	br[2] = 0
}

// ClearRcvBuffers marks the buffer registers as empty.
//
//go:nosplit
func ClearRcvBuffers() {
	br := BR()
	br[0] = 0
	br[1] = 0
}

// ObjFpage encodes an object flexpage for capability `cap` with the given
// rights (l4_obj_fpage(cap, 0, rights)).
//
//go:nosplit
func ObjFpage(cap uint64, rights uint64) uint64 {
	return (cap &^ (CapStride - 1)) | FpageTypeObj<<FpageTypeShift | (rights & 0xf)
}

// PutMapItem writes a capability send item (two words) at mr[i], mr[i+1]
// and returns the next free index. The item maps `cap` with `rights` into
// the receiver's buffer (L4::Ipc::Snd_fpage(cap, rights)).
//
//go:nosplit
func PutMapItem(mr *[64]uint64, i int, cap uint64, rights uint64) int {
	mr[i] = ItemMap | (rights & 0xf0)
	mr[i+1] = ObjFpage(cap, rights)
	return i + 2
}

// PutBytes copies b into the message registers starting at byte offset
// off and returns the byte offset following the copied data.
//
//go:nosplit
func PutBytes(mr *[64]uint64, off int, b []byte) int {
	dst := (*[64 * 8]byte)(unsafe.Pointer(mr))
	for i := 0; i < len(b); i++ {
		dst[off+i] = b[i]
	}
	return off + len(b)
}

// PutU32 stores a 32-bit value at byte offset off (4-byte aligned).
//
//go:nosplit
func PutU32(mr *[64]uint64, off int, v uint32) int {
	*(*uint32)(unsafe.Pointer(uintptr(unsafe.Pointer(mr)) + uintptr(off))) = v
	return off + 4
}

// PutU8 stores an 8-bit value at byte offset off.
//
//go:nosplit
func PutU8(mr *[64]uint64, off int, v uint8) int {
	*(*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(mr)) + uintptr(off))) = v
	return off + 1
}

// Align rounds off up to a multiple of n (n must be a power of two).
//
//go:nosplit
func Align(off, n int) int { return (off + n - 1) &^ (n - 1) }
