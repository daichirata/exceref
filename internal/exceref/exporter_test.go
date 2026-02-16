package exceref

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildExporter(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		format     string
		exporterTy any
	}{
		{
			name:       "json",
			format:     "json",
			exporterTy: &jsonExporter{},
		},
		{
			name:       "yaml",
			format:     "yaml",
			exporterTy: &yamlExporter{},
		},
		{
			name:       "default to csv",
			format:     "unknown",
			exporterTy: &csvExporter{},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			exporter := BuildExporter(tc.format, t.TempDir(), "prefix_")
			require.IsType(t, tc.exporterTy, exporter)
		})
	}
}

func TestCSVExporter_Export(t *testing.T) {
	t.Parallel()

	outDir := t.TempDir()
	exporter := NewCSVExporter(outDir, "p_")
	columns := []*Column{
		{Name: "id", Type: ColumnTypeInt, Index: 0},
		{Name: "name", Type: ColumnTypeString, Index: 1},
		{Name: "", Type: ColumnTypeString, Index: 2},
	}
	sheet := &Sheet{
		Name:    "Users",
		Columns: columns,
		Rows: []Row{
			{
				{Column: columns[0], Value: 1},
				{Column: columns[1], Value: "Alice"},
				{Column: columns[2], Value: "ignored"},
			},
		},
	}

	require.NoError(t, exporter.Export(sheet))

	path := filepath.Join(outDir, "p_Users.csv")
	body, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "id,name\n1,Alice\n", string(body))
}
