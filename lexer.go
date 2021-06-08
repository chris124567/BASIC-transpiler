package main

import (
	"errors"
	"fmt"
	"unicode"
)

type Lexer struct {
	source      []byte
	currentChar byte
	currentPos  int
}

func NewLexer(source []byte) *Lexer {
	lexer := &Lexer{
		source:      append(source, '\n'),
		currentChar: 0,
		currentPos:  -1,
	}
	lexer.nextChar()
	return lexer
}

func (l *Lexer) nextChar() {
	l.currentPos += 1
	if l.currentPos >= len(l.source) {
		l.currentChar = 0
	} else {
		l.currentChar = l.source[l.currentPos]
	}
}

func (l *Lexer) peek() byte {
	if l.currentPos+1 >= len(l.source) {
		return 0
	}
	return l.source[l.currentPos+1]
}

func (l *Lexer) skipWhitespace() {
	for l.currentChar == ' ' || l.currentChar == '\t' || l.currentChar == '\n' {
		l.nextChar()
	}
}

func (l *Lexer) skipComment() {
	for l.currentChar == '#' {
		for l.currentChar != '\n' {
			l.nextChar()
		}
	}
}

func (l *Lexer) getToken() (*Token, error) {
	l.skipWhitespace()
	l.skipComment()

	var token *Token
	switch l.currentChar {
	// plus
	case '+':
		token = NewTokenByte(l.currentChar, PLUS)
	// minus
	case '-':
		token = NewTokenByte(l.currentChar, MINUS)
	// asterisk
	case '*':
		token = NewTokenByte(l.currentChar, ASTERISK)
	// slash
	case '/':
		token = NewTokenByte(l.currentChar, SLASH)
	// equals
	case '=':
		if l.peek() == '=' {
			lastChar := l.currentChar
			l.nextChar()
			token = NewToken(string([]byte{lastChar, l.currentChar}), EQEQ)
		} else {
			token = NewTokenByte(l.currentChar, EQ)
		}
	// greater than
	case '>':
		if l.peek() == '=' {
			lastChar := l.currentChar
			l.nextChar()
			token = NewToken(string([]byte{lastChar, l.currentChar}), GTEQ)
		} else {
			token = NewTokenByte(l.currentChar, GT)
		}
	// less than
	case '<':
		if l.peek() == '=' {
			lastChar := l.currentChar
			l.nextChar()
			token = NewToken(string([]byte{lastChar, l.currentChar}), LTEQ)
		} else {
			token = NewTokenByte(l.currentChar, LT)
		}
	// not
	case '!':
		if l.peek() == '=' {
			lastChar := l.currentChar
			l.nextChar()
			token = NewToken(string([]byte{lastChar, l.currentChar}), NOTEQ)
		} else {
			return nil, fmt.Errorf("expected !=, got !%c", l.peek())
		}
	// quote
	case '"':
		l.nextChar()
		startPosition := l.currentPos
		for l.currentChar != '"' {
			if l.currentChar == '\r' || l.currentChar == '\n' || l.currentChar == '\t' || l.currentChar == '\\' || l.currentChar == '%' {
				return nil, fmt.Errorf("illegal character in string: %c", l.currentChar)
			}
			l.nextChar()
		}
		token = NewToken(string(l.source[startPosition:l.currentPos]), STRING)
	// newline
	case '\n':
		token = NewTokenByte(l.currentChar, NEWLINE)
	// null
	case 0:
		token = NewTokenByte(l.currentChar, EOF)
	// unknown token
	default:
		if unicode.IsDigit(rune(l.currentChar)) {
			startPosition := l.currentPos
			for unicode.IsDigit(rune(l.peek())) {
				l.nextChar()
			}
			// decimal
			if l.peek() == '.' {
				l.nextChar()
				// must have digit after decimal
				if !unicode.IsDigit(rune(l.peek())) {
					return nil, errors.New("must have digit(s) after decimal")
				}
				for unicode.IsDigit(rune(l.peek())) {
					l.nextChar()
				}
			}
			token = NewToken(string(l.source[startPosition:l.currentPos+1]), NUMBER)
		} else if unicode.IsLetter(rune(l.currentChar)) {
			startPosition := l.currentPos
			for unicode.IsLetter(rune(l.peek())) {
				l.nextChar()
			}
			text := string(l.source[startPosition : l.currentPos+1])
			token = GetKeywordToken(text)
			// if we don't get a token back (it's not a keyword)
			if token == nil {
				token = NewToken(text, IDENT)
			}
		} else {
			return nil, fmt.Errorf("unknown token: %c", l.currentChar)
		}
	}
	l.nextChar()
	return token, nil
}
