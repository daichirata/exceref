package errs_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/daichirata/exceref/internal/errs"
)

func TestWrap_Nil(t *testing.T) {
	t.Parallel()

	require.NoError(t, errs.Wrap(nil, "noop"))
}

func TestWrap_WithCallerAndUnwrap(t *testing.T) {
	t.Parallel()

	base := errors.New("base error")
	err := errs.Wrap(base, "test op")

	var wrapped *errs.Error
	require.ErrorAs(t, err, &wrapped)
	require.Equal(t, "test op", wrapped.Op)
	require.Equal(t, "errs_test.go", wrapped.File)
	require.Greater(t, wrapped.Line, 0)
	require.Equal(t, base, errors.Unwrap(err))
	require.True(t, strings.Contains(err.Error(), "test op (errs_test.go:"))
}
