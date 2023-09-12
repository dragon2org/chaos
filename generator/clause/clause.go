package clause

import (
	"go/ast"
	"go/types"
	"io"
	"os"
	"path"
	"text/template"

	"golang.org/x/tools/go/packages"
)

func GenClause(typeNames []string, output string) {
	g := Generator{
		output:   output,
		tmplText: clauseTmpl,
	}
	err := g.parsePackage()
	checkErr(err)

	models := make([]*Model, 0, len(typeNames))
	for _, typ := range typeNames {
		m, err := g.loadType(typ)
		checkErr(err)

		models = append(models, m)
	}

	err = g.render(models)
	checkErr(err)
}

func GenClauseV2(typeNames []string, output string, dest string) {
	generator := NewGenerator(typeNames, output, dest)
	generator.Execute()
}

type TypeObject struct {
	obj  types.Object
	ppkg *packages.Package
}

type GeneratorV2 struct {
	typeNames   []string
	output      string
	dest        string // destination package name
	selfPackage bool   // self package path
	selfPath    string // current package path

	necessaryTypeObject map[string]TypeObject
}

func NewGenerator(typeNames []string, output string, dest string) *GeneratorV2 {
	dir, _ := path.Split(output)
	selfPkg := dest == ""
	if dir != "" {
		selfPkg = false
	}

	return &GeneratorV2{
		typeNames:           typeNames,
		output:              output,
		dest:                dest,
		selfPackage:         selfPkg,
		necessaryTypeObject: map[string]TypeObject{},
	}
}

func (g *GeneratorV2) Execute() {
	g.parsePackages()
	scope := g.analyzeField()
	scope.PackageName = g.dest
	checkErr(g.rendering(scope))
}

func (g *GeneratorV2) astVisitor() func(node ast.Node) bool {
	names := map[string]bool{}
	for _, v := range g.typeNames {
		names[v] = true
	}

	return func(node ast.Node) bool {
		ts, ok := node.(*ast.TypeSpec)
		if !ok {
			return true
		}

		if _, ok := ts.Type.(*ast.StructType); !ok {
			return true
		}

		if !names[ts.Name.String()] {
			return true
		}

		return false
	}
}

func (g *GeneratorV2) parsePackages() {
	cfg := packages.Config{
		Mode: packages.NeedImports | packages.NeedTypes | packages.NeedTypesInfo |
			packages.NeedSyntax | packages.NeedName,
		Tests: true,
	}

	pkgs, err := packages.Load(&cfg, ".")
	checkErr(err)

	var autoFillTypeObj = func(ppkg *packages.Package) {
		for _, typeName := range g.typeNames {
			if obj := ppkg.Types.Scope().Lookup(typeName); obj != nil {
				g.necessaryTypeObject[typeName] = TypeObject{obj: obj, ppkg: ppkg}
			}
		}
	}

	for _, pkg := range pkgs {
		autoFillTypeObj(pkg)
	}

	g.selfPath = pkgs[0].PkgPath
	if g.dest == "" {
		g.dest = pkgs[0].Name
	}
}

func (g *GeneratorV2) loadTypeObject(typeName string) TypeObject {
	return g.necessaryTypeObject[typeName]
}

func (g *GeneratorV2) handleTypesStruct(s *types.Struct, structName string) ([]Field, []string) {
	var fields []Field
	var imports []string
	for i := 0; i < s.NumFields(); i++ {
		if field := s.Field(i); field.Exported() && field.Embedded() {
			if ss, ok := field.Type().Underlying().(*types.Struct); ok {
				sf, _imports := g.handleTypesStruct(ss, structName)
				fields = append(fields, sf...)
				imports = append(imports, _imports...)
			}

			continue
		}

		sf, _imports := g.handleField(s.Field(i), s.Tag(i), structName)
		if sf == nil {
			continue
		}
		fields = append(fields, *sf)
		imports = append(imports, _imports...)
	}

	return fields, imports
}

func (g *GeneratorV2) handleField(field *types.Var, tag string, structName string) (*Field, []string) {
	if !field.IsField() || !field.Exported() || sqlSkipTag(tag) {
		return nil, nil
	}

	column, exist := findColumnNameFromTag(tag)
	if !exist {
		return nil, nil
	}

	var imports []string
	var ft = field.Type()
	var fieldType string
	var pointerTyp bool

	switch tp := ft.(type) {
	case *types.Pointer:
		ft = tp.Elem()
		pointerTyp = true
	case *types.Basic:
		fieldType = tp.String()
	}

	if named, ok := ft.(*types.Named); ok {
		obj := named.Obj()
		if obj.Pkg().Path() == g.selfPath && g.selfPackage {
			fieldType = obj.Name()
		} else {
			fieldType = obj.Pkg().Name() + "." + obj.Name()
			if pointerTyp {
				fieldType = "*" + fieldType
			}
			imports = append(imports, named.Obj().Pkg().Path())
		}
	}

	sf := Field{
		StructName: structName,
		Column:     column,
		Name:       field.Name(),
		Type:       fieldType,
	}

	return &sf, imports
}

func (g *GeneratorV2) analyzeField() Scope {
	var scope Scope
	var imports []string

	removeDuplicate := func(slice []string) []string {
		var (
			keys = map[string]bool{}
			list []string
		)

		for _, e := range slice {
			if _, exist := keys[e]; exist {
				continue
			}
			keys[e] = true
			list = append(list, e)
		}

		return list
	}

	for _, typo := range g.necessaryTypeObject {
		s, ok := typo.obj.Type().Underlying().(*types.Struct)
		if !ok {
			continue
		}

		fields, _imports := g.handleTypesStruct(s, typo.obj.Name())

		imports = append(imports, _imports...)
		scope.Types = append(scope.Types, Typ{TypeName: typo.obj.Name(), Fields: fields})
	}

	scope.Imports = removeDuplicate(imports)

	return scope
}

func (g GeneratorV2) rendering(scope Scope) error {
	var output io.Writer = os.Stdout
	if g.output != "" {
		err := prepareDir(g.output)
		if err != nil {
			return err
		}

		file, err := os.Create(g.output)
		if err != nil {
			return err
		}
		defer func() {
			_ = file.Sync()
			_ = file.Close()
		}()

		output = file
	}

	t := template.Must(template.New("clause").Parse(clauseTmpl))
	return t.Execute(output, scope)
}
