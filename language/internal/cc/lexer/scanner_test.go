// Copyright 2025 EngFlow Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lexer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrequalifyToken(t *testing.T) {
	testCases := []struct {
		input      string
		noMoreData bool
		expected   TokenType
	}{
		{
			input:      "",
			noMoreData: true,
			expected:   TokenType_Incomplete,
		},
		{
			input:      " (",
			noMoreData: true,
			expected:   TokenType_Whitespace,
		},
		{
			input:      "    \n",
			noMoreData: false,
			expected:   TokenType_Whitespace,
		},
		{
			input:      "    \n",
			noMoreData: true,
			expected:   TokenType_Whitespace,
		},
		{
			input:      "\"string",
			noMoreData: true,
			expected:   TokenType_StringLiteral,
		},
		{
			input:      "R\"(raw string",
			noMoreData: true,
			expected:   TokenType_RawStringLiteral,
		},
		{
			// 'R' could be the start of a raw string literal, or it could be a word. We need more data to decide.
			input:      "R",
			noMoreData: false,
			expected:   TokenType_Incomplete,
		},
		{
			input:      "R",
			noMoreData: true,
			expected:   TokenType_Word,
		},
		{
			input:      "RR",
			noMoreData: true,
			expected:   TokenType_Word,
		},
		{
			input:      "// single line comment",
			noMoreData: true,
			expected:   TokenType_SingleLineComment,
		},
		{
			input:      "/* multi line comment",
			noMoreData: true,
			expected:   TokenType_MultiLineComment,
		},
		{
			input:      "/",
			noMoreData: false,
			expected:   TokenType_Incomplete,
		},
		{
			input:      "<iostream>",
			noMoreData: true,
			expected:   TokenType_Separator,
		},
		{
			input:      "int main()",
			noMoreData: true,
			expected:   TokenType_Word,
		},
	}

	for _, tc := range testCases {
		result := prequalifyToken([]byte(tc.input), tc.noMoreData)
		assert.Equal(t, tc.expected, result, "Input: %q, NoMoreData: %v", tc.input, tc.noMoreData)
	}
}

func TestExtractWordToken(t *testing.T) {
	testCases := []struct {
		input      string
		noMoreData bool
		expected   string
	}{
		{
			input:      "identifier123;",
			noMoreData: true,
			expected:   "identifier123",
		},
		{
			input:      "identifier123",
			noMoreData: true,
			expected:   "identifier123",
		},
		{
			input:      "identifier123",
			noMoreData: false,
			expected:   "",
		},
	}

	for _, tc := range testCases {
		result := extractWordToken([]byte(tc.input), tc.noMoreData)
		assert.Equal(t, tc.expected, string(result), "Input: %q, NoMoreData: %v", tc.input, tc.noMoreData)
	}
}

func TestExtractWhitespaceToken(t *testing.T) {
	testCases := []struct {
		input      string
		noMoreData bool
		expected   string
	}{
		{
			input:      "   \n\t  identifier",
			noMoreData: true,
			expected:   "   \n\t  ",
		},
		{
			input:      "   \n\t  ",
			noMoreData: true,
			expected:   "   \n\t  ",
		},
		{
			input:      "   \n\t  ",
			noMoreData: false,
			expected:   "",
		},
	}

	for _, tc := range testCases {
		result := extractWhitespaceToken([]byte(tc.input), tc.noMoreData)
		assert.Equal(t, tc.expected, string(result), "Input: %q, NoMoreData: %v", tc.input, tc.noMoreData)
	}
}

func TestExtractSingleLineCommentToken(t *testing.T) {
	testCases := []struct {
		input      string
		noMoreData bool
		expected   string
	}{
		{
			input:      "// This is a single line comment\nint main()",
			noMoreData: true,
			expected:   "// This is a single line comment",
		},
		{
			input:      "// This is a single line comment",
			noMoreData: true,
			expected:   "// This is a single line comment",
		},
		{
			input:      "// This is a single line comment",
			noMoreData: false,
			expected:   "",
		},
	}

	for _, tc := range testCases {
		result := extractSingleLineCommentToken([]byte(tc.input), tc.noMoreData)
		assert.Equal(t, tc.expected, string(result), "Input: %q, NoMoreData: %v", tc.input, tc.noMoreData)
	}
}

func TestExtractMultiLineCommentToken(t *testing.T) {
	testCases := []struct {
		input       string
		noMoreData  bool
		expected    string
		expectedErr string
	}{
		{
			input:      "/* This is a multi line comment */\nint main()",
			noMoreData: true,
			expected:   "/* This is a multi line comment */",
		},
		{
			input:       "/* This is a multi line comment",
			noMoreData:  true,
			expectedErr: "unterminated multi-line comment",
		},
		{
			input:      "/* This is a multi line comment",
			noMoreData: false,
			expected:   "",
		},
	}

	for _, tc := range testCases {
		result, err := extractMultiLineCommentToken([]byte(tc.input), tc.noMoreData)
		assert.Equal(t, tc.expected, string(result), "Input: %q, NoMoreData: %v", tc.input, tc.noMoreData)
		if tc.expectedErr == "" {
			assert.NoError(t, err, "Input: %q, NoMoreData: %v", tc.input, tc.noMoreData)
		} else {
			assert.EqualError(t, err, tc.expectedErr, "Input: %q, NoMoreData: %v", tc.input, tc.noMoreData)
		}
	}
}

func TestExtractStringLiteralToken(t *testing.T) {
	testCases := []struct {
		input       string
		noMoreData  bool
		expected    string
		expectedErr string
	}{
		{
			input:      `""`,
			noMoreData: true,
			expected:   `""`,
		},
		{
			input:      `"\""`,
			noMoreData: true,
			expected:   `"\""`,
		},
		{
			input:      `"This is a string literal"`,
			noMoreData: true,
			expected:   `"This is a string literal"`,
		},
		{
			input:      `"This is a string with an escaped quote: \" inside"`,
			noMoreData: true,
			expected:   `"This is a string with an escaped quote: \" inside"`,
		},
		{
			input:      `"This is an unterminated string literal`,
			noMoreData: false,
			expected:   "",
		},
		{
			input:       `"This is an unterminated string literal`,
			noMoreData:  true,
			expectedErr: "unterminated string literal",
		},
		{
			input:      `"Escaped backslash \\"; "different string"`,
			noMoreData: true,
			expected:   `"Escaped backslash \\"`,
		},
	}

	for _, tc := range testCases {
		result, err := extractStringLiteralToken([]byte(tc.input), tc.noMoreData)
		assert.Equal(t, tc.expected, string(result), "Input: %q, NoMoreData: %v", tc.input, tc.noMoreData)
		if tc.expectedErr == "" {
			assert.NoError(t, err, "Input: %q, NoMoreData: %v", tc.input, tc.noMoreData)
		} else {
			assert.EqualError(t, err, tc.expectedErr, "Input: %q, NoMoreData: %v", tc.input, tc.noMoreData)
		}
	}
}

func TestExtractRawStringLiteralToken(t *testing.T) {
	testCases := []struct {
		input       string
		noMoreData  bool
		expected    string
		expectedErr string
	}{
		{
			input:      `R"()"`,
			noMoreData: true,
			expected:   `R"()"`,
		},
		{
			input:      `R"delim(This is a raw string with a custom delimiter)delim"`,
			noMoreData: true,
			expected:   `R"delim(This is a raw string with a custom delimiter)delim"`,
		},
		{
			input:       `R"(This is an unterminated raw string literal`,
			noMoreData:  true,
			expectedErr: "unterminated raw string literal",
		},
		{
			input:      `R"(This is an unterminated raw string literal`,
			noMoreData: false,
			expected:   "",
		},
		{
			input:       `R"delim(This is an unterminated raw string literal)`,
			noMoreData:  true,
			expectedErr: "unterminated raw string literal",
		},
		{
			input:      `R"delim(This is an unterminated raw string literal)`,
			noMoreData: false,
			expected:   "",
		},
		{
			input:       `R"Missing parenthesis"`,
			noMoreData:  true,
			expectedErr: "missing opening delimiter '(' in raw string literal",
		},
	}

	for _, tc := range testCases {
		result, err := extractRawStringLiteralToken([]byte(tc.input), tc.noMoreData)
		assert.Equal(t, tc.expected, string(result), "Input: %q, NoMoreData: %v", tc.input, tc.noMoreData)
		if tc.expectedErr == "" {
			assert.NoError(t, err, "Input: %q, NoMoreData: %v", tc.input, tc.noMoreData)
		} else {
			assert.EqualError(t, err, tc.expectedErr, "Input: %q, NoMoreData: %v", tc.input, tc.noMoreData)
		}
	}
}

func TestExtractSeparatorToken(t *testing.T) {
	testCases := []struct {
		input      string
		noMoreData bool
		expected   string
	}{
		{
			input:      "(",
			noMoreData: true,
			expected:   "(",
		},
		{
			input:      "<=",
			noMoreData: true,
			expected:   "<=",
		},
		{
			input:      "<",
			noMoreData: true,
			expected:   "<",
		},
		{
			input:      "<",
			noMoreData: false,
			expected:   "",
		},
	}

	for _, tc := range testCases {
		result := extractSeparatorToken([]byte(tc.input), tc.noMoreData)
		assert.Equal(t, tc.expected, string(result), "Input: %q, NoMoreData: %v", tc.input, tc.noMoreData)
	}
}
