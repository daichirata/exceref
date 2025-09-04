package exceref

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

const (
	ReferenceDefinitionSheetName = "_references"
	ReferenceDataSheetName       = "_reference_data"
)

func Open(path string) (*File, error) {
	file, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	return &File{
		path: path,
		xlsx: file,
		data: make(map[string]*Sheet),
	}, nil
}

type File struct {
	path     string
	xlsx     *excelize.File
	data     map[string]*Sheet
	resolver *ReferenceResolver
}

func (f *File) Name() string {
	base := filepath.Base(f.path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func (f *File) Save() error {
	return f.xlsx.Save()
}

func (f *File) Close() error {
	return f.xlsx.Close()
}

func (f *File) DataSheet(name string) (*Sheet, error) {
	if sheet, ok := f.data[name]; ok {
		return sheet, nil
	}
	rows, err := f.xlsx.GetRows(name)
	if err != nil {
		return nil, err
	}
	sheet, err := NewDataSeet(name, rows)
	if err != nil {
		return nil, err
	}
	f.data[name] = sheet
	return f.data[name], nil
}

func (f *File) ReferenceDefinitionSheet() (*Sheet, error) {
	rows, err := f.xlsx.GetRows(ReferenceDefinitionSheetName)
	if err != nil {
		return nil, err
	}
	return NewReferenceDefinitionSheet(ReferenceDefinitionSheetName, rows), nil
}

func (f *File) ReferenceResolver() (*ReferenceResolver, error) {
	if f.resolver != nil {
		return f.resolver, nil
	}
	sheet, err := f.ReferenceDefinitionSheet()
	if err != nil {
		return nil, err
	}
	resolver, err := NewReferenceResolver(f, sheet)
	if err != nil {
		return nil, err
	}
	f.resolver = resolver
	return f.resolver, nil
}

func (f *File) DeleteReferenceData() {
	f.xlsx.DeleteSheet(ReferenceDataSheetName)
	f.xlsx.NewSheet(ReferenceDataSheetName)
}

func (f *File) DeleteDefinedNames() {
	for _, n := range f.xlsx.GetDefinedName() {
		if strings.HasPrefix("_", n.Name) {
			continue
		}
		if err := f.xlsx.DeleteDefinedName(&n); err != nil {
			slog.Debug("xlsx.DeleteDefinedName", "error", err, "name", n.Name)
		}
	}
}

func (f *File) UpdateReferenceData() error {
	f.DeleteReferenceData()
	f.DeleteDefinedNames()

	resolver, err := f.ReferenceResolver()
	if err != nil {
		return err
	}
	references, err := resolver.References()
	if err != nil {
		return err
	}
	for _, reference := range references {
		if reference.Definition.PolymorphicReference() {
			continue
		}

		keys := make([]string, len(reference.Keys))
		for i, cell := range reference.Keys {
			keys[i] = cell.Raw
		}
		cell, err := excelize.CoordinatesToCellName(reference.Definition.Index+1, 1)
		if err != nil {
			return err
		}
		f.xlsx.SetSheetCol(ReferenceDataSheetName, cell, &keys)

		if reference.Definition.ReferenceName != "" {
			first, err := excelize.CoordinatesToCellName(reference.Definition.Index+1, 1, true)
			if err != nil {
				return err
			}
			last, err := excelize.CoordinatesToCellName(reference.Definition.Index+1, len(keys), true)
			if err != nil {
				return err
			}
			err = f.xlsx.SetDefinedName(&excelize.DefinedName{
				Name:     reference.Definition.ReferenceName,
				RefersTo: fmt.Sprintf("%s!%s:%s", ReferenceDataSheetName, first, last),
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (f *File) DeleteDataValidations() error {
	m := make(map[string][]string)

	resolver, err := f.ReferenceResolver()
	if err != nil {
		return err
	}
	for _, referenceDefinition := range resolver.ReferenceDefinitions {
		if referenceDefinition.Sheet == "" {
			continue
		}
		sheet, err := f.DataSheet(referenceDefinition.Sheet)
		if err != nil {
			return err
		}
		sqref, _, err := sheet.Sqrefs(referenceDefinition)
		if err != nil {
			return err
		}
		m[referenceDefinition.Sheet] = append(m[referenceDefinition.Sheet], sqref)
	}
	for name := range m {
		f.xlsx.DeleteDataValidation(name) // FIXME: pass sqref
		slog.Debug("DeleteDataValidations", "sheet", name)
	}
	return nil
}

func (f *File) UpdateDataValidations() error {
	if err := f.DeleteDataValidations(); err != nil {
		return err
	}

	resolver, err := f.ReferenceResolver()
	if err != nil {
		return err
	}
	references, err := resolver.References()
	if err != nil {
		return err
	}
	for _, reference := range references {
		if reference.Definition.Sheet == "" {
			continue
		}
		sheet, err := f.DataSheet(reference.Definition.Sheet)
		if err != nil {
			return err
		}
		sqref, srcSqref, err := sheet.Sqrefs(reference.Definition)
		if err != nil {
			return err
		}
		dvRange := excelize.NewDataValidation(true)
		dvRange.Sqref = sqref
		if reference.Definition.PolymorphicReference() {
			name, err := excelize.ColumnNumberToName(reference.KeyColumn.Index + 1)
			if err != nil {
				return err
			}
			dvRange.SetSqrefDropList(fmt.Sprintf("INDIRECT($%s4)", name))
		} else {
			dvRange.SetSqrefDropList(ReferenceDataSheetName + "!" + srcSqref)
		}
		f.xlsx.AddDataValidation(reference.Definition.Sheet, dvRange)

		slog.Debug("AddDataValidation", "sheet", reference.Definition.Sheet, "dv", dvRange)
	}
	return nil
}

func (f *File) Export(exporter Exporter) error {
	resolver, err := f.ReferenceResolver()
	if err != nil {
		return err
	}

	for _, name := range f.xlsx.GetSheetMap() {
		if strings.HasPrefix(name, "_") {
			continue
		}

		sheet, err := f.DataSheet(name)
		if err != nil {
			return err
		}
		if err := resolver.Resolve(sheet); err != nil {
			return err
		}
		if err := exporter.Export(sheet); err != nil {
			return err
		}
	}
	return nil
}

func (f *File) ExportMetadata(outDir string) error {
	return NewMetadataExporter(outDir).Export(f)
}

func (f *File) Generate(generator Generator) error {
	resolver, err := f.ReferenceResolver()
	if err != nil {
		return err
	}

	for _, name := range f.xlsx.GetSheetMap() {
		if strings.HasPrefix(name, "_") {
			continue
		}

		sheet, err := f.DataSheet(name)
		if err != nil {
			return err
		}
		if err := resolver.Resolve(sheet); err != nil {
			return err
		}
		if err := generator.Generate(sheet); err != nil {
			return err
		}
	}
	return nil
}
