package clause

const clauseTmpl = `// This file is generated automatically by go generate.  Do not edit.

package {{.PackageName}}

import (
{{range $imp := .Imports}}	"{{$imp}}"
{{end}}
	gormClause "gorm.io/gorm/clause"
)

{{range $typ := .Types}}
type _{{$typ.TypeName}}Table struct {
{{range $field := $typ.Fields}}	{{$field.Name}}	string
{{end}}}

// export all db columns
func (_{{$typ.TypeName}}Table) Columns() []string {
	return []string{ {{range $field := $typ.Fields}}
		"{{$field.Column}}",{{end}}
	}
}

var {{$typ.TypeName}}Table = _{{$typ.TypeName}}Table {
{{range $field := $typ.Fields}}	{{$field.Name}}: "{{$field.Column}}",
{{end}}}
{{end}}

{{range $typ := .Types}}
type {{$typ.TypeName}}Clause struct {
	exprs []gormClause.Expression
}

func (cla *{{$typ.TypeName}}Clause) Clauses() []gormClause.Expression {return cla.exprs}
{{range $field := $typ.Fields}} 
func (cla *{{$field.StructName}}Clause) With{{$field.Name}}(v {{$field.Type}}) *{{$field.StructName}}Clause {
	expr := gormClause.Where{
		Exprs: []gormClause.Expression{gormClause.Eq{Column: {{$typ.TypeName}}Table.{{$field.Name}}, Value: v}},
	}	

	cla.exprs = append(cla.exprs, expr)
	return cla
}
{{end}}
{{end}}
`
