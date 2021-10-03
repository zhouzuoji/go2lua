package go2lua

import (
	"fmt"
	"math"
	"reflect"

	"github.com/Shopify/go-lua"
)

type TFieldAccessor struct {
	Getter, Setter lua.Function
}

type TClassMeta struct {
	Type       reflect.Type
	LuaMetaKey string
	Methods    map[string]lua.Function
	Fields     map[string]TFieldAccessor
}

func StringArg(vm *lua.State, idx int) string {
	val, ok := vm.ToString(idx)
	if !ok {
		vm.PushFString(fmt.Sprintf("bad type of argument %d, string expected", idx))
		vm.Error()
	}
	return val
}

func BoolArg(vm *lua.State, idx int) bool {
	val := vm.ToBoolean(idx)
	if false {
		vm.PushFString(fmt.Sprintf("bad type of argument %d, boolean expected", idx))
		vm.Error()
	}
	return val
}

func IntArg(vm *lua.State, idx int) int {
	val, ok := vm.ToInteger(idx)
	if !ok {
		vm.PushFString(fmt.Sprintf("bad type of argument %d, integer expected", idx))
		vm.Error()
	}
	return val
}

func Float64Arg(vm *lua.State, idx int) float64 {
	val, ok := vm.ToNumber(idx)
	if !ok {
		vm.PushFString(fmt.Sprintf("bad type of argument %d, number expected", idx))
		vm.Error()
	}
	return val
}

func Float32Arg(vm *lua.State, idx int) float64 {
	val, ok := vm.ToNumber(idx)
	if !ok {
		vm.PushFString(fmt.Sprintf("bad type of argument %d, number expected", idx))
		vm.Error()
	}
	if val > float64(math.MaxFloat32) {
		vm.PushFString(fmt.Sprintf("argument %d is outof range of float32", idx))
		vm.Error()
	}
	return val
}

func UserDataArg(vm *lua.State, idx int) interface{} {
	val := vm.ToUserData(idx)
	return val
}

func MakeMetaTable(vm *lua.State, cls *TClassMeta) {
	lua.NewMetaTable(vm, cls.LuaMetaKey)
	vm.PushGoFunction(cls.GetField)
	vm.SetField(-2, "__index")
	vm.PushGoFunction(cls.SetField)
	vm.SetField(-2, "__newindex")
	vm.Pop(1)
}

func (cls *TClassMeta) GetField(vm *lua.State) int {
	name, ok := vm.ToString(2)
	if !ok {
		return 0
	}
	m, ok := cls.Methods[name]
	if ok {
		vm.PushGoFunction(m)
		return 1
	}
	accessor, ok := cls.Fields[name]
	if !ok {
		return 0
	}
	return accessor.Getter(vm)
}

func (cls *TClassMeta) SetField(vm *lua.State) int {
	name, ok := vm.ToString(2)
	if !ok {
		return 0
	}
	accessor, ok := cls.Fields[name]
	if !ok {
		return 0
	}
	return accessor.Setter(vm)
}
