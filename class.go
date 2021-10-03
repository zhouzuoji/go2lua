package go2lua

import "reflect"

func (pkg *tPackage) AddClass(x interface{}) tClass {
	t, ok := x.(reflect.Type)
	if !ok {
		t = reflect.TypeOf(x)
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		panic("not struct")
	}
	cls, ok := pkg.classes[t]
	if ok {
		return cls
	}
	path := t.PkgPath()
	pkgName := pkg.addImport(path)
	pkg.classes[t] = tClass{
		t:       t,
		metaKey: path + "." + t.Name(),
		pkgName: pkgName.name,
	}
	return pkg.classes[t]
}

func (pkg *tPackage) registerClass(cls tClass) {
	file := pkg.NewFile()
	file.writeClass(cls)
	fp := pkg.createFile(cls.pkgName + "-" + cls.t.Name())
	defer fp.Close()
	file.finish(fp)
}
