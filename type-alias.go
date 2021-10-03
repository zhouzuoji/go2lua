package go2lua

import "unsafe"

type (
	i8    = int8
	u8    = uint8
	i6    = int16
	u16   = uint16
	i32   = int32
	u32   = uint32
	i64   = int64
	u64   = uint64
	uptr  = uintptr
	pvoid = unsafe.Pointer
)

const (
	_SELF_PATH = "mirgo/script/go2lua"
	_SELF_NAME = "go2lua"
	_LUA_PATH  = "github.com/Shopify/go-lua"
)
