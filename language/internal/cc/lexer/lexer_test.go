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
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const VALID_SOURCE = `#include <iostream>
#include <string>

#define SQUARE(x) \
    ((x) * (x))

namespace {

const std::string STR{"#include <fmt/core.h>"};
const std::wstring WSTR{L"Hello, world! 😎😎😎"}; // This comment starts at line 9, column 47

const std::string RAW_STR{R"delim(
This is a raw string.
)delim"};

}

/*
    Multi-line comment
*/

int main() {
    return 0;
}
`

const UNTERMINATED_MULTILINE_COMMENT = "const int expr{42 / 7}; /* unterminated comment..."

func readAll(l Lexer) ([]Token, error) {
	var tokens []Token
	for {
		token, err := l.Read()
		if err != nil {
			return tokens, err
		}
		tokens = append(tokens, token)
	}
}

func TestLexer(t *testing.T) {
	testCases := []struct {
		testCaseName   string
		input          string
		expectedTokens []Token
		expectedError  error
	}{
		{
			testCaseName: "VALID_SOURCE",
			input:        VALID_SOURCE,
			expectedTokens: []Token{
				{Type: TokenType_Separator, Location: Cursor{Line: 0, Column: 0}, Content: "#"},
				{Type: TokenType_Word, Location: Cursor{Line: 0, Column: 1}, Content: "include"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 0, Column: 8}, Content: " "},
				{Type: TokenType_Separator, Location: Cursor{Line: 0, Column: 9}, Content: "<"},
				{Type: TokenType_Word, Location: Cursor{Line: 0, Column: 10}, Content: "iostream"},
				{Type: TokenType_Separator, Location: Cursor{Line: 0, Column: 18}, Content: ">"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 0, Column: 19}, Content: "\n"},
				{Type: TokenType_Separator, Location: Cursor{Line: 1, Column: 0}, Content: "#"},
				{Type: TokenType_Word, Location: Cursor{Line: 1, Column: 1}, Content: "include"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 1, Column: 8}, Content: " "},
				{Type: TokenType_Separator, Location: Cursor{Line: 1, Column: 9}, Content: "<"},
				{Type: TokenType_Word, Location: Cursor{Line: 1, Column: 10}, Content: "string"},
				{Type: TokenType_Separator, Location: Cursor{Line: 1, Column: 16}, Content: ">"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 1, Column: 17}, Content: "\n\n"},
				{Type: TokenType_Separator, Location: Cursor{Line: 3, Column: 0}, Content: "#"},
				{Type: TokenType_Word, Location: Cursor{Line: 3, Column: 1}, Content: "define"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 3, Column: 7}, Content: " "},
				{Type: TokenType_Word, Location: Cursor{Line: 3, Column: 8}, Content: "SQUARE"},
				{Type: TokenType_Separator, Location: Cursor{Line: 3, Column: 14}, Content: "("},
				{Type: TokenType_Word, Location: Cursor{Line: 3, Column: 15}, Content: "x"},
				{Type: TokenType_Separator, Location: Cursor{Line: 3, Column: 16}, Content: ")"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 3, Column: 17}, Content: " "},
				{Type: TokenType_ContinueLine, Location: Cursor{Line: 3, Column: 18}, Content: "\\\n"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 4, Column: 0}, Content: "    "},
				{Type: TokenType_Separator, Location: Cursor{Line: 4, Column: 4}, Content: "("},
				{Type: TokenType_Separator, Location: Cursor{Line: 4, Column: 5}, Content: "("},
				{Type: TokenType_Word, Location: Cursor{Line: 4, Column: 6}, Content: "x"},
				{Type: TokenType_Separator, Location: Cursor{Line: 4, Column: 7}, Content: ")"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 4, Column: 8}, Content: " "},
				{Type: TokenType_Separator, Location: Cursor{Line: 4, Column: 9}, Content: "*"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 4, Column: 10}, Content: " "},
				{Type: TokenType_Separator, Location: Cursor{Line: 4, Column: 11}, Content: "("},
				{Type: TokenType_Word, Location: Cursor{Line: 4, Column: 12}, Content: "x"},
				{Type: TokenType_Separator, Location: Cursor{Line: 4, Column: 13}, Content: ")"},
				{Type: TokenType_Separator, Location: Cursor{Line: 4, Column: 14}, Content: ")"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 4, Column: 15}, Content: "\n\n"},
				{Type: TokenType_Word, Location: Cursor{Line: 6, Column: 0}, Content: "namespace"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 6, Column: 9}, Content: " "},
				{Type: TokenType_Separator, Location: Cursor{Line: 6, Column: 10}, Content: "{"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 6, Column: 11}, Content: "\n\n"},
				{Type: TokenType_Word, Location: Cursor{Line: 8, Column: 0}, Content: "const"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 8, Column: 5}, Content: " "},
				{Type: TokenType_Word, Location: Cursor{Line: 8, Column: 6}, Content: "std::string"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 8, Column: 17}, Content: " "},
				{Type: TokenType_Word, Location: Cursor{Line: 8, Column: 18}, Content: "STR"},
				{Type: TokenType_Separator, Location: Cursor{Line: 8, Column: 21}, Content: "{"},
				{Type: TokenType_StringLiteral, Location: Cursor{Line: 8, Column: 22}, Content: `"#include <fmt/core.h>"`},
				{Type: TokenType_Separator, Location: Cursor{Line: 8, Column: 45}, Content: "}"},
				{Type: TokenType_Separator, Location: Cursor{Line: 8, Column: 46}, Content: ";"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 8, Column: 47}, Content: "\n"},
				{Type: TokenType_Word, Location: Cursor{Line: 9, Column: 0}, Content: "const"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 9, Column: 5}, Content: " "},
				{Type: TokenType_Word, Location: Cursor{Line: 9, Column: 6}, Content: "std::wstring"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 9, Column: 18}, Content: " "},
				{Type: TokenType_Word, Location: Cursor{Line: 9, Column: 19}, Content: "WSTR"},
				{Type: TokenType_Separator, Location: Cursor{Line: 9, Column: 23}, Content: "{"},
				{Type: TokenType_Word, Location: Cursor{Line: 9, Column: 24}, Content: "L"},
				{Type: TokenType_StringLiteral, Location: Cursor{Line: 9, Column: 25}, Content: `"Hello, world! 😎😎😎"`},
				{Type: TokenType_Separator, Location: Cursor{Line: 9, Column: 44}, Content: "}"},
				{Type: TokenType_Separator, Location: Cursor{Line: 9, Column: 45}, Content: ";"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 9, Column: 46}, Content: " "},
				{Type: TokenType_SingleLineComment, Location: Cursor{Line: 9, Column: 47}, Content: "// This comment starts at line 9, column 47"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 9, Column: 90}, Content: "\n\n"},
				{Type: TokenType_Word, Location: Cursor{Line: 11, Column: 0}, Content: "const"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 11, Column: 5}, Content: " "},
				{Type: TokenType_Word, Location: Cursor{Line: 11, Column: 6}, Content: "std::string"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 11, Column: 17}, Content: " "},
				{Type: TokenType_Word, Location: Cursor{Line: 11, Column: 18}, Content: "RAW_STR"},
				{Type: TokenType_Separator, Location: Cursor{Line: 11, Column: 25}, Content: "{"},
				{Type: TokenType_RawStringLiteral, Location: Cursor{Line: 11, Column: 26}, Content: "R\"delim(\nThis is a raw string.\n)delim\""},
				{Type: TokenType_Separator, Location: Cursor{Line: 13, Column: 7}, Content: "}"},
				{Type: TokenType_Separator, Location: Cursor{Line: 13, Column: 8}, Content: ";"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 13, Column: 9}, Content: "\n\n"},
				{Type: TokenType_Separator, Location: Cursor{Line: 15, Column: 0}, Content: "}"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 15, Column: 1}, Content: "\n\n"},
				{Type: TokenType_MultiLineComment, Location: Cursor{Line: 17, Column: 0}, Content: "/*\n    Multi-line comment\n*/"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 19, Column: 2}, Content: "\n\n"},
				{Type: TokenType_Word, Location: Cursor{Line: 21, Column: 0}, Content: "int"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 21, Column: 3}, Content: " "},
				{Type: TokenType_Word, Location: Cursor{Line: 21, Column: 4}, Content: "main"},
				{Type: TokenType_Separator, Location: Cursor{Line: 21, Column: 8}, Content: "("},
				{Type: TokenType_Separator, Location: Cursor{Line: 21, Column: 9}, Content: ")"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 21, Column: 10}, Content: " "},
				{Type: TokenType_Separator, Location: Cursor{Line: 21, Column: 11}, Content: "{"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 21, Column: 12}, Content: "\n    "},
				{Type: TokenType_Word, Location: Cursor{Line: 22, Column: 4}, Content: "return"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 22, Column: 10}, Content: " "},
				{Type: TokenType_Word, Location: Cursor{Line: 22, Column: 11}, Content: "0"},
				{Type: TokenType_Separator, Location: Cursor{Line: 22, Column: 12}, Content: ";"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 22, Column: 13}, Content: "\n"},
				{Type: TokenType_Separator, Location: Cursor{Line: 23, Column: 0}, Content: "}"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 23, Column: 1}, Content: "\n"},
			},
			expectedError: io.EOF,
		},
		{
			testCaseName: "UNTERMINATED_MULTILINE_COMMENT",
			input:        UNTERMINATED_MULTILINE_COMMENT,
			expectedTokens: []Token{
				{Type: TokenType_Word, Location: Cursor{Line: 0, Column: 0}, Content: "const"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 0, Column: 5}, Content: " "},
				{Type: TokenType_Word, Location: Cursor{Line: 0, Column: 6}, Content: "int"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 0, Column: 9}, Content: " "},
				{Type: TokenType_Word, Location: Cursor{Line: 0, Column: 10}, Content: "expr"},
				{Type: TokenType_Separator, Location: Cursor{Line: 0, Column: 14}, Content: "{"},
				{Type: TokenType_Word, Location: Cursor{Line: 0, Column: 15}, Content: "42"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 0, Column: 17}, Content: " "},
				{Type: TokenType_Word, Location: Cursor{Line: 0, Column: 18}, Content: "/"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 0, Column: 19}, Content: " "},
				{Type: TokenType_Word, Location: Cursor{Line: 0, Column: 20}, Content: "7"},
				{Type: TokenType_Separator, Location: Cursor{Line: 0, Column: 21}, Content: "}"},
				{Type: TokenType_Separator, Location: Cursor{Line: 0, Column: 22}, Content: ";"},
				{Type: TokenType_Whitespace, Location: Cursor{Line: 0, Column: 23}, Content: " "},
			},
			expectedError: ErrMultiLineCommentUnterminated,
		},
	}

	for _, tc := range testCases {
		tokens, err := readAll(NewLexer(strings.NewReader(tc.input)))
		assert.Equal(t, tc.expectedTokens, tokens, "test case: %s", tc.testCaseName)
		assert.Equal(t, tc.expectedError, err, "test case: %s", tc.testCaseName)
	}
}
