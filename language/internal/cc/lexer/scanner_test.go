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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrequalifyToken(t *testing.T) {
	testCases := []struct {
		input    chunk
		expected TokenType
	}{
		{
			input:    chunk{data: []byte("")},
			expected: TokenType_Incomplete,
		},
		{
			input:    chunk{data: []byte(" (")},
			expected: TokenType_Whitespace,
		},
		{
			input:    chunk{data: []byte("    \n")},
			expected: TokenType_Whitespace,
		},
		{
			input:    chunk{data: []byte("    \n")},
			expected: TokenType_Whitespace,
		},
		{
			input:    chunk{data: []byte(`"string`)},
			expected: TokenType_StringLiteral,
		},
		{
			input:    chunk{data: []byte(`R"(raw string`)},
			expected: TokenType_RawStringLiteral,
		},
		{
			// 'R' could be the start of a raw string literal, or it could be a word. We need more data to decide.
			input:    chunk{data: []byte("R"), complete: false},
			expected: TokenType_Incomplete,
		},
		{
			input:    chunk{data: []byte("R"), complete: true},
			expected: TokenType_Word,
		},
		{
			input:    chunk{data: []byte("RR")},
			expected: TokenType_Word,
		},
		{
			input:    chunk{data: []byte("// single line comment")},
			expected: TokenType_SingleLineComment,
		},
		{
			input:    chunk{data: []byte("/* multi line comment")},
			expected: TokenType_MultiLineComment,
		},
		{
			input:    chunk{data: []byte("/"), complete: false},
			expected: TokenType_Incomplete,
		},
		{
			input:    chunk{data: []byte("/"), complete: true},
			expected: TokenType_Word,
		},
		{
			input:    chunk{data: []byte("/ 5")},
			expected: TokenType_Word,
		},
		{
			input:    chunk{data: []byte("<iostream>")},
			expected: TokenType_Separator,
		},
		{
			input:    chunk{data: []byte("int main()")},
			expected: TokenType_Word,
		},
		{
			input:    chunk{data: []byte("\\\r")},
			expected: TokenType_ContinueLine,
		},
	}

	for _, tc := range testCases {
		result := prequalifyToken(tc.input)
		assert.Equal(t, tc.expected, result, "Input: %v", tc.input)
	}
}

func TestExtractWordToken(t *testing.T) {
	testCases := []struct {
		input    chunk
		expected []byte
	}{
		{
			input:    chunk{data: []byte("identifier123;"), complete: true},
			expected: []byte("identifier123"),
		},
		{
			input:    chunk{data: []byte("identifier123"), complete: true},
			expected: []byte("identifier123"),
		},
		{
			input:    chunk{data: []byte("identifier123"), complete: false},
			expected: nil,
		},
	}

	for _, tc := range testCases {
		result := extractWordToken(tc.input)
		assert.Equal(t, tc.expected, result, "Input: %v", tc.input)
	}
}

func TestExtractWhitespaceToken(t *testing.T) {
	testCases := []struct {
		input    chunk
		expected []byte
	}{
		{
			input:    chunk{data: []byte("   \n\t  identifier"), complete: true},
			expected: []byte("   \n\t  "),
		},
		{
			input:    chunk{data: []byte("   \n\t  "), complete: true},
			expected: []byte("   \n\t  "),
		},
		{
			input:    chunk{data: []byte("   \n\t  "), complete: false},
			expected: nil,
		},
	}

	for _, tc := range testCases {
		result := extractWhitespaceToken(tc.input)
		assert.Equal(t, tc.expected, result, "Input: %v", tc.input)
	}
}

func TestExtractContinueLineToken(t *testing.T) {
	testCases := []struct {
		input         chunk
		expectedOk    []byte
		expectedError error
	}{
		{
			input:      chunk{data: []byte("\\\r\n")},
			expectedOk: []byte("\\\r\n"),
		},
		{
			input:      chunk{data: []byte("\\\n")},
			expectedOk: []byte("\\\n"),
		},
		{
			input:      chunk{data: []byte("\\\r"), complete: false},
			expectedOk: nil,
		},
		{
			input:         chunk{data: []byte("\\\r"), complete: true},
			expectedError: ErrContinueLineInvalid,
		},
		{
			input:      chunk{data: []byte("\\"), complete: false},
			expectedOk: nil,
		},
		{
			input:         chunk{data: []byte("\\"), complete: true},
			expectedError: ErrContinueLineInvalid,
		},
		{
			input:         chunk{data: []byte("\\ some text"), complete: true},
			expectedError: ErrContinueLineInvalid,
		},
	}

	for _, tc := range testCases {
		result, err := extractContinueLineToken(tc.input)
		assert.Equal(t, tc.expectedOk, result, "Input: %v", tc.input)
		assert.Equal(t, tc.expectedError, err, "Input: %v", tc.input)
	}
}

func TestExtractSingleLineCommentToken(t *testing.T) {
	testCases := []struct {
		input    chunk
		expected []byte
	}{
		{
			input:    chunk{data: []byte("// This is a single line comment\nint main()"), complete: true},
			expected: []byte("// This is a single line comment"),
		},
		{
			input:    chunk{data: []byte("// This is a single line comment"), complete: true},
			expected: []byte("// This is a single line comment"),
		},
		{
			input:    chunk{data: []byte("// This is a single line comment"), complete: false},
			expected: nil,
		},
	}

	for _, tc := range testCases {
		result := extractSingleLineCommentToken(tc.input)
		assert.Equal(t, tc.expected, result, "Input: %v", tc.input)
	}
}

func TestExtractMultiLineCommentToken(t *testing.T) {
	testCases := []struct {
		input         chunk
		expectedOk    []byte
		expectedError error
	}{
		{
			input:      chunk{data: []byte("/* This is a multi line comment */\nint main()"), complete: true},
			expectedOk: []byte("/* This is a multi line comment */"),
		},
		{
			input:         chunk{data: []byte("/* This is a multi line comment"), complete: true},
			expectedError: ErrMultiLineCommentUnterminated,
		},
		{
			input:      chunk{data: []byte("/* This is a multi line comment"), complete: false},
			expectedOk: nil,
		},
	}

	for _, tc := range testCases {
		result, err := extractMultiLineCommentToken(tc.input)
		assert.Equal(t, tc.expectedOk, result, "Input: %v", tc.input)
		assert.Equal(t, tc.expectedError, err, "Input: %v", tc.input)
	}
}

func TestExtractStringLiteralToken(t *testing.T) {
	testCases := []struct {
		input         chunk
		expectedOk    []byte
		expectedError error
	}{
		{
			input:      chunk{data: []byte(`""`), complete: true},
			expectedOk: []byte(`""`),
		},
		{
			input:      chunk{data: []byte(`"\""`), complete: true},
			expectedOk: []byte(`"\""`),
		},
		{
			input:      chunk{data: []byte(`"This is a string literal"`), complete: true},
			expectedOk: []byte(`"This is a string literal"`),
		},
		{
			input:      chunk{data: []byte(`"This is a string with an escaped quote: \" inside"`), complete: true},
			expectedOk: []byte(`"This is a string with an escaped quote: \" inside"`),
		},
		{
			input:      chunk{data: []byte(`"This is an unterminated string literal`), complete: false},
			expectedOk: nil,
		},
		{
			input:         chunk{data: []byte(`"This is an unterminated string literal`), complete: true},
			expectedError: ErrStringLiteralUnterminated,
		},
		{
			input:      chunk{data: []byte(`"Escaped backslash \\"; "different string"`), complete: true},
			expectedOk: []byte(`"Escaped backslash \\"`),
		},
	}

	for _, tc := range testCases {
		result, err := extractStringLiteralToken(tc.input)
		assert.Equal(t, tc.expectedOk, result, "Input: %v", tc.input)
		assert.Equal(t, tc.expectedError, err, "Input: %v", tc.input)
	}
}

func TestExtractRawStringLiteralToken(t *testing.T) {
	testCases := []struct {
		input         chunk
		expectedOk    []byte
		expectedError error
	}{
		{
			input:      chunk{data: []byte(`R"()"`), complete: true},
			expectedOk: []byte(`R"()"`),
		},
		{
			input:      chunk{data: []byte(`R"delim(This is a raw string with a custom delimiter)delim"`), complete: true},
			expectedOk: []byte(`R"delim(This is a raw string with a custom delimiter)delim"`),
		},
		{
			input:         chunk{data: []byte(`R"(This is an unterminated raw string literal`), complete: true},
			expectedError: ErrRawStringLiteralUnterminated,
		},
		{
			input:      chunk{data: []byte(`R"(This is an unterminated raw string literal`), complete: false},
			expectedOk: nil,
		},
		{
			input:         chunk{data: []byte(`R"delim(This is an unterminated raw string literal)`), complete: true},
			expectedError: ErrRawStringLiteralUnterminated,
		},
		{
			input:      chunk{data: []byte(`R"delim(This is an unterminated raw string literal)`), complete: false},
			expectedOk: nil,
		},
		{
			input:         chunk{data: []byte(`R"Missing parenthesis"`), complete: true},
			expectedError: ErrRawStringLiteralMissingOpeningDelimiter,
		},
	}

	for _, tc := range testCases {
		result, err := extractRawStringLiteralToken(tc.input)
		assert.Equal(t, tc.expectedOk, result, "Input: %v", tc.input)
		assert.Equal(t, tc.expectedError, err, "Input: %v", tc.input)
	}
}

func TestExtractSeparatorToken(t *testing.T) {
	testCases := []struct {
		input    chunk
		expected []byte
	}{
		{
			input:    chunk{data: []byte("("), complete: true},
			expected: []byte("("),
		},
		{
			input:    chunk{data: []byte("<="), complete: true},
			expected: []byte("<="),
		},
		{
			input:    chunk{data: []byte("<"), complete: true},
			expected: []byte("<"),
		},
		{
			input:    chunk{data: []byte("<"), complete: false},
			expected: nil,
		},
	}

	for _, tc := range testCases {
		result := extractSeparatorToken(tc.input)
		assert.Equal(t, tc.expected, result, "Input: %v", tc.input)
	}
}

func readAllTokens(source string) (tokens []string, err error) {
	scanner := newScanner(strings.NewReader(source))
	tokens = make([]string, 0)
	for scanner.Scan() {
		tokens = append(tokens, scanner.Text())
	}

	err = scanner.Err()
	return
}

func TestScanner(t *testing.T) {
	testCases := []struct {
		input          string
		expectedTokens []string
		expectedError  error
	}{
		{
			input:          "int main() { return 0; }",
			expectedTokens: []string{"int", " ", "main", "(", ")", " ", "{", " ", "return", " ", "0", ";", " ", "}"},
		},
		{
			input:          "/* int main() { return 0; } */\n\tint main() { return 0; }",
			expectedTokens: []string{"/* int main() { return 0; } */", "\n\t", "int", " ", "main", "(", ")", " ", "{", " ", "return", " ", "0", ";", " ", "}"},
		},
		{
			input:          "// int main() { return 0; }\nint main() { return 0; }",
			expectedTokens: []string{"// int main() { return 0; }", "\n", "int", " ", "main", "(", ")", " ", "{", " ", "return", " ", "0", ";", " ", "}"},
		},
		{
			input:          "#define FAVOURITE_LETTER R",
			expectedTokens: []string{"#", "define", " ", "FAVOURITE_LETTER", " ", "R"},
		},
		{
			input:          "int main() { /* unterminated comment\n return 0; }",
			expectedTokens: []string{"int", " ", "main", "(", ")", " ", "{", " "},
			expectedError:  ErrMultiLineCommentUnterminated,
		},
		{
			input:          `const char *raw_string = R"delim( #include <iostream> )delim";`,
			expectedTokens: []string{"const", " ", "char", " ", "*", "raw_string", " ", "=", " ", `R"delim( #include <iostream> )delim"`, ";"},
		},
	}

	for _, tc := range testCases {
		tokens, err := readAllTokens(tc.input)
		assert.Equal(t, tc.expectedTokens, tokens, "Input: %q", tc.input)
		assert.Equal(t, tc.expectedError, err, "Input: %q", tc.input)
	}
}
