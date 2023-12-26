package exceref

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Exporter interface {
	Export(sheet *Sheet) error
}

func BuildExporter(format, outDir string) Exporter {
	switch format {
	case "json":
		return NewJSONExporter(outDir)
	case "yaml":
		return NewYAMLExporter(outDir)
	default:
		return NewCSVExporter(outDir)
	}
}

func NewCSVExporter(outDir string) *csvExporter {
	return &csvExporter{outDir: outDir}
}

type csvExporter struct {
	outDir string
}

func (e *csvExporter) Export(sheet *Sheet) error {
	f, err := os.Create(filepath.Join(e.outDir, sheet.Name+".csv"))
	if err != nil {
		return err
	}
	defer f.Close()

	columns := make([]*Column, 0, len(sheet.Columns))
	for _, c := range sheet.Columns {
		if c.Name == "" {
			continue
		}
		columns = append(columns, c)
	}

	writer := csv.NewWriter(f)

	headers := make([]string, len(columns))
	for i, column := range columns {
		headers[i] = column.Name
	}
	if err := writer.Write(headers); err != nil {
		return err
	}
	for _, row := range sheet.Rows {
		records := make([]string, len(columns))
		for i, column := range columns {
			value, err := e.toCSVString(row[column.Index].Value)
			if err != nil {
				return err
			}
			records[i] = value
		}
		if err := writer.Write(records); err != nil {
			return err
		}
	}
	writer.Flush()

	return nil
}

func (e *csvExporter) toCSVString(value any) (string, error) {
	switch t := value.(type) {
	case string, int, int64, float64, bool:
		return fmt.Sprint(t), nil
	case time.Time:
		return t.Format(time.RFC3339), nil
	}
	return "", fmt.Errorf("unmatched type:%#v", value)
}

func NewJSONExporter(outDir string) *jsonExporter {
	return &jsonExporter{outDir: outDir}
}

type jsonExporter struct {
	outDir string
}

func (e *jsonExporter) Export(sheet *Sheet) error {
	f, err := os.Create(filepath.Join(e.outDir, sheet.Name+".json"))
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(sheet.Map())
}

func NewYAMLExporter(outDir string) *yamlExporter {
	return &yamlExporter{outDir: outDir}
}

type yamlExporter struct {
	outDir string
}

func (e *yamlExporter) Export(sheet *Sheet) error {
	f, err := os.Create(filepath.Join(e.outDir, sheet.Name+".yaml"))
	if err != nil {
		return err
	}
	defer f.Close()

	return yaml.NewEncoder(f).Encode(sheet.Map())
}
