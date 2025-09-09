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
	"errors"
	"fmt"
	"strings"
)

type TokenType int

const (
	TokenType_Incomplete TokenType = iota
	TokenType_Word
	TokenType_Whitespace
	TokenType_ContinueLine
	TokenType_SingleLineComment
	TokenType_MultiLineComment
	TokenType_StringLiteral
	TokenType_RawStringLiteral
	TokenType_Separator
)

var (
	ErrContinueLineInvalid                     = errors.New("missing newline character after line continuation backslash")
	ErrMultiLineCommentUnterminated            = errors.New("unterminated multi-line comment")
	ErrRawStringLiteralMissingOpeningDelimiter = errors.New("missing opening delimiter '(' in raw string literal")
	ErrRawStringLiteralUnterminated            = errors.New("unterminated raw string literal")
	ErrStringLiteralUnterminated               = errors.New("unterminated string literal")
)

// position in the source code, Line and Column are 0-based
type Cursor struct {
	Line, Column uint
}

func (c Cursor) String() string {
	return fmt.Sprintf("%d:%d", c.Line+1, c.Column+1)
}

func (c Cursor) advanceBy(s string) Cursor {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	tailBegin := 1 + strings.LastIndex(s, "\n")
	tailLength := uint(len([]rune(s[tailBegin:])))

	if newlinesCount := uint(strings.Count(s, "\n")); newlinesCount == 0 {
		c.Column += tailLength
	} else {
		c.Line += newlinesCount
		c.Column = tailLength
	}

	return c
}

type Token struct {
	Type     TokenType
	Location Cursor
	Content  string
}
