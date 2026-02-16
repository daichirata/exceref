package exceref

import (
	"bytes"
	"go/format"
	"os"
	"path/filepath"
	"text/template"

	"github.com/daichirata/exceref/internal/errs"
	"github.com/gobuffalo/flect"
	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
)

type GenerateOption struct {
	Prefix       string
	OutDir       string
	TemplatePath string
}

type Generator interface {
	Generate(sheet *Sheet) error
}

func BuildGenerator(lang string, option GenerateOption) Generator {
	switch lang {
	case "go":
		return NewGoGenerator(option)
	case "csharp":
		return NewCsharpGenerator(option)
	default:
		return NewGenerator(option)
	}
}

func NewGenerator(option GenerateOption) *generator {
	return &generator{
		option: option,
	}
}

type Field struct {
	Name       string
	ColumnName string
	Type       string
}

type generator struct {
	option GenerateOption
}

func (g *generator) Generate(sheet *Sheet) error {
	name := g.option.Prefix + sheet.Name

	columns := make([]*Column, 0, len(sheet.Columns))
	for _, c := range sheet.Columns {
		if !c.IsExportable() {
			continue
		}
		columns = append(columns, c)
	}

	var fields []*Field
	for _, column := range columns {
		field := &Field{
			Name:       strcase.ToCamel(column.Name),
			Type:       column.Type.String(),
			ColumnName: column.Name,
		}
		fields = append(fields, field)
	}

	data := map[string]any{
		"Name":   strcase.ToCamel(inflection.Singular(name)),
		"Fields": fields,
	}
	funcMap := map[string]any{
		"camelize":    flect.Camelize,
		"singularize": flect.Singularize,
	}

	templateBody, err := os.ReadFile(g.option.TemplatePath)
	if err != nil {
		return errs.Wrap(err, "read template file")
	}
	tpl := template.Must(template.New("").Funcs(funcMap).Parse(string(templateBody)))

	buf := new(bytes.Buffer)
	if err := tpl.Execute(buf, data); err != nil {
		return errs.Wrap(err, "execute template")
	}
	if err := os.WriteFile(filepath.Join(g.option.OutDir, strcase.ToCamel(inflection.Singular(name))+".gen"), buf.Bytes(), 0644); err != nil {
		return errs.Wrap(err, "write generated file")
	}
	return nil
}

func NewGoGenerator(option GenerateOption) *goGenerator {
	return &goGenerator{
		option: option,
	}
}

type goGenerator struct {
	option GenerateOption
}

func (g *goGenerator) Generate(sheet *Sheet) error {
	name := g.option.Prefix + sheet.Name

	columns := make([]*Column, 0, len(sheet.Columns))
	for _, c := range sheet.Columns {
		if !c.IsExportable() {
			continue
		}
		columns = append(columns, c)
	}

	var fields []*Field
	for _, column := range columns {
		field := &Field{
			Name:       flect.Pascalize(column.Name),
			Type:       g.toGoType(column.Type),
			ColumnName: column.Name,
		}
		fields = append(fields, field)
	}

	imports := g.collectImports(fields)

	data := map[string]any{
		"Imports": imports,
		"Name":    flect.Pascalize(flect.Singularize(name)),
		"Fields":  fields,
	}
	funcMap := map[string]any{
		"camelize":    flect.Camelize,
		"singularize": flect.Singularize,
	}

	templateBody, err := os.ReadFile(g.option.TemplatePath)
	if err != nil {
		return errs.Wrap(err, "read template file")
	}
	tpl := template.Must(template.New("").Funcs(funcMap).Parse(string(templateBody)))

	buf := new(bytes.Buffer)
	if err := tpl.Execute(buf, data); err != nil {
		return errs.Wrap(err, "execute template")
	}
	src, err := format.Source(buf.Bytes())
	if err != nil {
		return errs.Wrap(err, "format generated source")
	}
	if err := os.WriteFile(filepath.Join(g.option.OutDir, name+".gen.go"), src, 0644); err != nil {
		return errs.Wrap(err, "write generated go file")
	}
	return nil
}

func (g *goGenerator) collectImports(columns []*Field) []string {
	importSet := make(map[string]struct{})

	for _, c := range columns {
		switch c.Type {
		case "time.Time":
			importSet["time"] = struct{}{}
		case "civil.Date":
			importSet["cloud.google.com/go/civil"] = struct{}{}
		}
	}

	var imports []string
	for path := range importSet {
		imports = append(imports, path)
	}
	return imports
}

func (g *goGenerator) toGoType(t ColumnType) string {
	switch t {
	case ColumnTypeString:
		return "string"
	case ColumnTypeFloat:
		return "float64"
	case ColumnTypeInt, ColumnTypeUnixtime:
		return "int64"
	case ColumnTypeBool:
		return "bool"
	case ColumnTypeDatetime:
		return "time.Time"
	case ColumnTypeDate:
		return "civil.Date"
	default:
		return "any"
	}
}

func NewCsharpGenerator(option GenerateOption) *csharpGenerator {
	return &csharpGenerator{
		option: option,
	}
}

type csharpGenerator struct {
	option GenerateOption
}

func (g *csharpGenerator) Generate(sheet *Sheet) error {
	name := g.option.Prefix + sheet.Name

	columns := make([]*Column, 0, len(sheet.Columns))
	for _, c := range sheet.Columns {
		if !c.IsExportable() {
			continue
		}
		columns = append(columns, c)
	}

	var fields []*Field
	for _, column := range columns {
		field := &Field{
			Name:       strcase.ToCamel(column.Name),
			Type:       g.toCsharpType(column.Type),
			ColumnName: column.Name,
		}
		fields = append(fields, field)
	}

	data := map[string]any{
		"Name":   strcase.ToCamel(inflection.Singular(name)),
		"Fields": fields,
	}
	funcMap := map[string]any{
		"camelize":    flect.Camelize,
		"singularize": flect.Singularize,
	}

	templateBody, err := os.ReadFile(g.option.TemplatePath)
	if err != nil {
		return errs.Wrap(err, "read template file")
	}
	tpl := template.Must(template.New("").Funcs(funcMap).Parse(string(templateBody)))

	buf := new(bytes.Buffer)
	if err := tpl.Execute(buf, data); err != nil {
		return errs.Wrap(err, "execute template")
	}
	if err := os.WriteFile(filepath.Join(g.option.OutDir, strcase.ToCamel(inflection.Singular(name))+".cs"), buf.Bytes(), 0644); err != nil {
		return errs.Wrap(err, "write generated csharp file")
	}
	return nil
}

func (g *csharpGenerator) toCsharpType(t ColumnType) string {
	switch t {
	case ColumnTypeString:
		return "string"
	case ColumnTypeFloat:
		return "double"
	case ColumnTypeInt, ColumnTypeUnixtime:
		return "int"
	case ColumnTypeBool:
		return "bool"
	case ColumnTypeDatetime:
		return "DateTime"
	case ColumnTypeDate:
		return "DateOnly"
	default:
		return "object"
	}
}
