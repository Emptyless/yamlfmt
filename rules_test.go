package yamlfmt

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultOpenAPIRules(t *testing.T) {
	t.Parallel()
	// Arrange
	actual, err := os.ReadFile("testdata/simple/openapi.yaml")
	require.NoError(t, err)

	// Act
	b, err := LintBytes(actual, DefaultOpenAPIRules())

	// Assert
	require.NoError(t, err)
	expected, err := os.ReadFile("testdata/simple/openapi.fmt.yaml")
	require.NoError(t, err)
	require.Equal(t, string(expected), string(b))
}
