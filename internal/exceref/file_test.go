package exceref_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"

	"github.com/daichirata/exceref/internal/exceref"
)

func TestFile_UpdateReferenceData_EmptyReferenceSource(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	referencePath := filepath.Join(dir, "master.xlsx")
	targetPath := filepath.Join(dir, "target.xlsx")

	referenceBook := excelize.NewFile()
	require.NoError(t, referenceBook.SetSheetName("Sheet1", "master"))
	require.NoError(t, referenceBook.SetSheetRow("master", "A1", &[]any{"string", "string"}))
	require.NoError(t, referenceBook.SetSheetRow("master", "A2", &[]any{"code", "label"}))
	require.NoError(t, referenceBook.SetSheetRow("master", "A3", &[]any{"", ""}))
	require.NoError(t, referenceBook.SaveAs(referencePath))
	require.NoError(t, referenceBook.Close())

	targetBook := excelize.NewFile()
	require.NoError(t, targetBook.SetSheetName("Sheet1", exceref.ReferenceDefinitionSheetName))
	require.NoError(t, targetBook.SetSheetRow(
		exceref.ReferenceDefinitionSheetName,
		"A1",
		&[]any{"sheet", "column", "reference_file", "reference_sheet", "reference_key", "reference_value", "reference_name"},
	))
	require.NoError(t, targetBook.SetSheetRow(
		exceref.ReferenceDefinitionSheetName,
		"A2",
		&[]any{"", "", "master.xlsx", "master", "code", "label", "MasterCodes"},
	))
	require.NoError(t, targetBook.SaveAs(targetPath))
	require.NoError(t, targetBook.Close())

	file, err := exceref.Open(targetPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, file.Close())
	})

	require.NoError(t, file.UpdateReferenceData())
}
