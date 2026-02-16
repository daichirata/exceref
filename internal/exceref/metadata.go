package exceref

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/daichirata/exceref/internal/errs"
	"gopkg.in/yaml.v3"
)

type MetadataDataYAML struct {
	Sheet  string                 `yaml:"sheet"`
	Schema []MetadataColumnSchema `yaml:"schema"`
}

type MetadataColumnSchema struct {
	Name        string           `yaml:"name"`
	Type        ColumnType       `yaml:"type"`
	DisplayName string           `yaml:"display_name"`
	Ref         *MetadataRefSpec `yaml:"ref,omitempty"`
}

type MetadataRefSpec struct {
	File  string `yaml:"file"`
	Sheet string `yaml:"sheet"`
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

type MetadataReferencesYAML struct {
	References []MetadataReference `yaml:"references"`
}

type MetadataReference struct {
	Sheet          string `yaml:"sheet"`
	Column         string `yaml:"column"`
	ReferenceFile  string `yaml:"reference_file"`
	ReferenceSheet string `yaml:"reference_sheet"`
	ReferenceKey   string `yaml:"reference_key"`
	ReferenceValue string `yaml:"reference_value"`
	ReferenceName  string `yaml:"reference_name"`
}

func NewMetadataExporter(outDir string) *metadataExporter {
	return &metadataExporter{
		outDir: outDir,
	}
}

type metadataExporter struct {
	outDir string
}

func (e *metadataExporter) Export(file *File) error {
	referencesYaml := &MetadataReferencesYAML{}

	resolver, err := file.ReferenceResolver()
	if err != nil {
		return errs.Wrap(err, "load reference resolver")
	}
	for _, definition := range resolver.ReferenceDefinitions {
		referencesYaml.References = append(referencesYaml.References, MetadataReference{
			Sheet:          definition.Sheet,
			Column:         definition.Column,
			ReferenceFile:  definition.ReferenceFile,
			ReferenceSheet: definition.ReferenceSheet,
			ReferenceKey:   definition.ReferenceKey,
			ReferenceValue: definition.ReferenceValue,
			ReferenceName:  definition.ReferenceName,
		})
	}

	var dataYamls []*MetadataDataYAML
	for _, name := range file.xlsx.GetSheetMap() {
		dataYaml := &MetadataDataYAML{}
		switch {
		case strings.HasPrefix(name, "_") && name != "_types":
			continue
		case name == "_types":
			dataYaml.Sheet = file.Name() + "_types"
		default:
			dataYaml.Sheet = name
		}

		sheet, err := file.DataSheet(name)
		if err != nil {
			return errs.Wrap(err, "load metadata target sheet")
		}
		for _, col := range sheet.Columns {
			schema := MetadataColumnSchema{
				Name:        col.Name,
				Type:        col.Type,
				DisplayName: col.Description,
			}
			if col.Type == ColumnTypeRef {
				for _, reference := range referencesYaml.References {
					if !(sheet.Name == reference.Sheet && col.Name == reference.Column) {
						continue
					}
					schema.Ref = &MetadataRefSpec{
						File:  reference.ReferenceFile,
						Sheet: reference.ReferenceSheet,
						Key:   reference.ReferenceKey,
						Value: reference.ReferenceValue,
					}
				}
			}
			dataYaml.Schema = append(dataYaml.Schema, schema)
		}
		dataYamls = append(dataYamls, dataYaml)
	}

	if err := e.write(file.Name()+ReferenceDefinitionSheetName+".yaml", referencesYaml); err != nil {
		return errs.Wrap(err, "write metadata reference yaml")
	}
	for _, dataYaml := range dataYamls {
		if err := e.write(dataYaml.Sheet+".yaml", dataYaml); err != nil {
			return errs.Wrap(err, "write metadata data yaml")
		}
	}
	return nil
}

func (e *metadataExporter) write(name string, v any) error {
	f, err := os.Create(filepath.Join(e.outDir, name))
	if err != nil {
		return errs.Wrap(err, "create metadata file")
	}
	defer f.Close()

	return errs.Wrap(yaml.NewEncoder(f).Encode(v), "encode metadata yaml")
}
