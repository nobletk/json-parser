package parser

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/nobletk/json-parser/internal/ast"
	"github.com/nobletk/json-parser/internal/lexer"
	"github.com/nobletk/json-parser/internal/token"
	"github.com/nobletk/json-parser/pkg/mylog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutToJSON(t *testing.T) {
	input := `{
	       "key1": "value",
	       "key2": -123,
	       "key3": ["value", 1, true, null, -0.2e2],
	       "key4": {
	               "key4": null
	       }
	   }
	   `
	log := mylog.CreateLogger(true)
	l := lexer.New(log, string(input))
	p := New(l)
	jf, jsonErr := p.ParseFile()
	require.Empty(t, jsonErr, "jsonErr should be empty")
	assert.Len(t, jf.Elements, 1, "length of elements isn't correct")

	jsonData, err := json.MarshalIndent(jf.ToInterface(), "", "  ")
	assert.Empty(t, err, "err should be empty")

	expected := `{
  "key1": "value",
  "key2": -123,
  "key3": [
    "value",
    1,
    true,
    null,
    -20
  ],
  "key4": {
    "key4": null
  }
}`
	assert.Equal(t, expected, string(jsonData))
}

func TestValidJSONObject(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]interface{}
	}{
		{
			name:     "Empty JSON",
			input:    `{}`,
			expected: map[string]interface{}{},
		},
		{
			name:  "JSON With String Values",
			input: `{"key1": "value1", "key2": "value2", "key3": "value3"}`,
			expected: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name:  "JSON With Positive Number Values",
			input: `{"key1": 123, "key2": 95, "key3": 15}`,
			expected: map[string]interface{}{
				"key1": 123,
				"key2": 95,
				"key3": 15,
			},
		},
		{
			name:  "JSON With Negative Number Values",
			input: `{"key1": -100, "key2": -95, "key3": -15}`,
			expected: map[string]interface{}{
				"key1": -100,
				"key2": -95,
				"key3": -15,
			},
		},
		{
			name:  "JSON With Decimal Number Values",
			input: `{"key1": 1.69, "key2": -9.5, "key3": 0.15}`,
			expected: map[string]interface{}{
				"key1": 1.69,
				"key2": -9.5,
				"key3": 0.15,
			},
		},
		{
			name:  "JSON With Exponent Values",
			input: `{"key1": 1e2, "key2": 3E+44, "key3": -0.5e-6}`,
			expected: map[string]interface{}{
				"key1": 1e2,
				"key2": 3e+44,
				"key3": -0.5e-6,
			},
		},
		{
			name:  "JSON With Boolean Values",
			input: `{"key1": true, "key2": false}`,
			expected: map[string]interface{}{
				"key1": true,
				"key2": false,
			},
		},
		{
			name:  "JSON With Null Value",
			input: `{"key1": null}`,
			expected: map[string]interface{}{
				"key1": nil,
			},
		},
		{
			name: "JSON With Different Value Types",
			input: `{
	       "key1": "value",
	       "key2": -123,
	       "key3": ["value", 1, true, null, -0.2e2],
	       "key4": {
	               "key4": null
	       }
	   }`,
			expected: map[string]interface{}{
				"key1": "value",
				"key2": -123,
				"key3": []interface{}{"value", 1, true, nil, -0.2e2},
				"key4": map[string]interface{}{"key4": nil},
			},
		},
		{
			name:     "JSON With Escaped Sequences Strings",
			input:    "{\"\\\"\\\"key\\u00Fa\\b\\\"\": \"value\\n\\t\\f\"}",
			expected: map[string]interface{}{"\\\"\\\"key\\u00Fa\\b\\\"": "value\\n\\t\\f"},
		},
		{
			name:     "JSON With Escaped Sequences Array",
			input:    "{\"\\\"\\\"key\\u00FaL1\\b\\\"\": [\"value\\n\\t\\f\"]}",
			expected: map[string]interface{}{"\\\"\\\"key\\u00FaL1\\b\\\"": []interface{}{"value\\n\\t\\f"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := mylog.CreateLogger(true)
			l := lexer.New(log, string(tt.input))
			p := New(l)
			jf, jErr := p.ParseFile()
			require.NotEmpty(t, jf, "jf is empty")
			require.Empty(t, jErr, "jsonErr should be empty")
			require.Len(t, jf.Elements, 1, "length of elements isn't correct")
			assertObjectLiteral(t, jf.Elements[0], tt.expected)
		})
	}
}

func TestValidJSONArray(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []interface{}
	}{
		{
			name:     "Empty Array",
			input:    `[]`,
			expected: []interface{}{},
		},
		{
			name:     "Empty Multiple Arrays",
			input:    `[[[]]]`,
			expected: []interface{}{[]interface{}{[]interface{}{}}},
		},
		{
			name:     "Array With Different Value Types",
			input:    `["value", 1, true, null, -0.2e2, {"key": 123}]`,
			expected: []interface{}{"value", 1, true, nil, -0.2e2, map[string]interface{}{"key": 123}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := mylog.CreateLogger(true)
			l := lexer.New(log, string(tt.input))
			p := New(l)
			jf, jErr := p.ParseFile()
			require.NotEmpty(t, jf, "jf is empty")
			require.Empty(t, jErr, "jsonErr should be empty")
			require.Len(t, jf.Elements, 1, "length of elements isn't correct")
			assertArrayLiteral(t, jf.Elements[0], tt.expected)
		})
	}
}

func TestInvalidJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr *JSONErr
	}{
		{
			name:  "Empty Input",
			input: ``,
			expectedErr: &JSONErr{
				Msg: "Expected '{' or '[', got 'EOF' instead\n",
				Pos: token.Position{
					Column: 1,
					Line:   1,
				},
			},
		},
		{
			name:  "Multiple Empty Objects",
			input: `{{}}`,
			expectedErr: &JSONErr{
				Msg: "Expected 'STRING', '}', got '{' instead\n",
				Pos: token.Position{
					Column: 2,
					Line:   1,
				},
			},
		},
		{
			name:  "Not An Object Nor An Array",
			input: `"string"`,
			expectedErr: &JSONErr{
				Msg: "Expected '{' or '[', got 'STRING' instead\n",
				Pos: token.Position{
					Column: 1,
					Line:   1,
				},
			},
		},
		{
			name:  "Unclosed Object",
			input: `{`,
			expectedErr: &JSONErr{
				Msg: "Expected 'STRING', '}', got 'EOF' instead\n",
				Pos: token.Position{
					Column: 2,
					Line:   1,
				},
			},
		},
		{
			name:  "Trailing Comma After Left Brace",
			input: `{,`,
			expectedErr: &JSONErr{
				Msg: "Expected 'STRING', '}', got ',' instead\n",
				Pos: token.Position{
					Column: 2,
					Line:   1,
				},
			},
		},
		{
			name:  "Trailing Comma After Property",
			input: `{"key":,`,
			expectedErr: &JSONErr{
				Msg: "Expected 'STRING', 'NUMBER', 'NULL', 'TRUE', 'FALSE', '{', '[', got ',' instead\n",
				Pos: token.Position{
					Line:   1,
					Column: 8,
				},
			},
		},
		{
			name:  "Unclosed Array",
			input: `[`,
			expectedErr: &JSONErr{
				Msg: "Expected 'STRING', 'NUMBER', 'NULL', 'TRUE', 'FALSE', '{', '[', ']', got 'EOF' instead\n",
				Pos: token.Position{
					Column: 2,
					Line:   1,
				},
			},
		},
		{
			name:  "Wrong Closing For An Array",
			input: `[}`,
			expectedErr: &JSONErr{
				Msg: "Expected 'STRING', 'NUMBER', 'NULL', 'TRUE', 'FALSE', '{', '[', ']', got '}' instead\n",
				Pos: token.Position{
					Column: 2,
					Line:   1,
				},
			},
		},
		{
			name:  "Trailing Comma In An Array",
			input: `[,`,
			expectedErr: &JSONErr{
				Msg: "Expected 'STRING', 'NUMBER', 'NULL', 'TRUE', 'FALSE', '{', '[', ']', got ',' instead\n",
				Pos: token.Position{
					Column: 2,
					Line:   1,
				},
			},
		},
		{
			name:  "Duplicate JSON Properties",
			input: `{"key1": "value1", "key2": "value2", "key1": "value3"}`,
			expectedErr: &JSONErr{
				Msg: "Duplicate JSON property '\"key1\"'\n",
				Pos: token.Position{
					Column: 44,
					Line:   1,
				},
			},
		},
		{
			name:  "Minus Not Followed By A Number",
			input: `{"key1": - }`,
			expectedErr: &JSONErr{
				Msg: "Expected 'STRING', 'NUMBER', 'NULL', 'TRUE', 'FALSE', '{', '[', got 'ILLEGAL' instead\n",
				Pos: token.Position{
					Column: 10,
					Line:   1,
				},
			},
		},
		{
			name:  "Minus Followed By Space",
			input: `{"key1": - 1}`,
			expectedErr: &JSONErr{
				Msg: "Expected 'STRING', 'NUMBER', 'NULL', 'TRUE', 'FALSE', '{', '[', got 'ILLEGAL' instead\n",
				Pos: token.Position{
					Column: 10,
					Line:   1,
				},
			},
		},
		{
			name:  "Decimal With No Leading Digit",
			input: `{"key1": -.95}`,
			expectedErr: &JSONErr{
				Msg: "Expected 'STRING', 'NUMBER', 'NULL', 'TRUE', 'FALSE', '{', '[', got 'ILLEGAL' instead\n",
				Pos: token.Position{
					Column: 10,
					Line:   1,
				},
			},
		},
		{
			name:  "Invalid Exponent",
			input: `{"key1": -100e}`,
			expectedErr: &JSONErr{
				Msg: "Expected 'STRING', 'NUMBER', 'NULL', 'TRUE', 'FALSE', '{', '[', got 'ILLEGAL' instead\n",
				Pos: token.Position{
					Column: 10,
					Line:   1,
				},
			},
		},
		{
			name:  "Not A Property Quotations",
			input: `{key1: 0}`,
			expectedErr: &JSONErr{
				Msg: "Expected 'STRING', '}', got 'ILLEGAL' instead\n",
				Pos: token.Position{
					Column: 2,
					Line:   1,
				},
			},
		},
		{
			name:  "Trailing Comma In An Object",
			input: `{"key1": "value1", }`,
			expectedErr: &JSONErr{
				Msg: "Expected 'STRING', got '}' instead\n",
				Pos: token.Position{
					Column: 20,
					Line:   1,
				},
			},
		},
		{
			name:  "Trailing Comma In An Array",
			input: `["value1", ]`,
			expectedErr: &JSONErr{
				Msg: "Expected 'STRING', 'NUMBER', 'TRUE', 'FALSE', 'NULL'. got ']' instead\n",
				Pos: token.Position{
					Column: 12,
					Line:   1,
				},
			},
		},
		{
			name:  "Comma After A Right Bracket Of An Array Instead of EOF",
			input: `["value1"],`,
			expectedErr: &JSONErr{
				Msg: "Expected 'EOF', got ',' instead\n",
				Pos: token.Position{
					Column: 12,
					Line:   1,
				},
			},
		},
		{
			name:  "No String After Comma In A JSON Object",
			input: `{"key": ["value1"], }`,
			expectedErr: &JSONErr{
				Msg: "Expected 'STRING', got '}' instead\n",
				Pos: token.Position{
					Column: 21,
					Line:   1,
				},
			},
		},
		{
			name:  "String After Right Curly Brace Instead Of EOF",
			input: `{"key": "value1"} "misplaced quoted value"`,
			expectedErr: &JSONErr{
				Msg: "Expected 'EOF', got 'STRING' instead\n",
				Pos: token.Position{
					Column: 43,
					Line:   1,
				},
			},
		},
		{
			name:  "Illegal Token String Without Double Quotations",
			input: "{\"\\\"\\\"key\\u00Fa\\b\\\"\": [value, 1]}",
			expectedErr: &JSONErr{
				Msg: "Expected 'STRING', 'NUMBER', 'NULL', 'TRUE', 'FALSE', '{', '[', ']', got 'ILLEGAL' instead\n",
				Pos: token.Position{
					Column: 24,
					Line:   1,
				},
			},
		},
		{
			name:  "Illegal Token String With Closing Quotation",
			input: "[\"value]",
			expectedErr: &JSONErr{
				Msg: "Expected 'STRING', 'NUMBER', 'NULL', 'TRUE', 'FALSE', '{', '[', ']', got 'ILLEGAL' instead\n",
				Pos: token.Position{
					Column: 2,
					Line:   1,
				},
			},
		},
		{
			name:  "Invalid Hexadecimal Unicode Within String",
			input: "{\"key\\u00FZ\": 1}",
			expectedErr: &JSONErr{
				Msg: "Invalid unicode escape sequence\n",
				Pos: token.Position{
					Column: 2,
					Line:   1,
				},
			},
		},
		{
			name:  "Invalid Unicode Escape Sequence Less than 4 Hexadecimal Within String",
			input: "{\"key\\u00F\": 1}",
			expectedErr: &JSONErr{
				Msg: "Invalid unicode escape sequence\n",
				Pos: token.Position{
					Column: 2,
					Line:   1,
				},
			},
		},
		{
			name:  "Single Quotations Wrapping String",
			input: "{\"key\": ['value']}",
			expectedErr: &JSONErr{
				Msg: "Expected 'STRING', 'NUMBER', 'NULL', 'TRUE', 'FALSE', '{', '[', ']', got 'ILLEGAL' instead\n",
				Pos: token.Position{
					Column: 10,
					Line:   1,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := mylog.CreateLogger(true)
			l := lexer.New(log, string(tt.input))
			p := New(l)
			jf, jErr := p.ParseFile()

			assert.Empty(t, jf, "jsonFile should be empty")
			assert.Equal(t, tt.expectedErr, jErr)
		})
	}
}

func TestParseArray(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []interface{}
	}{
		{
			name:     "Empty Array",
			input:    `[]`,
			expected: []interface{}{},
		},
		{
			name:     "Array Of Integers",
			input:    `[1, 2, 3]`,
			expected: []interface{}{1, 2, 3},
		},
		{
			name:     "Array Of Strings",
			input:    `["x", "y", "z"]`,
			expected: []interface{}{"x", "y", "z"},
		},
		{
			name:     "Array Of Booleans And Null",
			input:    `[true, false, null]`,
			expected: []interface{}{true, false, nil},
		},
		{
			name:     "Array Of Arrays Of Integers",
			input:    `[[1, 2], [3, 4]]`,
			expected: []interface{}{[]interface{}{1, 2}, []interface{}{3, 4}},
		},
		{
			name:     "Array Of Nested Arrays And Strings",
			input:    `["x", ["y", "z"], "w"]`,
			expected: []interface{}{"x", []interface{}{"y", "z"}, "w"},
		},
		{
			name:     "Mixed-Type Array",
			input:    `[123, "Value", true, null, -4.56e2]`,
			expected: []interface{}{123, "Value", true, nil, -4.56e2},
		},
		{
			name:     "Array With String, Nested Array And Object",
			input:    `["test", [1, 2, 3], {"key": "value"}]`,
			expected: []interface{}{"test", []interface{}{1, 2, 3}, map[string]interface{}{"key": "value"}},
		},
		{
			name:     "Array With Nested Arrays Of Booleans And Mixed Values",
			input:    `[[true, false], [null, 123]]`,
			expected: []interface{}{[]interface{}{true, false}, []interface{}{nil, 123}},
		},
		{
			name:     "Deeply Nested Empty Arrays",
			input:    `[[[]]]`,
			expected: []interface{}{[]interface{}{[]interface{}{}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := mylog.CreateLogger(true)
			l := lexer.New(log, string(tt.input))
			p := New(l)
			actual, jsonErr := p.parseArray()
			require.Empty(t, jsonErr, "jsonErr should be empty")
			assertArrayLiteral(t, actual, tt.expected)
		})
	}
}

func TestParseString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Regular String",
			input:    "\"key1\"",
			expected: "key1",
		},
		{
			name:     "String with Escaped Quotation Marks",
			input:    "\"\\\"key\\\"\"",
			expected: "\\\"key\\\"",
		},
		{
			name:     "String with Escaped Unbalanced Quotation Marks",
			input:    "\"\\\"\\\"\\\"key\\\"\"",
			expected: "\\\"\\\"\\\"key\\\"",
		},
		{
			name:     "String with Escaped Reverse Solidus",
			input:    "\"key\\\\\"",
			expected: "key\\\\",
		},
		{
			name:     "String with Escaped Solidus",
			input:    "\"key\\/\"",
			expected: "key\\/",
		},
		{
			name:     "String with Escaped Backspace",
			input:    "\"key\\b\"",
			expected: "key\\b",
		},
		{
			name:     "String with Escaped Formfeed",
			input:    "\"key\\f\"",
			expected: "key\\f",
		},
		{
			name:     "String with Escaped Linefeed",
			input:    "\"key\\n\"",
			expected: "key\\n",
		},
		{
			name:     "String with Escaped Carriage Return",
			input:    "\"key\\r\"",
			expected: "key\\r",
		},
		{
			name:     "String with Escaped Horizontal Tab",
			input:    "\"key\\t\"",
			expected: "key\\t",
		},
		{
			name:     "String with Escaped Unicode Sequence",
			input:    "\"key\\u00Fa\"",
			expected: "key\\u00Fa",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := mylog.CreateLogger(true)
			l := lexer.New(log, string(tt.input))
			p := New(l)
			actual, jsonErr := p.parseString()
			require.Empty(t, jsonErr, "jsonErr should be empty")
			assertStringLiteral(t, actual, tt.expected)
		})
	}
}

func TestParseNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name:     "Positive Integer",
			input:    "1234567890",
			expected: 1234567890,
		},
		{
			name:     "Negative Integer",
			input:    "-1234567890",
			expected: -1234567890,
		},
		{
			name:     "Positive Float",
			input:    "0987654321.1234567890",
			expected: 0987654321.1234567890,
		},
		{
			name:     "Negative Float",
			input:    "-0987654321.1234567890",
			expected: -0987654321.1234567890,
		},
		{
			name:     "Scientific Notation",
			input:    "1.23e4",
			expected: 1.23e4,
		},
		{
			name:     "Uppercase Exponent Negative Power",
			input:    "5E-10",
			expected: 5e-10,
		},
		{
			name:     "Float With Uppercase Exponent",
			input:    "2.0E3",
			expected: 2.0e3,
		},
		{
			name:     "Negative Float Scientific Notation",
			input:    "-1.23e-4",
			expected: -1.23e-4,
		},
		{
			name:     "Negative Float With Uppercase Exponent",
			input:    "-6E10",
			expected: -6e10,
		},
		{
			name:     "Very Large Number Upper Boundary",
			input:    "1e308",
			expected: 1e308,
		},
		{
			name:     "Very Large Negative Number",
			input:    "-1e308",
			expected: -1e308,
		},
		{
			name:     "Very Small Number Lower Boundary",
			input:    "5e-324",
			expected: 5e-324,
		},
		{
			name:     "Very Small Negative Number",
			input:    "-5e-324",
			expected: -5e-324,
		},
		{
			name:     "Representation Of Zero",
			input:    "-0",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := mylog.CreateLogger(true)
			l := lexer.New(log, string(tt.input))
			p := New(l)
			actual, jsonErr := p.parseNumber()
			require.Empty(t, jsonErr, "jsonErr should be empty")
			assertNumberLiteral(t, actual, tt.expected)
		})
	}
}

func assertArrayLiteral(t *testing.T, actual ast.Element, expected []interface{}) bool {
	al, ok := actual.(*ast.ArrayLiteral)
	if !ok {
		t.Errorf("al not *ast.ArrayLiteral. got=%T", actual)
		return false
	}

	if len(al.Elements) != len(expected) {
		t.Errorf("len(expected)=%d. got=%d", len(expected), len(al.Elements))
		return false
	}

	for i, exp := range expected {
		switch e := exp.(type) {
		case string:
			if !assertStringLiteral(t, al.Elements[i], e) {
				return false
			}
		case int:
			if !assertNumberLiteral(t, al.Elements[i], float64(e)) {
				return false
			}
		case float64:
			if !assertNumberLiteral(t, al.Elements[i], e) {
				return false
			}
		case bool:
			if !assertBooleanLiteral(t, al.Elements[i], e) {
				return false
			}
		case nil:
			if !assertNullLiteral(t, al.Elements[i], "null") {
				return false
			}
		case []interface{}:
			if !assertArrayLiteral(t, al.Elements[i], e) {
				return false
			}
		case map[string]interface{}:
			if !assertObjectLiteral(t, al.Elements[i], e) {
				return false
			}
		default:
			t.Errorf("unexpected type '%T' at index %d", exp, i)
			return false
		}
	}

	return true
}

func assertObjectLiteral(t *testing.T, actual ast.Element, expected map[string]interface{}) bool {
	obj, ok := actual.(*ast.Object)
	if !ok {
		t.Errorf("obj not *ast.Object. got=%T", actual)
		return false
	}

	if len(obj.Pairs) != len(expected) {
		t.Errorf("len(obj.Pairs)=%d. expected=%d", len(obj.Pairs), len(expected))
		return false
	}

	for expectedKey, expectedValue := range expected {
		matched := false

		for k, v := range obj.Pairs {
			if k.TokenLiteral() == expectedKey {
				matched = true
				switch expectedValue := expectedValue.(type) {
				case string:
					if !assertStringLiteral(t, v, expectedValue) {
						return false
					}
				case int:
					if !assertNumberLiteral(t, v, float64(expectedValue)) {
						return false
					}
				case float64:
					if !assertNumberLiteral(t, v, expectedValue) {
						return false
					}
				case bool:
					if !assertBooleanLiteral(t, v, expectedValue) {
						return false
					}
				case nil:
					if !assertNullLiteral(t, v, "null") {
						return false
					}
				case map[string]interface{}:
					if !assertObjectLiteral(t, v, expectedValue) {
						return false
					}
				case []interface{}:
					if !assertArrayLiteral(t, v, expectedValue) {
						return false
					}
				default:
					t.Errorf("unexpected type '%T' for key '%s'", expectedValue, expectedKey)
					return false
				}
				break
			}
		}

		if !matched {
			t.Errorf("key '%s' not found in obj.Pairs", expectedKey)
			return false
		}
	}
	return true
}

func assertStringLiteral(t *testing.T, actual ast.Element, expected string) bool {
	str, ok := actual.(*ast.StringLiteral)
	if !ok {
		t.Errorf("el not *ast.StringLiteral. got=%T", actual)
		return false
	}

	if str.Value != expected {
		t.Errorf("str.Value not %s. got=%s", expected, str.Value)
		return false
	}

	if str.TokenLiteral() != expected {
		t.Errorf("str.TokenLiteral not %s. got = %s", expected,
			str.TokenLiteral())
		return false
	}

	return true
}

func assertBooleanLiteral(t *testing.T, actual ast.Element, expected bool) bool {
	bl, ok := actual.(*ast.Boolean)
	if !ok {
		t.Errorf("el not *ast.Boolean. got=%T", actual)
		return false
	}

	if bl.Value != expected {
		t.Errorf("bl.Value not %t, got=%t", expected, bl.Value)
		return false
	}

	if bl.TokenLiteral() != fmt.Sprintf("%t", expected) {
		t.Errorf("bl.TokenLiteral not %t, got=%t", expected, bl.Value)
		return false
	}

	return true
}

func assertNullLiteral(t *testing.T, actual ast.Element, expected string) bool {
	nl, ok := actual.(*ast.Null)
	if !ok {
		t.Errorf("il not *ast.Null. got=%T", actual)
		return false
	}

	if nl.Value != expected {
		t.Errorf("nl.String() not %s. got=%s", expected, nl.Value)
		return false
	}

	if nl.TokenLiteral() != expected {
		t.Errorf("integ.TokenLiteral not %s. got=%s", expected,
			nl.TokenLiteral())
		return false
	}

	return true
}

func assertNumberLiteral(t *testing.T, actual ast.Element, expected float64) bool {
	num, ok := actual.(*ast.NumberLiteral)
	if !ok {
		t.Errorf("num not *ast.IntegerLiteral. got=%T", actual)
		return false
	}

	if num.Value != expected {
		t.Errorf("num.Value not %v. got=%v", expected, num.Value)
		return false
	}

	return true
}
