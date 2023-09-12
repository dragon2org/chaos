package clause

const scopeTmpl = `// This file is generated automatically by go generate.  Do not edit.

package {{.PackageName}}

import (
	"gorm.io/gorm"
)

{{range $typ := .Types}}
type {{$typ.TypeName}}Scope struct {
	scopes []func(db *gorm.DB) *gorm.DB
}

func (scope *{{$typ.TypeName}}Scope) Scopes() []func(*gorm.DB) *gorm.DB {return scope.scopes}
{{range $field := $typ.Fields}} 
func (scope *{{$field.StructName}}Scope) With{{$field.Name}}(v {{$field.Type}}) *{{$field.StructName}}Scope {
	query := func(db *gorm.DB) *gorm.DB {
		return db.Where("{{$field.Column}} = ?", v)
	}

	scope.scopes = append(scope.scopes, query)
	return scope
}
{{end}}
{{end}}
`

type Scope struct {
	PackageName string
	Types       []Typ
	Imports     []string
}

type Typ struct {
	TypeName string
	Fields   []Field
}

type Field struct {
	StructName string
	Column     string
	Name       string
	Type       string
}
