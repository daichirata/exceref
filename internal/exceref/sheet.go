package exceref

import (
	"fmt"
	"strconv"
	"time"

	"github.com/samber/lo"
	"github.com/xuri/excelize/v2"
)

const (
	DataSheetIndexColumnType        = 0
	DataSheetIndexColumnName        = 1
	DataSheetIndexColumnDescription = 2
	DataSheetIndexBody              = 3

	ReferenceDefinitionSheetIndexColumnName = 0
	ReferenceDefinitionSheetIndexBody       = 1
)

type Column struct {
	Name        string
	Type        string
	Index       int
	Description string
}

type Row []*Cell

func (r Row) Cells(names ...string) []*Cell {
	cells := make([]*Cell, 0, len(names))
	for _, cell := range r {
		for _, name := range names {
			if cell.Column.Name == name {
				cells = append(cells, cell)
			}
		}
	}
	return cells
}

type Cell struct {
	Column *Column
	Value  any
	Raw    string
}

type Sheet struct {
	Name    string
	Columns []*Column
	Rows    []Row
}

func (s *Sheet) Column(name string) (*Column, error) {
	for _, c := range s.Columns {
		if name == c.Name {
			return c, nil
		}
	}
	return nil, fmt.Errorf("sheet:%s column:%s not found", s.Name, name)
}

func (s *Sheet) Map() []map[string]any {
	data := make([]map[string]any, len(s.Rows))

	for i, row := range s.Rows {
		data[i] = make(map[string]any)

		for _, column := range s.Columns {
			if column.Name == "" {
				continue
			}
			data[i][column.Name] = row[column.Index].Value
		}
	}
	return data
}

func (s *Sheet) Sqrefs(referenceDefinition *ReferenceDefinition) (string, string, error) {
	column, err := s.Column(referenceDefinition.Column)
	if err != nil {
		return "", "", err
	}
	dstBegin, err := excelize.CoordinatesToCellName(column.Index+1, DataSheetIndexBody+1)
	if err != nil {
		return "", "", err
	}
	dstEnd, err := excelize.CoordinatesToCellName(column.Index+1, 9999)
	if err != nil {
		return "", "", err
	}
	dstSqref := fmt.Sprintf("%s:%s", dstBegin, dstEnd)

	srcBegin, err := excelize.CoordinatesToCellName(referenceDefinition.Index+1, 1, true)
	if err != nil {
		return "", "", err
	}
	srcEnd, err := excelize.CoordinatesToCellName(referenceDefinition.Index+1, 9999, true)
	if err != nil {
		return "", "", err
	}
	srcSqref := fmt.Sprintf("%s:%s", srcBegin, srcEnd)

	return dstSqref, srcSqref, nil
}

func NewDataSeet(name string, rows [][]string) (*Sheet, error) {
	sheet := &Sheet{
		Name: name,
	}
	for i, r := range rows {
		switch i {
		case DataSheetIndexColumnType:
			for j, value := range r {
				sheet.Columns = append(sheet.Columns, &Column{
					Type:  value,
					Index: j,
				})
			}
		case DataSheetIndexColumnName:
			for j, value := range r {
				sheet.Columns[j].Name = value
			}
		case DataSheetIndexColumnDescription:
			for j, value := range r {
				sheet.Columns[j].Description = value
			}
		default:
			var row Row
			for _, column := range sheet.Columns {
				var rawValue string
				if len(r) > column.Index {
					rawValue = r[column.Index]
				}
				value, err := parseValue(column.Type, rawValue)
				if err != nil {
					return nil, err
				}
				row = append(row, &Cell{Column: column, Value: value, Raw: rawValue})
			}
			sheet.Rows = append(sheet.Rows, row)
		}
	}
	return sheet, nil
}

func NewReferenceDefinitionSheet(name string, rows [][]string) *Sheet {
	sheet := &Sheet{
		Name: name,
	}
	for i, r := range rows {
		switch i {
		case ReferenceDefinitionSheetIndexColumnName:
			for i, value := range r {
				if value == "" {
					continue
				}
				sheet.Columns = append(sheet.Columns, &Column{
					Type:  "string",
					Name:  value,
					Index: i,
				})
			}
		default:
			var row Row
			for _, column := range sheet.Columns {
				row = append(row, &Cell{
					Column: column,
					Value:  lo.NthOr(r, column.Index, ""),
					Raw:    lo.NthOr(r, column.Index, ""),
				})
			}
			sheet.Rows = append(sheet.Rows, row)
		}
	}
	return sheet
}

func parseValue(columnType, value string) (any, error) {
	if value == "" {
		switch columnType {
		case "string":
			return "", nil
		case "int":
			return 0, nil
		case "float":
			return float64(0), nil
		case "bool":
			return false, nil
		case "datetime":
			return time.Time{}, nil
		case "date":
			return time.Time{}.Format(time.DateOnly), nil
		case "unixtime":
			return int64(0), nil
		case "ref":
			return "", nil
		case "":
			return "", nil
		}
		return nil, fmt.Errorf("unmatched type:%s", columnType)
	}
	switch columnType {
	case "string":
		return value, nil
	case "int":
		return strconv.Atoi(value)
	case "float":
		return strconv.ParseFloat(value, 64)
	case "bool":
		return strconv.ParseBool(value)
	case "datetime":
		return time.Parse(time.RFC3339, value)
	case "date":
		t, err := time.ParseInLocation(time.DateOnly, value, time.UTC)
		if err != nil {
			return nil, err
		}
		return t.Format(time.DateOnly), nil
	case "unixtime":
		t, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return nil, err
		}
		return t.Unix(), nil
	case "ref":
		return value, nil
	case "":
		return value, nil
	}
	return nil, fmt.Errorf("unmatched type:%s", columnType)
}
