package exceref

import (
	"fmt"
	"path/filepath"
	"strings"
)

type ReferenceDefinition struct {
	Index          int
	BaseDir        string
	Sheet          string
	Column         string
	ReferenceFile  string
	ReferenceSheet string
	ReferenceKey   string
	ReferenceValue string
	ReferenceName  string
}

func (r *ReferenceDefinition) ReferenceFileName() string {
	if strings.Contains(r.ReferenceFile, ".") {
		return r.ReferenceFile
	}
	return r.ReferenceFile + ".xlsx"
}

func (r *ReferenceDefinition) ReferenceFilePath() string {
	return filepath.Join(r.BaseDir, r.ReferenceFileName())
}

type Reference struct {
	Definition *ReferenceDefinition
	Column     *Column
	Keys       []*Cell
	Values     []*Cell
}

func NewReferenceResolver(file *File, sheet *Sheet) (*ReferenceResolver, error) {
	resolver := &ReferenceResolver{SheetReader: NewXLSXReader()}

	for i, row := range sheet.Rows {
		definition := &ReferenceDefinition{
			Index:   i,
			BaseDir: filepath.Dir(file.path),
		}
		for _, cell := range row {
			switch cell.Column.Name {
			case "sheet":
				definition.Sheet = cell.Raw
			case "column":
				definition.Column = cell.Raw
			case "reference_file":
				definition.ReferenceFile = cell.Raw
			case "reference_sheet":
				definition.ReferenceSheet = cell.Raw
			case "reference_key":
				definition.ReferenceKey = cell.Raw
			case "reference_value":
				definition.ReferenceValue = cell.Raw
			case "reference_name":
				definition.ReferenceName = cell.Raw
			default:
				return nil, fmt.Errorf("unknown column: %s", cell.Column.Name)
			}
		}
		resolver.ReferenceDefinitions = append(resolver.ReferenceDefinitions, definition)
	}
	return resolver, nil
}

type ReferenceResolver struct {
	SheetReader          SheetReader
	ReferenceDefinitions []*ReferenceDefinition
}

func (r *ReferenceResolver) References() ([]*Reference, error) {
	references := make([]*Reference, len(r.ReferenceDefinitions))

	for i, referenceDefinition := range r.ReferenceDefinitions {
		referenceSheet, err := r.SheetReader.Open(referenceDefinition.ReferenceFilePath(), referenceDefinition.ReferenceSheet)
		if err != nil {
			return nil, err
		}

		k, err := referenceSheet.Column(referenceDefinition.ReferenceKey)
		if err != nil {
			return nil, err
		}
		v, err := referenceSheet.Column(referenceDefinition.ReferenceValue)
		if err != nil {
			return nil, err
		}
		reference := &Reference{
			Definition: referenceDefinition,
			Column:     referenceSheet.Columns[v.Index],
			Keys:       make([]*Cell, len(referenceSheet.Rows)),
			Values:     make([]*Cell, len(referenceSheet.Rows)),
		}
		for j, row := range referenceSheet.Rows {
			reference.Keys[j] = row[k.Index]
			reference.Values[j] = row[v.Index]
		}
		references[i] = reference
	}
	return references, nil
}

func (r *ReferenceResolver) Resolve(sheet *Sheet) error {
	references, err := r.References()
	if err != nil {
		return err
	}
	for _, reference := range references {
		if reference.Definition.Sheet != sheet.Name {
			continue
		}
		for i, column := range sheet.Columns {
			if reference.Definition.Column != column.Name {
				continue
			}
			sheet.Columns[i].Type = reference.Column.Type

			m := make(map[string]*Cell)
			for i, k := range reference.Keys {
				m[k.Raw] = reference.Values[i]
			}
			for _, row := range sheet.Rows {
				if v, ok := m[row[column.Index].Raw]; ok {
					row[column.Index] = v
				}
			}
		}
	}
	return nil
}

type SheetReader interface {
	Open(file string, sheet string) (*Sheet, error)
}

func NewXLSXReader() SheetReader {
	return &XLSXReader{file: make(map[string]*File)}
}

type XLSXReader struct {
	file map[string]*File
}

func (r *XLSXReader) Open(path string, sheet string) (*Sheet, error) {
	if f, ok := r.file[path]; ok {
		return f.DataSheet(sheet)
	}
	file, err := Open(path)
	if err != nil {
		return nil, err
	}
	r.file[path] = file
	return file.DataSheet(sheet)
}

type MemoryReader struct {
	Sheet map[string]*Sheet
}

func (r *MemoryReader) Open(path string, sheet string) (*Sheet, error) {
	if s, ok := r.Sheet[sheet]; ok {
		return s, nil
	}
	return NewDataSeet(sheet, nil)
}
