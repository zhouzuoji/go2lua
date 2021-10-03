package go2lua

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"reflect"
	"runtime"
	"strings"
)

type tFile struct {
	pkg     *tPackage
	sb      bytes.Buffer
	imports map[string]tPkgName
}

func (file *tFile) addImport(pkgPath string) string {
	if pkgPath == "" {
		return ""
	}
	pn := file.pkg.addImport(pkgPath)
	file.imports[pkgPath] = pn
	return pn.name
}

func (file *tFile) writeClass(cls tClass) {
	file.addImport(_SELF_PATH)
	file.addImport(_LUA_PATH)
	file.addImport("reflect")
	t := cls.t
	ptrT := reflect.PtrTo(t)
	pkgName := file.addImport(t.PkgPath())
	file.fmtWriteln("var %s_%s = &%s.TClassMeta {", pkgName, t.Name(), _SELF_NAME)
	file.indent(1)
	file.fmtWriteln("LuaMetaKey: \"%s\",", cls.metaKey)
	file.indent(1)
	file.fmtWriteln("Type: reflect.TypeOf(%s.%s{}),", pkgName, t.Name())
	file.indent(1)
	file.write("Methods: map[string]go_lua.Function {\n")
	for i := 0; i < ptrT.NumMethod(); i++ {
		m := ptrT.Method(i)
		file.indent(2)
		file.fmtWrite(`"%s": `, m.Name)
		file.writeMethod(m, 2)
		file.write(",\n")
	}
	file.indent(1)
	file.write("},\n")
	file.indent(1)
	file.fmtWrite("Fields: map[string]%s.TFieldAccessor{\n", _SELF_NAME)
	file.indent(1)
	file.write("},\n}")
}

func (file *tFile) writeMethod(m reflect.Method, indentDepth int) {
	t := m.Type
	numIn := t.NumIn()
	file.indent(indentDepth)
	file.writeln("func (vm *go_lua.State) int {")
	file.parseSelf(t.In(0), indentDepth+1)
	for i := 1; i < numIn; i++ {
		file.parseArg(t.In(i), i+1, indentDepth+1)
	}

	file.indent(indentDepth + 1)
	numOut := t.NumOut()
	if numOut > 0 {
		file.write("ret1")
		for i := 1; i < numOut; i++ {
			file.fmtWrite(", ret%d", i+1)
		}
		file.write(" := ")
	}

	file.fmtWrite("arg1.%s(", m.Name)
	if numIn > 1 {
		file.write("arg2")
		for i := 3; i <= numIn; i++ {
			file.fmtWrite(", arg%d", i)
		}
	}
	file.write(")\n")

	for i := 0; i < numOut; i++ {
		file.pushReturnValue(t.Out(i), i+1, indentDepth+1)
	}

	file.indent(indentDepth + 1)
	file.fmtWriteln("return %d", numOut)
	file.indent(indentDepth)
	file.write("}")
}

func (file *tFile) writeFunctoin(f interface{}) (luaFuncName, goFuncName string) {
	file.addImport(_LUA_PATH)
	v := reflect.ValueOf(f)
	t := v.Type()
	if t.Kind() != reflect.Func {
		log.Panicf("not function: %#v", f)
	}

	pkgPath, typeName, luaFuncName := funcNameEx(v)
	if typeName != "" {
		panic(pkgPath + "." + typeName + "." + luaFuncName + " is not pure function")
	}
	pkgName := file.addImport(pkgPath)
	goFuncName = pkgName + "_" + luaFuncName
	file.fmtWriteln("func %s(vm *go_lua.State) int {", goFuncName)
	numIn := t.NumIn()
	for i := 0; i < numIn; i++ {
		file.parseArg(t.In(i), i+1, 1)
	}

	file.indent(1)
	numOut := t.NumOut()
	if numOut > 0 {
		file.write("ret1")
		for i := 1; i < numOut; i++ {
			file.fmtWrite(", ret%d", i+1)
		}
		file.write(" := ")
	}
	file.fmtWrite("%s.%s(", pkgName, luaFuncName)
	if numIn > 0 {
		file.write("arg1")
		for i := 2; i <= numIn; i++ {
			file.fmtWrite(", arg%d", i)
		}
	}
	file.write(")\n")

	for i := 0; i < numOut; i++ {
		file.pushReturnValue(t.Out(i), i+1, 1)
	}
	file.indent(1)
	file.fmtWriteln("return %d", numOut)
	file.write("}\n\n")
	return
}

const _INDENT = "  "

func (file *tFile) finish(w io.Writer) {
	var sb bytes.Buffer
	sb.WriteString("package ")
	sb.WriteString(file.pkg.name)
	sb.WriteString("\n\n")

	if len(file.imports) > 0 {
		sb.WriteString("import (\n")
		for pkgPath, pkgName := range file.imports {
			sb.WriteString(_INDENT)
			if pkgName.isAlias {
				sb.WriteString(pkgName.name)
				sb.WriteByte(' ')
			}
			sb.WriteByte('"')
			sb.WriteString(pkgPath)
			sb.WriteString("\"\n")
		}
		sb.WriteString(")\n\n")
	}
	w.Write(sb.Bytes())
	w.Write(file.sb.Bytes())
}

func (file *tFile) write(s string) {
	file.sb.WriteString(s)
}

func (file *tFile) writeln(s string) {
	file.sb.WriteString(s)
	file.sb.WriteByte('\n')
}

func (file *tFile) fmtWrite(format string, a ...interface{}) {
	file.sb.WriteString(fmt.Sprintf(format, a...))
}

func (file *tFile) fmtWriteln(format string, a ...interface{}) {
	file.sb.WriteString(fmt.Sprintf(format, a...))
	file.sb.WriteByte('\n')
}

func (file *tFile) indent(depth int) {
	for i := 0; i < depth; i++ {
		file.write(_INDENT)
	}
}

func funcNameEx(v reflect.Value) (pkgPath, typeName, name string) {
	fullName := runtime.FuncForPC(v.Pointer()).Name()
	separator := strings.LastIndexByte(fullName, '/')
	if separator >= 0 {
		separator += strings.IndexByte(fullName[separator:], '.')
	} else {
		separator = strings.IndexByte(fullName, '.')
	}
	pkgPath = fullName[:separator]
	name = fullName[separator+1:]
	separator = strings.IndexByte(name, '.')
	if separator >= 0 {
		typeName = name[:separator]
	}
	name = name[separator+1:]
	return
}
