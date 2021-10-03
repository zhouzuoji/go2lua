package go2lua

import (
	"reflect"
)

func (file *tFile) writeTypeName(t reflect.Type) {
	for t.Kind() == reflect.Ptr {
		file.write("*")
		t = t.Elem()
	}
	pkgName := file.addImport(t.PkgPath())
	if pkgName != "" {
		file.write(pkgName)
		file.write(".")
	}
	name := t.Name()
	if name == "" {
		name = t.String()
	}
	file.write(name)
}

func (file *tFile) parseArg(t reflect.Type, idx, indentDepth int) {
	k := t.Kind()
	if k <= reflect.Float64 || k == reflect.String {
		file.indent(indentDepth)
		file.addImport(_SELF_PATH)
		file.fmtWrite("arg%d := ", idx)
		file.writeTypeName(t)
		file.fmtWriteln("(%s.%s(vm, %d))", _SELF_NAME, argParserNames[k], idx)
		return
	}
	switch k {
	case reflect.Interface:
		file.indent(indentDepth)
		file.fmtWrite("arg%d := vm.ToValue(%d).(", idx, idx)
		file.writeTypeName(t)
		file.write(")\n")
	case reflect.Struct:
		file.parseUserData(t, idx, indentDepth)
	case reflect.Ptr:
		file.parseUserData(t, idx, indentDepth)
	case reflect.Array:
	case reflect.Func:

	case reflect.Map:
	case reflect.Slice:
	default:
		panic("unsupported type in lua")
	}
}

func (file *tFile) parseUserData(t reflect.Type, idx, indentDepth int) {
	file.addImport(_SELF_PATH)
	et := t
	if t.Kind() == reflect.Ptr {
		et = t.Elem()
	}
	if et.Kind() != reflect.Struct {
		panic("only struct type can be wrapped into lua userdata")
	}
	file.indent(indentDepth)
	file.fmtWrite("arg%d := %s.UserDataArg(vm, %d).(", idx, _SELF_NAME, idx)
	file.writeTypeName(t)
	file.write(")\n")
}

func (file *tFile) parseSelf(t reflect.Type, indentDepth int) {
	file.addImport(_SELF_PATH)
	et := t
	if t.Kind() == reflect.Ptr {
		et = t.Elem()
	}
	if et.Kind() != reflect.Struct {
		panic("only struct type can be wrapped into lua userdata")
	}
	file.indent(indentDepth)
	file.write("var arg1 *")
	file.writeTypeName(et)
	file.write("\n")
	file.indent(indentDepth)
	file.fmtWrite("arg1, ok := %s.UserDataArg(vm, 1).(*", _SELF_NAME)
	file.writeTypeName(et)
	file.write(")\n")
	file.indent(indentDepth)
	file.write("if !ok {\n")
	file.indent(indentDepth + 1)
	file.fmtWrite("tmp := %s.UserDataArg(vm, 1).(", _SELF_NAME)
	file.writeTypeName(et)
	file.write(")\n")
	file.indent(indentDepth + 1)
	file.write("arg1 = &tmp\n")
	file.indent(indentDepth)
	file.write("}\n")
}

func (file *tFile) pushReturnValue(t reflect.Type, idx, indentDepth int) {
	k := t.Kind()
	file.indent(indentDepth)
	if k >= reflect.Int && k <= reflect.Uintptr {
		file.fmtWriteln("vm.PushInteger(int(ret%d))", idx)
		return
	}
	switch k {
	case reflect.Bool:
		file.fmtWriteln("vm.PushBoolean(bool(ret%d))", idx)
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		file.fmtWriteln("vm.PushNumber(float64(ret%d))", idx)
	case reflect.String:
		file.fmtWriteln("vm.PushString(string(ret%d))", idx)
	case reflect.Interface:
		fallthrough
	case reflect.Struct:
		fallthrough
	case reflect.Ptr:
		file.addImport(_SELF_PATH)
		file.fmtWriteln("PushUserData(vm, ret%d)", idx)
	case reflect.Array:
	case reflect.Func:
	case reflect.Map:
	case reflect.Slice:
	default:
		panic("unsupported type in lua")
	}
}

var argParserNames = [...]string{
	"",
	"BoolArg",
	"IntArg",
	"IntArg",
	"IntArg",
	"IntArg",
	"IntArg",
	"IntArg",
	"IntArg",
	"IntArg",
	"IntArg",
	"IntArg",
	"IntArg",
	"Float32Arg",
	"Float64Arg",
	"",
	"",
	"",
	"",
	"",
	"",
	"",
	"UserDataArg",
	"",
	"StringArg",
	"",
	"",
}
