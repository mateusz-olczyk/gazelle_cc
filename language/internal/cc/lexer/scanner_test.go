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
			// 'R' could be the start of a raw string literal, or it could be a word. We need more data to decide.
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
