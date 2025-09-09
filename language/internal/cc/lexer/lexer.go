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
	"bufio"
	"io"
)

type Lexer interface {
	Read() (Token, error)
}

type lexer struct {
	scanner *bufio.Scanner
	cursor  Cursor
}

func NewLexer(r io.Reader) Lexer {
	return &lexer{
		scanner: newScanner(r),
	}
}

func (l *lexer) Read() (token Token, err error) {
	if l.scanner.Scan() {
		content := l.scanner.Text()
		token = Token{
			Type:     prequalifyToken(chunk{data: []byte(content), complete: true}),
			Location: l.cursor,
			Content:  content,
		}
		l.cursor = l.cursor.advanceBy(content)
	} else {
		err = l.scanner.Err()
		if err == nil {
			err = io.EOF
		}
	}

	return
}
