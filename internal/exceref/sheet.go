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

type ColumnType string

const (
	ColumnTypeString   ColumnType = "string"
	ColumnTypeFloat    ColumnType = "float"
	ColumnTypeInt      ColumnType = "int"
	ColumnTypeBool     ColumnType = "bool"
	ColumnTypeDatetime ColumnType = "datetime"
	ColumnTypeDate     ColumnType = "date"
	ColumnTypeUnixtime ColumnType = "unixtime"
	ColumnTypeRef      ColumnType = "ref"
)

func (c ColumnType) String() string {
	return string(c)
}

func NewColumnType(s string) (ColumnType, error) {
	switch ColumnType(s) {
	case ColumnTypeString, ColumnTypeFloat, ColumnTypeInt, ColumnTypeBool,
		ColumnTypeDatetime, ColumnTypeDate, ColumnTypeUnixtime, ColumnTypeRef:
		return ColumnType(s), nil
	case "":
		return "", nil
	default:
		return "", fmt.Errorf("unknown column type: %s", s)
	}
}

type Column struct {
	Name        string
	Type        ColumnType
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
	first, err := excelize.CoordinatesToCellName(column.Index+1, DataSheetIndexBody+1)
	if err != nil {
		return "", "", err
	}
	last, err := excelize.CoordinatesToCellName(column.Index+1, 9999)
	if err != nil {
		return "", "", err
	}
	sqref := fmt.Sprintf("%s:%s", first, last)

	srcFirst, err := excelize.CoordinatesToCellName(referenceDefinition.Index+1, 1, true)
	if err != nil {
		return "", "", err
	}
	srcLast, err := excelize.CoordinatesToCellName(referenceDefinition.Index+1, 9999, true)
	if err != nil {
		return "", "", err
	}
	srcSqref := fmt.Sprintf("%s:%s", srcFirst, srcLast)

	return sqref, srcSqref, nil
}

func NewDataSeet(name string, rows [][]string) (*Sheet, error) {
	sheet := &Sheet{
		Name: name,
	}
	for i, r := range rows {
		switch i {
		case DataSheetIndexColumnType:
			for j, value := range r {
				columnType, err := NewColumnType(value)
				if err != nil {
					return nil, err
				}
				sheet.Columns = append(sheet.Columns, &Column{
					Type:  columnType,
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

func parseValue(columnType ColumnType, value string) (any, error) {
	if columnType == "" {
		return "", nil
	}
	if value == "" {
		switch columnType {
		case ColumnTypeString:
			return "", nil
		case ColumnTypeInt:
			return 0, nil
		case ColumnTypeFloat:
			return float64(0), nil
		case ColumnTypeBool:
			return false, nil
		case ColumnTypeDatetime:
			return time.Time{}, nil
		case ColumnTypeDate:
			return time.Time{}.Format(time.DateOnly), nil
		case ColumnTypeUnixtime:
			return int64(0), nil
		case ColumnTypeRef:
			return "", nil
		}
		return nil, fmt.Errorf("unmatched type:%s", columnType)
	}
	switch columnType {
	case ColumnTypeString:
		return value, nil
	case ColumnTypeInt:
		return strconv.Atoi(value)
	case ColumnTypeFloat:
		return strconv.ParseFloat(value, 64)
	case ColumnTypeBool:
		return strconv.ParseBool(value)
	case ColumnTypeDatetime:
		return time.Parse(time.RFC3339, value)
	case ColumnTypeDate:
		t, err := time.ParseInLocation(time.DateOnly, value, time.UTC)
		if err != nil {
			return nil, err
		}
		return t.Format(time.DateOnly), nil
	case ColumnTypeUnixtime:
		t, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return nil, err
		}
		return t.Unix(), nil
	case ColumnTypeRef:
		return value, nil
	}
	return nil, fmt.Errorf("unmatched type:%s", columnType)
}
