package unsafe_core

import (
	"errors"
	"math"
	"unsafe"
)

type LayoutSample struct {
	A byte
	B int32
	C int64
}

type LayoutInfo struct {
	Size    uintptr
	Align   uintptr
	OffsetA uintptr
	OffsetB uintptr
	OffsetC uintptr
}

// StructLayoutDemo demonstrates Sizeof/Alignof/Offsetof.
func StructLayoutDemo() LayoutInfo {
	var s LayoutSample
	return LayoutInfo{
		Size:    unsafe.Sizeof(s),
		Align:   unsafe.Alignof(s),
		OffsetA: unsafe.Offsetof(s.A),
		OffsetB: unsafe.Offsetof(s.B),
		OffsetC: unsafe.Offsetof(s.C),
	}
}

// BytesToStringNoCopy converts []byte to string without allocation.
// WARNING: if b is modified later, returned string view changes as well.
func BytesToStringNoCopy(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// StringToBytesNoCopy converts string to []byte view without allocation.
// WARNING: never mutate returned bytes; it may point to read-only memory.
func StringToBytesNoCopy(s string) []byte {
	type sliceHeader struct {
		Data uintptr
		Len  int
		Cap  int
	}
	type stringHeader struct {
		Data uintptr
		Len  int
	}
	sh := (*stringHeader)(unsafe.Pointer(&s))
	bh := sliceHeader{Data: sh.Data, Len: sh.Len, Cap: sh.Len}
	return *(*[]byte)(unsafe.Pointer(&bh))
}

// Float64BitsUnsafe reinterprets float64 bits as uint64.
func Float64BitsUnsafe(f float64) uint64 {
	return *(*uint64)(unsafe.Pointer(&f))
}

// Uint64ToFloat64Unsafe reinterprets uint64 bits as float64.
func Uint64ToFloat64Unsafe(u uint64) float64 {
	return *(*float64)(unsafe.Pointer(&u))
}

func IsLittleEndian() bool {
	var x uint16 = 0x0102
	b := (*[2]byte)(unsafe.Pointer(&x))
	return b[0] == 0x02
}

// BytesToUint32NativeUnsafe reads uint32 directly from first 4 bytes.
// Result uses machine native endianness.
func BytesToUint32NativeUnsafe(b []byte) (uint32, error) {
	if len(b) < 4 {
		return 0, errors.New("need at least 4 bytes")
	}
	return *(*uint32)(unsafe.Pointer(&b[0])), nil
}

// EqualFloatBits validates unsafe conversion against math package.
func EqualFloatBits(v float64) bool {
	return Float64BitsUnsafe(v) == math.Float64bits(v)
}
