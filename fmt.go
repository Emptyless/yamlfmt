package yamlfmt

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// root path matches the yaml document node
const root = "$"

// all path matches any array or map element
const all = "*"

// whitespace to use when encoding the yaml.Node to []byte
const whitespace = 2

// LintBytes is a utility method that unmarshaled the provided bytes into a yaml.Node,
// applies Lint on the node and returns the marshaled result
func LintBytes(b []byte, rules []Rule) ([]byte, error) {
	if len(b) == 0 {
		return b, nil
	}

	// unmarshal into yaml.Node
	node := new(yaml.Node)
	err := yaml.Unmarshal(b, node)
	if err != nil {
		return nil, err
	}

	// lint the node with provided rules
	Lint(node, rules)

	// encode back into bytes
	writer := new(bytes.Buffer)
	encoder := yaml.NewEncoder(writer)
	encoder.SetIndent(whitespace)
	err = encoder.Encode(node)
	if err != nil {
		return nil, err
	}

	return writer.Bytes(), nil
}

// Lint a yaml.Node the provided slice of Rule
func Lint(node *yaml.Node, rules []Rule) { //nolint:cyclop // accepted
	if node == nil || len(rules) == 0 {
		return
	}

	cursor := node
	var path string
	if node.Kind == yaml.DocumentNode { // document node, the root path starts with $
		path = root

		if len(cursor.Content) == 0 {
			return // nothing to order
		}
		cursor = cursor.Content[0]
	}

	for _, rule := range rules {
		queue := map[string]*yaml.Node{path: cursor}
		for len(queue) > 0 {
			// dequeue key=value pair
			var key string
			var node *yaml.Node
			for k, v := range queue {
				key = k
				node = v
				delete(queue, k)
				break // dequeue
			}

			// match rule
			if !rule.contains(key) {
				continue
			}

			// add next to queue
			// note that this is not optimal if a match is final as the last layer will be added
			// even though it can never match, this is accepted to reduce the complexity of the solution
			for k, v := range next(key, node) {
				queue[k] = v
			}

			// if match, run Fn
			if rule.match(key) {
				rule.Run(key, node)
			}
		}
	}
}

// Validate rules that there are no parse errors
func Validate(rules []Rule) error {
	if len(rules) == 0 {
		return nil
	}

	var err error
	for _, rule := range rules {
		_, ruleErr := parts(rule.Path)
		if ruleErr != nil {
			err = errors.Join(err, ruleErr)
		}
	}

	return err
}

// OrderFn is executed on some yaml.Node with the key being a JSONPath like value
type OrderFn func(key string, value *yaml.Node)

// Rule combines a JSONPath like syntax to an OrderFn to apply on the node
type Rule struct {
	// Path using a JSONPath like syntax (except filtering):
	// '$' is the document root
	// '$.key' to select a key
	// '$[0]' to select some index
	// '$[*]' or '$.*' for wildcard searches
	// '$.some[*].*.name' can combine any of above rules
	Path string
	// Functions to execute if there is a Path match
	Functions []OrderFn
}

// NewRule constructor for Rule
func NewRule(path string, fns ...OrderFn) Rule {
	return Rule{Path: path, Functions: fns}
}

// Run Rule.Functions for given key and value
func (r *Rule) Run(key string, value *yaml.Node) {
	for _, fn := range r.Functions {
		fn(key, value)
	}
}

// contains returns true iff the path is still possible from the Rule.Path
func (r *Rule) contains(path string) bool {
	if !strings.HasPrefix(r.Path, root) {
		return true // if the match is relative (e.g. '.schema') traverse everything
	}

	pathParts, pathErr := parts(path)
	if pathErr != nil {
		panic(pathErr) // invalid rules supplied, use Validate to catch ahead of time
	}
	rulePathParts, rulePathErr := parts(r.Path)
	if rulePathErr != nil {
		panic(rulePathErr) // invalid rules supplied, use Validate to catch ahead of  time
	}

	// iterate over path parts to check if it matches the rule parts
	for i, pathPart := range pathParts {
		if i > len(rulePathParts)-1 {
			return false // to deep
		}

		rulePart := rulePathParts[i]
		if !strings.EqualFold(pathPart, rulePart) &&
			!strings.EqualFold(rulePart, all) &&
			!strings.EqualFold(rulePart, indexOpen+all+indexClose) {
			return false
		}
	}

	return true
}

// match returns true iff the Rule.Path matches the provided path
func (r *Rule) match(path string) bool { //nolint:cyclop // accepted
	pathParts, pathErr := parts(path)
	if pathErr != nil {
		panic(pathErr) // invalid rules supplied, use Validate to catch ahead of  time
	}
	rulePathParts, rulePathErr := parts(r.Path)
	if rulePathErr != nil {
		panic(rulePathErr) // invalid rules supplied, use Validate to catch ahead of  time
	}

	relative := !strings.HasPrefix(r.Path, root)
	if !relative && len(pathParts) != len(rulePathParts) {
		return false
	}

	// if relative, go right to left from perspective of rule
	if relative {
		for i := len(rulePathParts) - 1; i >= 0; i-- {
			rulePart := rulePathParts[i]
			pathPartIndex := (len(pathParts) - 1) - ((len(rulePathParts) - 1) - i)
			if pathPartIndex > len(pathParts)-1 {
				return false // path did not match
			}
			pathPart := pathParts[pathPartIndex]
			if !strings.EqualFold(pathPart, rulePart) &&
				!strings.EqualFold(rulePart, delimiter+all) &&
				!strings.EqualFold(rulePart, indexOpen+all+indexClose) {
				return false
			}
		}

		return true
	}

	// iterate over path parts to check if it matches the rule parts
	for i, pathPart := range pathParts {
		if i > len(rulePathParts)-1 {
			return false // to deep
		}

		rulePart := rulePathParts[i]
		if !strings.EqualFold(pathPart, rulePart) &&
			!strings.EqualFold(rulePart, delimiter+all) &&
			!strings.EqualFold(rulePart, indexOpen+all+indexClose) {
			return false
		}
	}

	return true
}

// check if two string parts match
// func check(rulePart string, pathPart string) bool {
//	if strings.HasPrefix(rulePart, delimiter) {
//		return strings.EqualFold(rulePart, pathPart)
//	}
//}

// NewSimpleOrdering sorts the keys of a yaml.MappingNode in the order provided with "keys". If keys that are present
// in the YAML are not present in the supplied keys, the original order of that key will be preserved after all supplied
// keys are processed. e.g. NewSimpleOrdering("a", "b", "c") on a yaml node with keys ["c", "f", "b", "e", "a"] will result
// in ["a", "b", "c", "f", "e"] (notice how a, b, c are at the front in order provided and remaining keys are in their original order)
func NewSimpleOrdering(keys ...string) OrderFn { //nolint:cyclop // accepted
	return func(_ string, value *yaml.Node) {
		if len(keys) == 0 || value == nil || len(value.Content) == 0 || value.Kind != yaml.MappingNode {
			return
		}

		type Pair struct {
			Key   *yaml.Node
			Value *yaml.Node
		}

		// gather ordering
		var nodes []Pair
		for i, node := range value.Content {
			if i%2 == 0 {
				continue
			}

			nodes = append(nodes, Pair{Key: value.Content[i-1], Value: node})
		}

		sorted := slices.SortedFunc(slices.Values(nodes), func(e Pair, e2 Pair) int {
			eIdx := slices.Index(keys, e.Key.Value)
			e2Idx := slices.Index(keys, e2.Key.Value)
			switch {
			case eIdx >= 0 && e2Idx >= 0: // if both are contained in the keys, use standard compare
				return cmp.Compare(eIdx, e2Idx)
			case eIdx >= 0: // if e2 is not contained, e < e2
				return -1
			case e2Idx >= 0: // if e is not contained e > e2
				return 1
			}

			return cmp.Compare(slices.Index(nodes, e), slices.Index(nodes, e2))
		})

		for i, pair := range sorted {
			value.Content[i*2] = pair.Key
			value.Content[i*2+1] = pair.Value
		}
	}
}

// StringOrderingFn sorts yaml.MappingNode and yaml.SequenceNode on their yaml.Node.Value using default cmp.Compare function
func StringOrderingFn(_ string, value *yaml.Node) {
	if value == nil || len(value.Content) == 0 || (value.Kind != yaml.MappingNode && value.Kind != yaml.SequenceNode) {
		return // only sort mapping nodes that have values
	}

	if value.Kind == yaml.SequenceNode {
		sorted := slices.SortedFunc(slices.Values(value.Content), func(e *yaml.Node, e2 *yaml.Node) int {
			return cmp.Compare(e.Value, e2.Value)
		})

		copy(value.Content, sorted)

		return
	}

	// Pair utility to have a single array entry that has both the Key and Value
	type Pair struct {
		Key   *yaml.Node
		Value *yaml.Node
	}

	nodes := make([]Pair, len(value.Content)/2) //nolint:mnd // 2 denotes that a key=value pair is two yaml.Node's
	for i, node := range value.Content {
		if i%2 == 0 {
			continue
		}

		nodes[i/2] = Pair{Key: value.Content[i-1], Value: node}
	}

	sorted := slices.SortedFunc(slices.Values(nodes), func(e Pair, e2 Pair) int {
		return cmp.Compare(e.Key.Value, e2.Key.Value)
	})

	for i, pair := range sorted {
		value.Content[i*2] = pair.Key
		value.Content[i*2+1] = pair.Value
	}
}

// token is a special character used during parts parsing of a path
type token = string

// delimiter token denotes a level in the yaml hierarchy
const delimiter token = "."

// indexOpen denotes the start of an array indexing operation
const indexOpen token = "["

// indexClose desnotes the closure of an array indexing operation
const indexClose token = "]"

// ErrIllegalToken is returned when a token is used that is not expected, e.g. two indexOpen tokens [[ sequentially
var ErrIllegalToken = errors.New("illegal token")

// parseToken char tokens and if parseToken check if it is an allowed token based on allowed
// if parseToken, return the new set of allowed tokens
func parseToken(char string, allowed []token) (bool, []token, error) {
	res := char == delimiter || char == indexOpen || char == indexClose
	if res && !slices.Contains(allowed, char) {
		return false, nil, fmt.Errorf("char %q: %w", char, ErrIllegalToken)
	}

	if res {
		switch char {
		case indexClose:
			allowed = []token{delimiter, indexOpen}
		case delimiter:
			allowed = []token{delimiter, indexOpen}
		case indexOpen:
			allowed = []token{indexClose}
		}
	}

	return res, allowed, nil
}

// parts of the path using a very simple token parser. There is room for improvement as currently there is no
// escaping of tokens in a path that could be valid in yaml, e.g. a yaml key "key.Name" can only be used in the document
// by also writing a rule for '$.key.name' instead of e.g. '$.['key.Name']'
func parts(path string) ([]string, error) {
	if path == "" {
		return []string{}, nil
	}

	var res []string
	var start int
	var cursor int
	allowed := []token{indexOpen, delimiter}
	for cursor < len(path) {
		var isToken bool
		var err error
		isToken, allowed, err = parseToken(string(path[cursor]), allowed)
		if err != nil {
			return nil, fmt.Errorf("invalid path %q: %w", path, err)
		}

		// if token, add part
		if isToken && start != cursor && (string(path[cursor]) == delimiter || string(path[cursor]) == indexOpen) {
			res = append(res, path[start:cursor])
			start = cursor
			cursor++
			continue
		}

		cursor++
	}

	if start != cursor {
		res = append(res, path[start:cursor])
	}

	return res, nil
}

// next paths that can be taken on the node which will be suffixed to the passed cursor,
// e.g. for a cursor '$' and mapping node with key 'key' => '$.key' or
// a cursor '$' and a sequence node => '$[0]'
func next(cursor string, node *yaml.Node) map[string]*yaml.Node {
	res := map[string]*yaml.Node{}
	if node == nil {
		return res
	}

	if node.Content != nil && node.Kind == yaml.MappingNode {
		var key string
		for i, content := range node.Content {
			if i%2 == 0 {
				key = cursor + delimiter + content.Value
				continue
			}

			res[key] = content
		}
	} else if node.Content != nil && node.Kind == yaml.SequenceNode {
		for i, content := range node.Content {
			key := cursor + indexOpen + strconv.Itoa(i) + indexClose
			res[key] = content
		}
	}

	return res
}
