package exceref

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
	"gopkg.in/yaml.v3"
)

func TestMetadataExporter_Export(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	bookPath := filepath.Join(dir, "book.xlsx")
	buildMetadataTestBook(t, bookPath)

	file, err := Open(bookPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, file.Close())
	})

	outDir := filepath.Join(dir, "out")
	require.NoError(t, os.MkdirAll(outDir, 0755))
	require.NoError(t, file.ExportMetadata(outDir))

	// _references metadata
	referencesBody, err := os.ReadFile(filepath.Join(outDir, "book_references.yaml"))
	require.NoError(t, err)

	var references MetadataReferencesYAML
	require.NoError(t, yaml.Unmarshal(referencesBody, &references))
	require.Len(t, references.References, 1)
	require.Equal(t, "Items", references.References[0].Sheet)
	require.Equal(t, "status", references.References[0].Column)
	require.Equal(t, "master.xlsx", references.References[0].ReferenceFile)

	// data sheet metadata
	itemsBody, err := os.ReadFile(filepath.Join(outDir, "Items.yaml"))
	require.NoError(t, err)

	var items MetadataDataYAML
	require.NoError(t, yaml.Unmarshal(itemsBody, &items))
	require.Equal(t, "Items", items.Sheet)
	require.Len(t, items.Schema, 2)
	require.Equal(t, "id", items.Schema[0].Name)
	require.Equal(t, ColumnTypeString, items.Schema[0].Type)
	require.Nil(t, items.Schema[0].Ref)
	require.Equal(t, "status", items.Schema[1].Name)
	require.NotNil(t, items.Schema[1].Ref)
	require.Equal(t, "master.xlsx", items.Schema[1].Ref.File)
	require.Equal(t, "Master", items.Schema[1].Ref.Sheet)
	require.Equal(t, "code", items.Schema[1].Ref.Key)
	require.Equal(t, "label", items.Schema[1].Ref.Value)

	// _types sheet metadata is renamed to <book>_types
	typesBody, err := os.ReadFile(filepath.Join(outDir, "book_types.yaml"))
	require.NoError(t, err)

	var types MetadataDataYAML
	require.NoError(t, yaml.Unmarshal(typesBody, &types))
	require.Equal(t, "book_types", types.Sheet)
	require.Len(t, types.Schema, 1)
	require.Equal(t, "kind", types.Schema[0].Name)
}

func buildMetadataTestBook(t *testing.T, path string) {
	t.Helper()

	f := excelize.NewFile()
	require.NoError(t, f.SetSheetName("Sheet1", "Items"))
	require.NoError(t, f.SetSheetRow("Items", "A1", &[]any{"string", "ref"}))
	require.NoError(t, f.SetSheetRow("Items", "A2", &[]any{"id", "status"}))
	require.NoError(t, f.SetSheetRow("Items", "A3", &[]any{"ID", "Status"}))
	require.NoError(t, f.SetSheetRow("Items", "A4", &[]any{"1", "A"}))

	_, err := f.NewSheet(ReferenceDefinitionSheetName)
	require.NoError(t, err)
	require.NoError(t, f.SetSheetRow(
		ReferenceDefinitionSheetName,
		"A1",
		&[]any{"sheet", "column", "reference_file", "reference_sheet", "reference_key", "reference_value", "reference_name"},
	))
	require.NoError(t, f.SetSheetRow(
		ReferenceDefinitionSheetName,
		"A2",
		&[]any{"Items", "status", "master.xlsx", "Master", "code", "label", "StatusMaster"},
	))

	_, err = f.NewSheet("_types")
	require.NoError(t, err)
	require.NoError(t, f.SetSheetRow("_types", "A1", &[]any{"string"}))
	require.NoError(t, f.SetSheetRow("_types", "A2", &[]any{"kind"}))
	require.NoError(t, f.SetSheetRow("_types", "A3", &[]any{"Kind"}))

	require.NoError(t, f.SaveAs(path))
	require.NoError(t, f.Close())
}

