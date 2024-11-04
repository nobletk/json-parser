package lexer

import (
	"testing"

	"github.com/nobletk/json-parser/internal/token"
	"github.com/nobletk/json-parser/pkg/mylog"
	"github.com/stretchr/testify/assert"
)

func TestNextToken(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{
			name:            "Left Curly Brace",
			input:           "{",
			expectedType:    token.LBRACE,
			expectedLiteral: "{",
		},
		{
			name:            "Right Curly Brace",
			input:           "}",
			expectedType:    token.RBRACE,
			expectedLiteral: "}",
		},
		{
			name:            "Left Bracket",
			input:           "[",
			expectedType:    token.LBRACKET,
			expectedLiteral: "[",
		},
		{
			name:            "Right Bracket",
			input:           "]",
			expectedType:    token.RBRACKET,
			expectedLiteral: "]",
		},
		{
			name:            "Comma",
			input:           ",",
			expectedType:    token.COMMA,
			expectedLiteral: ",",
		},
		{
			name:            "Colon",
			input:           ":",
			expectedType:    token.COLON,
			expectedLiteral: ":",
		},
		{
			name:            "True",
			input:           "true",
			expectedType:    token.TRUE,
			expectedLiteral: "true",
		},
		{
			name:            "False",
			input:           "false",
			expectedType:    token.FALSE,
			expectedLiteral: "false",
		},
		{
			name:            "Null",
			input:           "null",
			expectedType:    token.NULL,
			expectedLiteral: "null",
		},
		{
			name:            "Regular String",
			input:           "\"key1\"",
			expectedType:    token.STRING,
			expectedLiteral: "key1",
		},
		{
			name:            "String with Escaped Quotation Marks",
			input:           "\"\\\"key\\\"\"",
			expectedType:    token.STRING,
			expectedLiteral: "\\\"key\\\"",
		},
		{
			name:            "String with Escaped Unbalanced Quotation Marks",
			input:           "\"\\\"\\\"\\\"key\\\"\"",
			expectedType:    token.STRING,
			expectedLiteral: "\\\"\\\"\\\"key\\\"",
		},
		{
			name:            "String with Escaped Reverse Solidus",
			input:           "\"key\\\\\"",
			expectedType:    token.STRING,
			expectedLiteral: "key\\\\",
		},
		{
			name:            "String with Escaped Solidus",
			input:           "\"key\\/\"",
			expectedType:    token.STRING,
			expectedLiteral: "key\\/",
		},
		{
			name:            "String with Escaped Backspace",
			input:           "\"key\\b\"",
			expectedType:    token.STRING,
			expectedLiteral: "key\\b",
		},
		{
			name:            "String with Escaped Formfeed",
			input:           "\"key\\f\"",
			expectedType:    token.STRING,
			expectedLiteral: "key\\f",
		},
		{
			name:            "String with Escaped Linefeed",
			input:           "\"key\\n\"",
			expectedType:    token.STRING,
			expectedLiteral: "key\\n",
		},
		{
			name:            "String with Escaped Carriage Return",
			input:           "\"key\\r\"",
			expectedType:    token.STRING,
			expectedLiteral: "key\\r",
		},
		{
			name:            "String with Escaped Horizontal Tab",
			input:           "\"key\\t\"",
			expectedType:    token.STRING,
			expectedLiteral: "key\\t",
		},
		{
			name:            "String with Escaped Unicode Sequence",
			input:           "\"key\\u00Fa\"",
			expectedType:    token.STRING,
			expectedLiteral: "key\\u00Fa",
		},
		{
			name:            "String With Multiple Escaped Backslash Sequences And Escaped Quotation",
			input:           "\"\\/\\\\\\\"string\"",
			expectedType:    token.STRING,
			expectedLiteral: "\\/\\\\\\\"string",
		},
		{
			name:            "Illegal String With No Closing Quotation Mark",
			input:           "\"key",
			expectedType:    token.ILLEGAL,
			expectedLiteral: "key",
		},
		{
			name:            "Illegal String With Unescaped Reverse Solidus",
			input:           "\"string\\\"",
			expectedType:    token.ILLEGAL,
			expectedLiteral: "string\\\"",
		},
		{
			name:            "Illegal String With Unescaped Backspace",
			input:           "\"string\b\"",
			expectedType:    token.ILLEGAL,
			expectedLiteral: "string",
		},
		{
			name:            "Illegal String With Unescaped Formfeed",
			input:           "\"string\f\"",
			expectedType:    token.ILLEGAL,
			expectedLiteral: "string",
		},
		{
			name:            "Illegal String With Unescaped Linefeed",
			input:           "\"string\n\"",
			expectedType:    token.ILLEGAL,
			expectedLiteral: "string",
		},
		{
			name:            "Illegal String With Unescaped Carriage Return",
			input:           "\"string\r\"",
			expectedType:    token.ILLEGAL,
			expectedLiteral: "string",
		},
		{
			name:            "Illegal String With Unescaped Horizontal",
			input:           "\"string\t\"",
			expectedType:    token.ILLEGAL,
			expectedLiteral: "string",
		},
		{
			name:            "Illegal String With Vertical Tab",
			input:           "\"string\v\"",
			expectedType:    token.ILLEGAL,
			expectedLiteral: "string",
		},
		{
			name:            "Number",
			input:           "111",
			expectedType:    token.NUMBER,
			expectedLiteral: "111",
		},
		{
			name:            "Decimal Number",
			input:           "111.987",
			expectedType:    token.NUMBER,
			expectedLiteral: "111.987",
		},
		{
			name:            "Negative Number",
			input:           "-123456",
			expectedType:    token.NUMBER,
			expectedLiteral: "-123456",
		},
		{
			name:            "Negative Decimal Number",
			input:           "-0.987",
			expectedType:    token.NUMBER,
			expectedLiteral: "-0.987",
		},
		{
			name:            "Scientific Notation Positive Exponent",
			input:           "2e12",
			expectedType:    token.NUMBER,
			expectedLiteral: "2e12",
		},
		{
			name:            "Scientific Notation Negative Exponent",
			input:           "2e-12",
			expectedType:    token.NUMBER,
			expectedLiteral: "2e-12",
		},
		{
			name:            "Scientific Notation Decimal Mantissa Positive Exponent",
			input:           "0.5678e12",
			expectedType:    token.NUMBER,
			expectedLiteral: "0.5678e12",
		},
		{
			name:            "Scientific Notation Decimal Mantissa Negative Exponent",
			input:           "0.5678e-12",
			expectedType:    token.NUMBER,
			expectedLiteral: "0.5678e-12",
		},
		{
			name:            "Scientific Notation Negative Decimal Mantissa Positive Exponent",
			input:           "-0.5678e+12",
			expectedType:    token.NUMBER,
			expectedLiteral: "-0.5678e+12",
		},
		{
			name:            "Scientific Notation Negative Decimal Mantissa Negative Exponent",
			input:           "-0.5678e-12",
			expectedType:    token.NUMBER,
			expectedLiteral: "-0.5678e-12",
		},
		{
			name:            "Scientific Notation Zero Exponent",
			input:           "5678E0",
			expectedType:    token.NUMBER,
			expectedLiteral: "5678E0",
		},
		{
			name:            "Ignoring Characters After Positive Number",
			input:           "56foo",
			expectedType:    token.NUMBER,
			expectedLiteral: "56",
		},
		{
			name:            "Ignoring Characters After Positive Decimal Number",
			input:           "0.56foo",
			expectedType:    token.NUMBER,
			expectedLiteral: "0.56",
		},
		{
			name:            "Ignoring Characters After Negative Decimal Number",
			input:           "-0.56foo",
			expectedType:    token.NUMBER,
			expectedLiteral: "-0.56",
		},
		{
			name:            "Ignoring Characters After Scientific Notation Positive Exponent",
			input:           "2e012*",
			expectedType:    token.NUMBER,
			expectedLiteral: "2e012",
		},
		{
			name:            "Ignoring Characters After Scientific Notation Negative Exponent",
			input:           "2e-12*",
			expectedType:    token.NUMBER,
			expectedLiteral: "2e-12",
		},
		{
			name:            "Ignoring Characters After Scientific Notation Decimal Mantissa Positive Exponent",
			input:           "0.5678e12*",
			expectedType:    token.NUMBER,
			expectedLiteral: "0.5678e12",
		},
		{
			name:            "Ignoring Characters After Scientific Notation Decimal Mantissa Negative Exponent",
			input:           "0.5678e-12*",
			expectedType:    token.NUMBER,
			expectedLiteral: "0.5678e-12",
		},
		{
			name:            "Ignoring Characters After Scientific Notation Negative Decimal Mantissa Positive Exponent",
			input:           "-0.5678e+12*",
			expectedType:    token.NUMBER,
			expectedLiteral: "-0.5678e+12",
		},
		{
			name:            "Ignoring Characters After Scientific Notation Negative Decimal Mantissa Negative Exponent",
			input:           "-0.5678e-12*",
			expectedType:    token.NUMBER,
			expectedLiteral: "-0.5678e-12",
		},
		{
			name:            "Illegal Negative Number",
			input:           "-056",
			expectedType:    token.ILLEGAL,
			expectedLiteral: "-056",
		},
		{
			name:            "Illegal Negative Number",
			input:           "-f123",
			expectedType:    token.ILLEGAL,
			expectedLiteral: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := mylog.CreateLogger(true)
			l := New(log, string(tt.input))
			tok := l.NextToken()

			assert.Equal(t, tt.expectedType, tok.Type, "tokenType isn't correct")
			assert.Equal(t, tt.expectedLiteral, tok.Literal, "token.Literal isn't correct")
		})
	}
}
