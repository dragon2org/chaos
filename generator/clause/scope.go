package clause

import (
	"errors"
	"fmt"
	"go/ast"
	"go/types"
	"html/template"
	"os"
	"path"
	"regexp"

	"golang.org/x/tools/go/packages"
)

var ErrNoPackageFound = errors.New("未提供 package")

type (
	ErrTypeNotFound  string
	ErrNotNamedType  string
	ErrNotStructType string
)

func (e ErrTypeNotFound) Error() string {
	return string(e) + " type not found"
}

func (e ErrNotNamedType) Error() string {
	return string(e) + " not named type"
}

func (e ErrNotStructType) Error() string {
	return string(e) + " not struct type"
}

func checkErr(e error) {
	if e == nil {
		return
	}

	fmt.Println(e)
	os.Exit(1)
}

func Gen(typeNames []string, output string) {
	g := Generator{
		output:   output,
		tmplText: scopeTmpl,
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

type Model struct {
	obj types.Object
	pkg *packages.Package
}

type Generator struct {
	output string

	tmplText string

	pkgs []*packages.Package
}

func prepareDir(p string) error {
	dir, _ := path.Split(p)
	if dir == "" {
		return nil
	}

	_, err := os.Stat(dir)
	if err == nil {
		return nil
	}

	if os.IsNotExist(err) {
		return os.MkdirAll(dir, 0766)
	}

	return err
}

func (g Generator) renderTemplate(scope Scope) error {
	t := template.Must(template.New("scope").Parse(g.tmplText))
	writer := os.Stdout
	if g.output != "" {
		if err := prepareDir(g.output); err != nil {
			return err
		}
		file, err := os.Create(g.output)
		if err != nil {
			return err
		}
		defer file.Close()
		writer = file
	}

	return t.Execute(writer, scope)
}

func (g *Generator) parsePackage() error {
	cfg := packages.Config{
		Mode: packages.NeedImports | packages.NeedTypes | packages.NeedName |
			packages.NeedSyntax | packages.NeedTypesInfo,
		Tests: true,
	}

	pkgs, err := packages.Load(&cfg, ".")
	if err != nil {
		panic(fmt.Errorf("load package failed %w", err))
	}

	if len(pkgs) == 0 {
		return ErrNoPackageFound
	}

	g.pkgs = pkgs

	return nil
}

func (g *Generator) loadType(typeName string) (*Model, error) {
	var find = func(name string) (types.Object, *packages.Package, error) {
		for _, pkg := range g.pkgs {
			obj := pkg.Types.Scope().Lookup(name)
			if obj == nil {
				continue
			}

			return obj, pkg, nil
		}
		return nil, nil, ErrTypeNotFound(name)
	}

	obj, pkg, err := find(typeName)
	if err != nil {
		return nil, err
	}

	if _, ok := obj.(*types.TypeName); !ok {
		return nil, ErrNotNamedType(obj.Name())
	}

	m := &Model{obj: obj, pkg: pkg}
	return m, nil
}

func (g Generator) astInspector(model *Model, fields *[]Field, imports map[string]bool) func(node ast.Node) bool {
	obj := model.obj
	return func(node ast.Node) bool {
		t, ok := node.(*ast.TypeSpec)
		if !ok || t.Name.Name != obj.Name() {
			return true
		}

		s, ok := t.Type.(*ast.StructType)
		if !ok {
			return true
		}

		for _, field := range s.Fields.List {
			name, column, skip := checkFieldSkip(field)
			if skip {
				continue
			}

			sf := Field{
				StructName: t.Name.Name,
				Column:     column,
				Name:       name,
			}
			switch ft := field.Type.(type) {
			case *ast.Ident:
				sf.Type = ft.Name

			case *ast.SelectorExpr:
				sel := model.pkg.TypesInfo.Uses[ft.Sel]
				imports[sel.Pkg().Path()] = true
				sf.Type = sel.Pkg().Name() + "." + sel.Name()
			default:
				// not supported type
				continue
			}

			*fields = append(*fields, sf)
		}

		return false
	}
}

func (g *Generator) render(models []*Model) error {
	scope := Scope{}
	imports := map[string]bool{}

	for _, model := range models {
		obj := model.obj
		pkg := model.pkg
		scope.PackageName = model.pkg.Name

		var fields = make([]Field, 0)
		for _, file := range pkg.Syntax {
			ast.Inspect(file, g.astInspector(model, &fields, imports))
		}
		scope.Types = append(scope.Types, Typ{TypeName: obj.Name(), Fields: fields})
	}

	for k := range imports {
		scope.Imports = append(scope.Imports, k)
	}
	return g.renderTemplate(scope)
}

func findColumnNameFromTag(tag string) (string, bool) {
	pattern := regexp.MustCompile(`column:(\w+)`)
	results := pattern.FindAllStringSubmatch(tag, -1)

	if len(results) == 0 {
		return "", false
	}

	regexp.MustCompile(`sql:"(\w+)"`)

	return results[0][1], true
}

func sqlSkipTag(tag string) bool {
	return regexp.MustCompile(`sql:"-"`).MatchString(tag) || regexp.MustCompile(`gorm:"-"`).MatchString(tag)
}

// checkFieldSkip 检查 field 是否需要跳过
// @return (fieldName, columnName, skip)
func checkFieldSkip(field *ast.Field) (string, string, bool) {
	// bypass type "type"
	if len(field.Names) == 0 || !field.Names[0].IsExported() {
		return "", "", true
	}

	// bypass field if tag not exist
	if field.Tag == nil {
		return "", "", true
	}

	if sqlSkipTag(field.Tag.Value) {
		return "", "", true
	}

	column, exist := findColumnNameFromTag(field.Tag.Value)
	return field.Names[0].Name, column, !exist
}
