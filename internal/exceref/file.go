package exceref

import (
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

func (f *File) ReferenceResolver() (*ReferenceResolver, error) {
	if f.resolver != nil {
		return f.resolver, nil
	}
	rows, err := f.xlsx.GetRows(ReferenceDefinitionSheetName)
	if err != nil {
		return nil, err
	}
	resolver, err := NewReferenceResolver(f, NewReferenceDefinitionSheet(ReferenceDefinitionSheetName, rows))
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

func (f *File) UpdateReferenceData() error {
	f.DeleteReferenceData()

	resolver, err := f.ReferenceResolver()
	if err != nil {
		return err
	}
	references, err := resolver.References()
	if err != nil {
		return err
	}
	for _, reference := range references {
		keys := make([]string, len(reference.Keys))
		for i, cell := range reference.Keys {
			keys[i] = cell.Raw
		}
		cell, err := excelize.CoordinatesToCellName(reference.Definition.Index+1, 1)
		if err != nil {
			return err
		}
		f.xlsx.SetSheetCol(ReferenceDataSheetName, cell, &keys)
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
		sheet, err := f.DataSheet(referenceDefinition.Sheet)
		if err != nil {
			return err
		}
		dst, _, err := sheet.Sqrefs(referenceDefinition)
		if err != nil {
			return err
		}
		m[referenceDefinition.Sheet] = append(m[referenceDefinition.Sheet], dst)
	}
	for name := range m {
		f.xlsx.DeleteDataValidation(name) // FIXME
	}
	return nil
}

func (f *File) UpdateDataValidations() error {
	f.DeleteDataValidations()

	resolver, err := f.ReferenceResolver()
	if err != nil {
		return err
	}
	for _, referenceDefinition := range resolver.ReferenceDefinitions {
		sheet, err := f.DataSheet(referenceDefinition.Sheet)
		if err != nil {
			return err
		}
		dst, src, err := sheet.Sqrefs(referenceDefinition)
		if err != nil {
			return err
		}
		dvRange := excelize.NewDataValidation(true)
		dvRange.Sqref = dst
		dvRange.SetSqrefDropList(ReferenceDataSheetName + "!" + src)
		f.xlsx.AddDataValidation(referenceDefinition.Sheet, dvRange)
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
