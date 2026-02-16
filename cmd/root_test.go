package cmd

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	apperrs "github.com/daichirata/exceref/internal/errs"
)

func TestPrintErrorChain(t *testing.T) {
	t.Parallel()

	base := errors.New("base failure")
	err := apperrs.Wrap(apperrs.Wrap(base, "inner op"), "outer op")

	prevStderr := os.Stderr
	r, w, pipeErr := os.Pipe()
	require.NoError(t, pipeErr)
	os.Stderr = w
	t.Cleanup(func() {
		os.Stderr = prevStderr
		_ = r.Close()
	})

	printErrorChain(err)

	require.NoError(t, w.Close())
	out, readErr := io.ReadAll(r)
	require.NoError(t, readErr)

	s := string(out)
	require.Contains(t, s, "Error: outer op (")
	require.Contains(t, s, "Caused by: inner op (")
	require.Contains(t, s, "Caused by: base failure")
}
