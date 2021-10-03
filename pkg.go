package go2lua

import (
	"log"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
)

type tPkgName struct {
	name    string
	isAlias bool
}

type tClass struct {
	t       reflect.Type
	metaKey string // key of class meta-table
	pkgName string
}

type tModule struct {
	funcNameMap map[string]string // lua function name => wrapper go function name
	funcs       []interface{}
}

type tPackage struct {
	path     string
	name     string
	imports  map[string]tPkgName
	uniNames map[string]string
	classes  map[reflect.Type]tClass
	modules  map[string]*tModule
}

func NewPackage(path string) *tPackage {
	return &tPackage{
		path:     path,
		name:     pkgShortName(path),
		imports:  map[string]tPkgName{},
		uniNames: map[string]string{},
		classes:  map[reflect.Type]tClass{},
		modules:  map[string]*tModule{},
	}
}

func (pkg *tPackage) NewFile() *tFile {
	return &tFile{
		pkg:     pkg,
		imports: map[string]tPkgName{},
	}
}

func (pkg *tPackage) AddFunctions(moduleName string, funcs []interface{}) {
	m := pkg.modules[moduleName]
	if m == nil {
		m = &tModule{funcNameMap: map[string]string{}}
		pkg.modules[moduleName] = m
	}
	for _, f := range funcs {
		t := reflect.TypeOf(f)
		if t.Kind() != reflect.Func {
			log.Panicf("not function: %#v", f)
		}
		m.funcs = append(m.funcs, f)
	}
}

func (pkg *tPackage) registerFuncs() {
	for mName, m := range pkg.modules {
		file := pkg.NewFile()
		for _, f := range m.funcs {
			luaName, goName := file.writeFunctoin(f)
			m.funcNameMap[luaName] = goName
		}
		fp := pkg.createFile(mName)
		defer fp.Close()
		file.finish(fp)
	}
}

func (pkg *tPackage) Finish() {
	pkg.registerFuncs()
	for _, cls := range pkg.classes {
		pkg.registerClass(cls)
	}
	pkg.makeRegistry()
}

func (pkg *tPackage) makeRegistry() {
	file := pkg.NewFile()
	file.addImport(_LUA_PATH)
	file.addImport(_SELF_PATH)
	file.addImport("reflect")
	file.fmtWrite("var classes map[reflect.Type]*%s.TClassMeta\n", _SELF_NAME)
	file.fmtWrite("var _classes = map[reflect.Type]*%s.TClassMeta {\n", _SELF_NAME)
	for _, cls := range pkg.classes {
		file.indent(1)
		file.fmtWriteln("%s_%s.Type: %s_%s,", cls.pkgName, cls.t.Name(), cls.pkgName, cls.t.Name())
	}
	file.write("}\n\nvar modules = map[string]map[string]go_lua.Function {\n")
	for mName, m := range pkg.modules {
		file.indent(1)
		file.fmtWriteln(`"%s": {`, mName)
		for luaName, goName := range m.funcNameMap {
			file.indent(2)
			file.fmtWriteln(`"%s": %s,`, luaName, goName)
		}
		file.indent(1)
		file.write("},\n")
	}
	file.write("}\n\n")
	file.write(`
func PushUserData(vm *go_lua.State, x interface{}) {
	t := reflect.TypeOf(x)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	vm.PushUserData(x)
	if cls, ok := classes[t]; ok {
		go_lua.SetMetaTableNamed(vm, cls.LuaMetaKey)
	}
}

func Register(vm *go_lua.State) {
	for mName, funcMap := range modules {
		vm.CreateTable(0, len(funcMap))
		for k, v := range funcMap {
			vm.PushGoFunction(v)
			vm.SetField(-2, k)
		}
		vm.SetGlobal(mName)
	}
	for _, mt := range classes {
		go2lua.MakeMetaTable(vm, mt)		
	}
}

func init() {
	classes = _classes
}
	`)
	fp := pkg.createFile("registry")
	defer fp.Close()
	file.finish(fp)
}

func (pkg *tPackage) createFile(fileName string) *os.File {
	os.MkdirAll(pkg.path, os.ModePerm)
	fp, err := os.Create(path.Join(pkg.path, fileName+".go"))
	if err != nil {
		panic(err)
	}
	return fp
}

func (pkg *tPackage) addImport(path string) tPkgName {
	if path == "" {
		return tPkgName{}
	}
	if pn, ok := pkg.imports[path]; ok {
		return pn
	}
	i := strings.LastIndexByte(path, '/')
	name := path[i+1:]
	alias := strings.ReplaceAll(name, "-", "_")
	suffixedName := alias
	i = 0
	for {
		if _, ok := pkg.uniNames[suffixedName]; !ok {
			break
		}
		i++
		suffixedName = alias + strconv.Itoa(i)
	}
	alias = suffixedName
	pkg.uniNames[alias] = path
	ans := tPkgName{name: alias, isAlias: alias != name}
	pkg.imports[path] = ans
	return ans
}

func pkgShortName(fullName string) string {
	i := strings.LastIndexByte(fullName, '/')
	return fullName[i+1:]
}
