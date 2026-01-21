package exceref_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/daichirata/exceref/internal/exceref"
)

var columns = []*exceref.Column{
	{Name: "column_a", Type: "int", Index: 0},
	{Name: "column_b", Type: "bool", Index: 1},
	{Name: "column_c", Type: "string", Index: 2},
	{Name: "", Type: "string", Index: 3},
	{Name: "column_x", Type: "", Index: 4},
}

var sheet = &exceref.Sheet{
	Columns: columns,
	Rows: []exceref.Row{
		{
			{Column: columns[0], Value: 1},
			{Column: columns[1], Value: true},
			{Column: columns[2], Value: "one"},
			{Column: columns[3], Value: "skip_empty_name"},
			{Column: columns[4], Value: "skip_empty_type"},
		},
		{
			{Column: columns[0], Value: 2},
			{Column: columns[1], Value: false},
			{Column: columns[2], Value: "three"},
			{Column: columns[3], Value: "skip_empty_name"},
			{Column: columns[4], Value: "skip_empty_type"},
		},
	},
}

func TestRow_Cells(t *testing.T) {
	row := sheet.Rows[0]

	cells := row.Cells("column_a", "column_b")
	require.Equal(t, []*exceref.Cell{row[0], row[1]}, cells)

	cells = row.Cells("column_a", "column_d")
	require.Equal(t, []*exceref.Cell{row[0]}, cells)

	cells = row.Cells("column_d")
	require.Equal(t, []*exceref.Cell{}, cells)
}

func TestSheet_Column(t *testing.T) {
	column, err := sheet.Column("column_a")
	require.NoError(t, err)
	require.Equal(t, sheet.Columns[0], column)

	_, err = sheet.Column("column_d")
	require.Error(t, err)
}

func TestColumn_IsExportable(t *testing.T) {
	require.True(t, (&exceref.Column{Name: "col", Type: "int"}).IsExportable())
	require.False(t, (&exceref.Column{Name: "", Type: "int"}).IsExportable())
	require.False(t, (&exceref.Column{Name: "col", Type: ""}).IsExportable())
}

func TestSheet_Map(t *testing.T) {
	require.Equal(t, []map[string]any{
		{"column_a": 1, "column_b": true, "column_c": "one"},
		{"column_a": 2, "column_b": false, "column_c": "three"},
	}, sheet.Map())
}

func TestSheet_Sqrefs(t *testing.T) {
	referenceDefinition := &exceref.ReferenceDefinition{
		Index:  0,
		Column: "column_b",
	}
	dst, src, err := sheet.Sqrefs(referenceDefinition)
	require.NoError(t, err)
	require.Equal(t, "B4:B9999", dst)
	require.Equal(t, "$A$1:$A$9999", src)
}

func TestNewDataSeet(t *testing.T) {
	rows := [][]string{
		{"string", "int", "", "float", "bool", "datetime", "date", "unixtime", "ref"},
		{"column_a", "column_b", "", "column_c", "column_d", "column_e", "column_f", "column_g", "column_h"},
		{"desc_a", "desc_b", "", "desc_c", "desc_d", "desc_e", "desc_f", "desc_g", "desc_h"},
		{"a", "1", "foo", "0.1", "true", "2024-01-01T00:00:00Z", "2024-01-01", "2024-01-01T00:00:00Z", "ref_a"},
		{"b", "2", "bar", "0.2", "false", "2024-01-02T00:00:00Z", "2024-01-02", "2024-01-02T00:00:00Z", "ref_b"},
		{"c", "3", "buzz", "0.3", "true", "2024-01-03T00:00:00Z", "2024-01-03", "2024-01-03T00:00:00Z", "ref_a"},
		{"", "", "", "", "", "", "", "", ""},
	}
	sheet, err := exceref.NewDataSeet("test_sheet", rows)
	require.NoError(t, err)
	require.Equal(t, sheet.Name, "test_sheet")
	require.Equal(t, []exceref.Row{
		{
			{Column: sheet.Columns[0], Value: "a", Raw: "a"},
			{Column: sheet.Columns[1], Value: 1, Raw: "1"},
			{Column: sheet.Columns[2], Value: "", Raw: "foo"},
			{Column: sheet.Columns[3], Value: 0.1, Raw: "0.1"},
			{Column: sheet.Columns[4], Value: true, Raw: "true"},
			{Column: sheet.Columns[5], Value: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Raw: "2024-01-01T00:00:00Z"},
			{Column: sheet.Columns[6], Value: "2024-01-01", Raw: "2024-01-01"},
			{Column: sheet.Columns[7], Value: int64(1704067200), Raw: "2024-01-01T00:00:00Z"},
			{Column: sheet.Columns[8], Value: "ref_a", Raw: "ref_a"},
		},
		{
			{Column: sheet.Columns[0], Value: "b", Raw: "b"},
			{Column: sheet.Columns[1], Value: 2, Raw: "2"},
			{Column: sheet.Columns[2], Value: "", Raw: "bar"},
			{Column: sheet.Columns[3], Value: 0.2, Raw: "0.2"},
			{Column: sheet.Columns[4], Value: false, Raw: "false"},
			{Column: sheet.Columns[5], Value: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), Raw: "2024-01-02T00:00:00Z"},
			{Column: sheet.Columns[6], Value: "2024-01-02", Raw: "2024-01-02"},
			{Column: sheet.Columns[7], Value: int64(1704153600), Raw: "2024-01-02T00:00:00Z"},
			{Column: sheet.Columns[8], Value: "ref_b", Raw: "ref_b"},
		},
		{
			{Column: sheet.Columns[0], Value: "c", Raw: "c"},
			{Column: sheet.Columns[1], Value: 3, Raw: "3"},
			{Column: sheet.Columns[2], Value: "", Raw: "buzz"},
			{Column: sheet.Columns[3], Value: 0.3, Raw: "0.3"},
			{Column: sheet.Columns[4], Value: true, Raw: "true"},
			{Column: sheet.Columns[5], Value: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), Raw: "2024-01-03T00:00:00Z"},
			{Column: sheet.Columns[6], Value: "2024-01-03", Raw: "2024-01-03"},
			{Column: sheet.Columns[7], Value: int64(1704240000), Raw: "2024-01-03T00:00:00Z"},
			{Column: sheet.Columns[8], Value: "ref_a", Raw: "ref_a"},
		},
		{
			{Column: sheet.Columns[0], Value: "", Raw: ""},
			{Column: sheet.Columns[1], Value: 0, Raw: ""},
			{Column: sheet.Columns[2], Value: "", Raw: ""},
			{Column: sheet.Columns[3], Value: float64(0), Raw: ""},
			{Column: sheet.Columns[4], Value: false, Raw: ""},
			{Column: sheet.Columns[5], Value: time.Time{}, Raw: ""},
			{Column: sheet.Columns[6], Value: "0001-01-01", Raw: ""},
			{Column: sheet.Columns[7], Value: int64(0), Raw: ""},
			{Column: sheet.Columns[8], Value: "", Raw: ""},
		},
	}, sheet.Rows)

	sheet, err = exceref.NewDataSeet("test_sheet", rows[:exceref.DataSheetIndexBody])
	require.NoError(t, err)
	require.Len(t, sheet.Rows, 0)
}

func TestNewReferenceDefinitionSheet(t *testing.T) {
	rows := [][]string{
		{"sheet", "column", "reference_file", "reference_sheet", "reference_name", "reference_value"},
		{"Book1_sheet", "column_a", "Book2.xlsx", "Book2_sheet1", "ref_column_a", "ref_column_b"},
		{"Book1_sheet", "column_b", "Book2.xlsx", "Book2_sheet2", "ref_column_c", "ref_column_d"},
	}
	sheet := exceref.NewReferenceDefinitionSheet("test_sheet", rows)
	require.Equal(t, sheet.Name, "test_sheet")
	require.Equal(t, []exceref.Row{
		{
			{Column: sheet.Columns[0], Value: "Book1_sheet", Raw: "Book1_sheet"},
			{Column: sheet.Columns[1], Value: "column_a", Raw: "column_a"},
			{Column: sheet.Columns[2], Value: "Book2.xlsx", Raw: "Book2.xlsx"},
			{Column: sheet.Columns[3], Value: "Book2_sheet1", Raw: "Book2_sheet1"},
			{Column: sheet.Columns[4], Value: "ref_column_a", Raw: "ref_column_a"},
			{Column: sheet.Columns[5], Value: "ref_column_b", Raw: "ref_column_b"},
		},
		{
			{Column: sheet.Columns[0], Value: "Book1_sheet", Raw: "Book1_sheet"},
			{Column: sheet.Columns[1], Value: "column_b", Raw: "column_b"},
			{Column: sheet.Columns[2], Value: "Book2.xlsx", Raw: "Book2.xlsx"},
			{Column: sheet.Columns[3], Value: "Book2_sheet2", Raw: "Book2_sheet2"},
			{Column: sheet.Columns[4], Value: "ref_column_c", Raw: "ref_column_c"},
			{Column: sheet.Columns[5], Value: "ref_column_d", Raw: "ref_column_d"},
		},
	}, sheet.Rows)
	require.Len(t, sheet.Rows, 2)

	sheet = exceref.NewReferenceDefinitionSheet("test_sheet", rows[:exceref.ReferenceDefinitionSheetIndexBody])
	require.Len(t, sheet.Rows, 0)
}
