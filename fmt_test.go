package yamlfmt

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestLintBytes_NoBytes(t *testing.T) {
	t.Parallel()
	// Arrange
	var b []byte

	// Act
	actual, err := LintBytes(b, []Rule{NewRule("$", StringOrderingFn)})

	// Assert
	require.NoError(t, err)
	assert.Equal(t, b, actual)
}

func TestLint_DocumentWithoutContent(t *testing.T) {
	t.Parallel()
	// Arrange
	expected := &yaml.Node{
		Kind: yaml.DocumentNode,
	}

	actual := &yaml.Node{
		Kind: yaml.DocumentNode,
	}

	// Act
	Lint(actual, []Rule{NewRule("$", StringOrderingFn)})

	// Assert
	assert.True(t, reflect.DeepEqual(actual, expected))
}

func TestLint_DocumentNode(t *testing.T) {
	t.Parallel()
	// Arrange
	node := &yaml.Node{
		Kind: yaml.DocumentNode,
		Content: []*yaml.Node{
			{
				Kind: yaml.MappingNode,
				Content: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "foo"},
					{Kind: yaml.ScalarNode, Value: "bar"},
					{Kind: yaml.ScalarNode, Value: "test"},
					{Kind: yaml.ScalarNode, Value: "value"},
					{Kind: yaml.ScalarNode, Value: "abc"},
					{Kind: yaml.ScalarNode, Value: "def"},
				},
			},
		},
	}

	// Act
	// first sort alphabetically and then pull 'test' (if exists) to front
	Lint(node, []Rule{NewRule("$", StringOrderingFn, NewSimpleOrdering("test"))})

	// Assert
	assert.Equal(t, "test", node.Content[0].Content[0].Value)
	assert.Equal(t, "value", node.Content[0].Content[1].Value)
	assert.Equal(t, "abc", node.Content[0].Content[2].Value)
	assert.Equal(t, "def", node.Content[0].Content[3].Value)
	assert.Equal(t, "foo", node.Content[0].Content[4].Value)
	assert.Equal(t, "bar", node.Content[0].Content[5].Value)
}

func TestLint_Nil(t *testing.T) {
	t.Parallel()
	// Arrange
	var node *yaml.Node

	// Act
	Lint(node, []Rule{})

	// Assert
	assert.Nil(t, node)
}

func TestValidate_Errors(t *testing.T) {
	t.Parallel()
	// Arrange
	rule1 := NewRule("$[[")

	// Act
	err := Validate([]Rule{rule1})

	// Assert
	require.EqualError(t, err, "invalid path \"$[[\": char \"[\": illegal token")
}

func TestValidate_NoErrors(t *testing.T) {
	t.Parallel()
	// Arrange
	rule1 := NewRule("$.key")

	// Act
	err := Validate([]Rule{rule1})

	// Assert
	assert.NoError(t, err)
}

func TestValidate_NoRules(t *testing.T) {
	t.Parallel()
	// Act
	err := Validate(nil)

	// Assert
	require.NoError(t, err)
}

func TestContains_OnMatch(t *testing.T) {
	t.Parallel()
	// Arrange
	path := "$.some.key"
	rule := NewRule("$.some.key")

	// Act
	ok := rule.contains(path)

	// Assert
	assert.True(t, ok)
}

func TestContains_BeforeMatch(t *testing.T) {
	t.Parallel()
	// Arrange
	path := "$.some"
	rule := NewRule("$.some.key")

	// Act
	ok := rule.contains(path)

	// Assert
	assert.True(t, ok)
}

func TestContains_AbsoluteToDeep(t *testing.T) {
	t.Parallel()
	// Arrange
	path := "$.some.key.to.deep"
	rule := NewRule("$.some.key")

	// Act
	ok := rule.contains(path)

	// Assert
	assert.False(t, ok)
}

func TestContains_PathError(t *testing.T) {
	t.Parallel()
	// Arrange
	path := "$[["

	// Act
	fn := func() {
		rule := NewRule("$.key")
		rule.contains(path)
	}

	// Assert
	assert.Panics(t, fn)
}

func TestContains_RulePathError(t *testing.T) {
	t.Parallel()
	// Arrange
	path := "$.key"

	// Act
	fn := func() {
		rule := NewRule("$[[")
		rule.contains(path)
	}

	// Assert
	assert.Panics(t, fn)
}

func TestContains_RelativeAlwaysTrue(t *testing.T) {
	t.Parallel()
	// Arrange
	path := "$.anything" // in theory there can be any path below $.anything like $.anything.name which would match
	rule := NewRule(".name")

	// Act
	ok := rule.contains(path)

	// Assert
	assert.True(t, ok)
}

func TestMatch_Absolute_MatchesPath(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		Rule Rule
		Path string
	}{
		"one key": {
			Rule: NewRule("$.key"),
			Path: "$.key",
		},
		"two keys": {
			Rule: NewRule("$.some.key"),
			Path: "$.some.key",
		},
		"all key": {
			Rule: NewRule("$.some.*"),
			Path: "$.some.key",
		},
		"all with nesting": {
			Rule: NewRule("$.with.*.name"),
			Path: "$.with.some.name",
		},
		"index key": {
			Rule: NewRule("$[0]"),
			Path: "$[0]",
		},
		"index key with nesting": {
			Rule: NewRule("$.name[1]"),
			Path: "$.name[1]",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Act
			ok := test.Rule.match(test.Path)

			// Assert
			assert.True(t, ok)
		})
	}
}

func TestMatch_Relative_MatchesPath(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		Rule Rule
		Path string
	}{
		"one key": {
			Rule: NewRule(".key"),
			Path: "$.some.key",
		},
		"two keys": {
			Rule: NewRule(".some.key"),
			Path: "$[2].with.some.key",
		},
		"all key": {
			Rule: NewRule(".some.*"),
			Path: "$[3].with.some.key",
		},
		"all with nesting": {
			Rule: NewRule(".with.*.name"),
			Path: "$[1].with.some.name",
		},
		"index key": {
			Rule: NewRule("[0]"),
			Path: "$[0]",
		},
		"index key with nesting": {
			Rule: NewRule(".name[1]"),
			Path: "$.name[1]",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Act
			ok := test.Rule.match(test.Path)

			// Assert
			assert.True(t, ok)
		})
	}
}

func TestMatch_AbsolutePathDifferentParts(t *testing.T) {
	t.Parallel()
	// Arrange
	path := "$.key"
	rule := NewRule("$.key.value")

	// Act
	ok := rule.match(path)

	// Assert
	assert.False(t, ok)
}

func TestMatch_RulePathError(t *testing.T) {
	t.Parallel()
	// Arrange
	path := "$.key"

	// Act
	fn := func() {
		rule := NewRule("$[[")
		rule.match(path)
	}

	// Assert
	assert.Panics(t, fn)
}

func TestMatch_PathError(t *testing.T) {
	t.Parallel()
	// Arrange
	path := "$[["

	// Act
	fn := func() {
		rule := NewRule("$.key")
		rule.match(path)
	}

	// Assert
	assert.Panics(t, fn)
}

func TestNewSimpleOrdering(t *testing.T) {
	t.Parallel()
	// Arrange
	node := &yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "c"},
			{Kind: yaml.ScalarNode, Value: "c-value"},
			{Kind: yaml.ScalarNode, Value: "f"},
			{Kind: yaml.ScalarNode, Value: "f-value"},
			{Kind: yaml.ScalarNode, Value: "b"},
			{Kind: yaml.ScalarNode, Value: "b-value"},
			{Kind: yaml.ScalarNode, Value: "e"},
			{Kind: yaml.ScalarNode, Value: "e-value"},
			{Kind: yaml.ScalarNode, Value: "a"},
			{Kind: yaml.ScalarNode, Value: "a-value"},
		},
	}

	// Act
	NewSimpleOrdering("a", "b", "c")("", node)

	// Assert
	assert.Equal(t, "a", node.Content[0].Value)
	assert.Equal(t, "a-value", node.Content[1].Value)
	assert.Equal(t, "b", node.Content[2].Value)
	assert.Equal(t, "b-value", node.Content[3].Value)
	assert.Equal(t, "c", node.Content[4].Value)
	assert.Equal(t, "c-value", node.Content[5].Value)
	assert.Equal(t, "f", node.Content[6].Value)
	assert.Equal(t, "f-value", node.Content[7].Value)
	assert.Equal(t, "e", node.Content[8].Value)
	assert.Equal(t, "e-value", node.Content[9].Value)
}

func TestNewSimpleOrdering_IrrelevantValue(t *testing.T) {
	t.Parallel()
	keys := map[string][]string{
		"nil keys": nil,
		"0 keys":   {},
		">1 keys":  {"a", "b", "c"},
	}

	tests := map[string]struct {
		// Node as a func such that the original value is reproducible
		Node func() *yaml.Node
	}{
		"nil node": {
			Node: func() *yaml.Node { return nil },
		},
		"node without content": {
			Node: func() *yaml.Node {
				return &yaml.Node{Kind: yaml.MappingNode}
			},
		},
		"not a mapping or sequence node": {
			Node: func() *yaml.Node {
				return &yaml.Node{Kind: yaml.ScalarNode, Content: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "test"},
				}}
			},
		},
	}
	for testName, testData := range tests {
		for keysName, keysData := range keys {
			t.Run(fmt.Sprintf("%s - %s", testName, keysName), func(t *testing.T) {
				t.Parallel()
				// Arrange
				expected := testData.Node()

				// Act
				actual := testData.Node()
				NewSimpleOrdering(keysData...)("", actual)

				// Assert
				assert.True(t, reflect.DeepEqual(expected, actual))
			})
		}
	}
}

func TestStringOrderingFn_SortsMap(t *testing.T) {
	t.Parallel()
	// Arrange
	node := &yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "ckey"},
			{Kind: yaml.ScalarNode, Value: "cvalue"},
			{Kind: yaml.ScalarNode, Value: "bkey"},
			{Kind: yaml.ScalarNode, Value: "bvalue"},
			{Kind: yaml.ScalarNode, Value: "akey"},
			{Kind: yaml.ScalarNode, Value: "avalue"},
		},
	}

	// Act
	StringOrderingFn("", node)

	// Assert
	assert.Equal(t, "akey", node.Content[0].Value)
	assert.Equal(t, "avalue", node.Content[1].Value)
	assert.Equal(t, "bkey", node.Content[2].Value)
	assert.Equal(t, "bvalue", node.Content[3].Value)
	assert.Equal(t, "ckey", node.Content[4].Value)
	assert.Equal(t, "cvalue", node.Content[5].Value)
}

func TestStringOrderingFn_SortsSequence(t *testing.T) {
	t.Parallel()
	// Arrange
	node := &yaml.Node{
		Kind: yaml.SequenceNode,
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "c"},
			{Kind: yaml.ScalarNode, Value: "b"},
			{Kind: yaml.ScalarNode, Value: "a"},
		},
	}

	// Act
	StringOrderingFn("", node)

	// Assert
	assert.Equal(t, "a", node.Content[0].Value)
	assert.Equal(t, "b", node.Content[1].Value)
	assert.Equal(t, "c", node.Content[2].Value)
}

func TestStringOrderingFn_UnprocessableNode(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		// Node as a func such that the original value is reproducible
		Node func() *yaml.Node
	}{
		"nil node": {
			Node: func() *yaml.Node { return nil },
		},
		"node without content": {
			Node: func() *yaml.Node {
				return &yaml.Node{Kind: yaml.MappingNode}
			},
		},
		"not a mapping or sequence node": {
			Node: func() *yaml.Node {
				return &yaml.Node{Kind: yaml.ScalarNode, Content: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "test"},
				}}
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Arrange
			expected := test.Node()

			// Act
			actual := test.Node()
			StringOrderingFn("", test.Node())

			// Assert
			assert.True(t, reflect.DeepEqual(expected, actual))
		})
	}
}

func TestParts_ParsesComplexPath(t *testing.T) {
	t.Parallel()
	// Arrange
	path := "$.key[0][*].name[1].ending"

	// Act
	p, err := parts(path)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []string{"$", ".key", "[0]", "[*]", ".name", "[1]", ".ending"}, p)
}

func TestParts_ParsesRelativeKey(t *testing.T) {
	t.Parallel()
	// Arrange
	path := ".name[*]"

	// Act
	p, err := parts(path)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []string{".name", "[*]"}, p)
}

func TestParts_ParsesRelativeIndex(t *testing.T) {
	t.Parallel()
	// Arrange
	path := "[1]"

	// Act
	p, err := parts(path)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []string{"[1]"}, p)
}

func TestParts_DisallowedToken(t *testing.T) {
	t.Parallel()
	// Arrange
	path := "$[["

	// Act
	p, err := parts(path)

	// Assert
	require.EqualError(t, err, "invalid path \"$[[\": char \"[\": illegal token")
	require.Nil(t, p)
}

func TestParts_EmptyPath(t *testing.T) {
	t.Parallel()
	// Arrange
	path := ""

	// Act
	p, err := parts(path)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, p)
}

func TestNext_NilNode(t *testing.T) {
	t.Parallel()
	// Arrange
	var node *yaml.Node

	// Act
	n := next("", node)

	// Assert
	assert.Empty(t, n)
}

func TestNext_MappingNode(t *testing.T) {
	t.Parallel()
	// Arrange
	node := &yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			{
				Kind:  yaml.ScalarNode,
				Value: "key",
			},
			{
				Kind:  yaml.ScalarNode,
				Value: "content",
			},
		},
	}

	// Act
	n := next("$", node)

	// Assert
	require.NotEmpty(t, n)
	require.Contains(t, n, "$.key")
	assert.Equal(t, n["$.key"], node.Content[1])
}

func TestNext_SequenceNode(t *testing.T) {
	t.Parallel()
	// Arrange
	node := &yaml.Node{
		Kind: yaml.SequenceNode,
		Content: []*yaml.Node{
			{
				Kind:  yaml.ScalarNode,
				Value: "value1",
			},
			{
				Kind:  yaml.ScalarNode,
				Value: "value2",
			},
		},
	}

	// Act
	n := next("$", node)
	require.Len(t, n, 2)
	require.Contains(t, n, "$[0]")
	assert.Equal(t, n["$[0]"], node.Content[0])
	require.Contains(t, n, "$[1]")
	assert.Equal(t, n["$[1]"], node.Content[1])
}
