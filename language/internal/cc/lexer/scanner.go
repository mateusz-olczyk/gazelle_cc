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
	"bytes"
	"io"
)

func prequalifyToken(data []byte, noMoreData bool) TokenType {
	if len(data) == 0 {
		return TokenType_Incomplete
	}

	switch data[0] {
	case '\t', '\n', '\v', '\f', '\r', ' ':
		return TokenType_Whitespace
	case '/':
		if len(data) > 1 && data[1] == '/' {
			return TokenType_SingleLineComment
		} else if len(data) > 1 && data[1] == '*' {
			return TokenType_MultiLineComment
		}
	case '"':
		return TokenType_StringLiteral
	case 'R':
		if len(data) > 1 && data[1] == '"' {
			return TokenType_RawStringLiteral
		} else if len(data) > 1 || noMoreData {
			return TokenType_Word
		}
	case '(', ')', '[', ']', '{', '}', ',', ';', '<', '>', '=', '!':
		return TokenType_Separator
	default:
		return TokenType_Word
	}

	return TokenType_Incomplete
}

func extractWordToken(data []byte, noMoreData bool) []byte {
	for i := 1; i < len(data); i++ {
		if prequalifyToken(data[i:], noMoreData) != TokenType_Word {
			return data[:i]
		}
	}

	if noMoreData {
		return data
	}

	return nil
}

func extractWhitespaceToken(data []byte, noMoreData bool) []byte {
	for i := 1; i < len(data); i++ {
		if prequalifyToken(data[i:], noMoreData) != TokenType_Whitespace {
			return data[:i]
		}
	}

	if noMoreData {
		return data
	}

	return nil
}

func extractSingleLineCommentToken(data []byte, noMoreData bool) []byte {
	if newlineIndex := bytes.IndexAny(data, "\r\n"); newlineIndex >= 0 {
		return data[:newlineIndex]
	}

	if noMoreData {
		return data
	}

	return nil
}

func extractMultiLineCommentToken(data []byte) []byte {
	if endIndex := bytes.Index(data, []byte("*/")); endIndex >= 0 {
		return data[:endIndex+2]
	}

	return nil
}

func extractStringLiteralToken(data []byte) []byte {
	start := 1
	for {
		relIndex := bytes.IndexByte(data[start:], '"')
		if relIndex < 0 {
			return nil
		}

		absIndex := start + relIndex
		if data[absIndex-1] != '\\' || data[absIndex-2] == '\\' {
			return data[:absIndex+1]
		}

		start = absIndex + 1
	}
}

func extractRawStringLiteralToken(data []byte) []byte {
	startIndex := bytes.IndexByte(data, '(')
	if startIndex < 0 {
		return nil
	}

	delimiter := data[2:startIndex]
	endDelimiter := append([]byte{')'}, delimiter...)
	endIndex := bytes.Index(data, endDelimiter)
	if endIndex < 0 {
		return nil
	}

	return data[:endIndex+len(endDelimiter)]
}

func extractSeparatorToken(data []byte, noMoreData bool) []byte {
	if len(data) == 0 {
		return nil
	}

	switch data[0] {
	case '(', ')', '[', ']', '{', '}', ',', ';':
		return data[:1]
	case '<', '>', '=', '!':
		if len(data) > 1 && data[1] == '=' {
			return data[:2]
		} else if len(data) > 1 || noMoreData {
			return data[:1]
		}
	}

	return nil
}

func extractToken(data []byte, noMoreData bool, t TokenType) []byte {
	switch t {
	case TokenType_Word:
		return extractWordToken(data, noMoreData)
	case TokenType_Whitespace:
		return extractWhitespaceToken(data, noMoreData)
	case TokenType_SingleLineComment:
		return extractSingleLineCommentToken(data, noMoreData)
	case TokenType_MultiLineComment:
		return extractMultiLineCommentToken(data)
	case TokenType_StringLiteral:
		return extractStringLiteralToken(data)
	case TokenType_RawStringLiteral:
		return extractRawStringLiteralToken(data)
	case TokenType_Separator:
		return extractSeparatorToken(data, noMoreData)
	}

	return nil
}

func tokenizer(data []byte, atEOF bool) (advance int, token []byte, err error) {
	token = extractToken(data, atEOF, prequalifyToken(data, atEOF))
	advance = len(token)
	return
}

func newScanner(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	scanner.Split(tokenizer)
	return scanner
}
