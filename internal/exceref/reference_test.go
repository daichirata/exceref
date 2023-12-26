package exceref_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/daichirata/exceref/internal/exceref"
)

func TestReferenceDefinition_ReferenceFileName(t *testing.T) {
	def1 := exceref.ReferenceDefinition{ReferenceFile: "Book1.xlsx"}
	require.Equal(t, "Book1.xlsx", def1.ReferenceFileName())

	def2 := exceref.ReferenceDefinition{ReferenceFile: "Book1"}
	require.Equal(t, "Book1.xlsx", def2.ReferenceFileName())
}

func TestReferenceDefinition_ReferenceFilePath(t *testing.T) {
	def := exceref.ReferenceDefinition{ReferenceFile: "Book1.xlsx", BaseDir: "/tmp/exceref"}
	require.Equal(t, "/tmp/exceref/Book1.xlsx", def.ReferenceFilePath())
}
