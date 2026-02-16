package exceref

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildGenerator(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		lang        string
		generatorTy any
	}{
		{name: "go", lang: "go", generatorTy: &goGenerator{}},
		{name: "csharp", lang: "csharp", generatorTy: &csharpGenerator{}},
		{name: "default", lang: "unknown", generatorTy: &generator{}},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g := BuildGenerator(tc.lang, GenerateOption{})
			require.IsType(t, tc.generatorTy, g)
		})
	}
}

func TestGoGenerator_Generate(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	templatePath := filepath.Join(dir, "model.tmpl")
	require.NoError(t, os.WriteFile(templatePath, []byte(`package models

type {{ .Name }} struct {
{{- range .Fields }}
	{{ .Name }} {{ .Type }}
{{- end }}
}
`), 0644))

	g := NewGoGenerator(GenerateOption{
		OutDir:       dir,
		Prefix:       "App",
		TemplatePath: templatePath,
	})
	sheet := &Sheet{
		Name: "users",
		Columns: []*Column{
			{Name: "id", Type: ColumnTypeInt, Index: 0},
			{Name: "created_at", Type: ColumnTypeDatetime, Index: 1},
			{Name: "", Type: ColumnTypeString, Index: 2},
		},
	}

	require.NoError(t, g.Generate(sheet))

	outPath := filepath.Join(dir, "Appusers.gen.go")
	body, err := os.ReadFile(outPath)
	require.NoError(t, err)
	content := string(body)
	require.Contains(t, content, "type Appuser struct")
	require.Contains(t, content, "ID        int64")
	require.Contains(t, content, "CreatedAt time.Time")
}
