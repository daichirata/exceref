package exceref

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/daichirata/exceref/internal/errs"
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

func (r *ReferenceDefinition) PolymorphicReference() bool {
	return r.ReferenceValue == ""
}

type Reference struct {
	Definition  *ReferenceDefinition
	KeyColumn   *Column
	ValueColumn *Column
	Keys        []*Cell
	Values      []*Cell
	ValueMap    map[string]*Cell
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
		if definition.PolymorphicReference() {
			if definition.Sheet != definition.ReferenceSheet {
				return nil, fmt.Errorf("PolymorphicReference Sheet(%s) and ReferenceSheet(%s) must match", definition.Sheet, definition.ReferenceSheet)
			}
		}
		resolver.ReferenceDefinitions = append(resolver.ReferenceDefinitions, definition)
	}
	return resolver, nil
}

type ReferenceResolver struct {
	SheetReader          SheetReader
	ReferenceDefinitions []*ReferenceDefinition

	references []*Reference
}

func (r *ReferenceResolver) References() ([]*Reference, error) {
	if r.references != nil {
		return r.references, nil
	}
	r.references = make([]*Reference, len(r.ReferenceDefinitions))

	for i, referenceDefinition := range r.ReferenceDefinitions {
		referenceSheet, err := r.SheetReader.Open(referenceDefinition.ReferenceFilePath(), referenceDefinition.ReferenceSheet)
		if err != nil {
			return nil, errs.Wrap(err, "open reference sheet")
		}

		if referenceDefinition.PolymorphicReference() {
			k, err := referenceSheet.Column(referenceDefinition.ReferenceKey)
			if err != nil {
				return nil, errs.Wrap(err, "find polymorphic reference key column")
			}
			reference := &Reference{
				Definition: referenceDefinition,
				KeyColumn:  referenceSheet.Columns[k.Index],
			}
			r.references[i] = reference
		} else {
			k, err := referenceSheet.Column(referenceDefinition.ReferenceKey)
			if err != nil {
				return nil, errs.Wrap(err, "find reference key column")
			}
			v, err := referenceSheet.Column(referenceDefinition.ReferenceValue)
			if err != nil {
				return nil, errs.Wrap(err, "find reference value column")
			}
			reference := &Reference{
				Definition:  referenceDefinition,
				KeyColumn:   referenceSheet.Columns[k.Index],
				ValueColumn: referenceSheet.Columns[v.Index],
				Keys:        make([]*Cell, len(referenceSheet.Rows)),
				Values:      make([]*Cell, len(referenceSheet.Rows)),
				ValueMap:    make(map[string]*Cell),
			}
			for j, row := range referenceSheet.Rows {
				reference.Keys[j] = row[k.Index]
				reference.Values[j] = row[v.Index]
				reference.ValueMap[reference.Keys[j].Raw] = reference.Values[j]
			}
			r.references[i] = reference
		}
	}
	return r.references, nil
}

func (r *ReferenceResolver) Resolve(sheet *Sheet) error {
	references, err := r.References()
	if err != nil {
		return errs.Wrap(err, "load references for resolve")
	}

	// Resolve the poymorphic reference first, since poymorphic reference resolution cannot be performed
	// if the value of reference_key has already been rewritten.
	names := make(map[string]*Reference)
	for _, r := range references {
		if name := r.Definition.ReferenceName; name != "" {
			names[name] = r
		}
	}
	for _, reference := range references {
		if reference.Definition.Sheet != sheet.Name {
			continue
		}
		for _, column := range sheet.Columns {
			if reference.Definition.Column != column.Name {
				continue
			}
			if !reference.Definition.PolymorphicReference() {
				continue
			}

			for i, row := range sheet.Rows {
				r, ok := names[row[reference.KeyColumn.Index].Raw]
				if !ok {
					return fmt.Errorf("sheet:%s row:%d column:%s reference_name:%s not found",
						sheet.Name, i+1, column.Name, row[reference.KeyColumn.Index].Raw)
				}
				if v, ok := r.ValueMap[row[column.Index].Raw]; ok {
					row[column.Index] = v
				} else {
					return fmt.Errorf("sheet:%s row:%d column:%s reference:%s value not found from %s:%s",
						sheet.Name, i+1, column.Name, row[column.Index].Raw, reference.Definition.ReferenceSheet, reference.Definition.ReferenceKey)
				}
				if column.Type == "" || column.Type == "ref" {
					column.Type = r.ValueColumn.Type
				} else if column.Type != r.ValueColumn.Type {
					return fmt.Errorf("sheet:%s row:%d column:%s value type mismatch: %s, %s", sheet.Name, i+1, column.Name, column.Type, r.ValueColumn.Type)
				}
			}
		}
	}

	// Resolve a regular reference after resolving a poymorphic reference
	for _, reference := range references {
		if reference.Definition.Sheet != sheet.Name {
			continue
		}
		for _, column := range sheet.Columns {
			if reference.Definition.Column != column.Name {
				continue
			}
			if reference.Definition.PolymorphicReference() {
				continue
			}

			for i, row := range sheet.Rows {
				if row[column.Index].Raw == "" {
					continue
				}
				if v, ok := reference.ValueMap[row[column.Index].Raw]; ok {
					row[column.Index] = v
				} else {
					return fmt.Errorf("sheet: %s, row: %d, column: %s, reference: %s, value not found from %s:%s",
						sheet.Name, i+1, column.Name, row[column.Index].Raw, reference.Definition.ReferenceSheet, reference.Definition.ReferenceKey)
				}
			}
			column.Type = reference.ValueColumn.Type
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
		return nil, errs.Wrap(err, "open reference file")
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
