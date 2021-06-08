package main

const (
	EOF = iota - 1
	NEWLINE
	NUMBER
	IDENT
	STRING
	// Keywords
	LABEL
	GOTO
	PRINT
	INPUT
	LET
	IF
	THEN
	ENDIF
	WHILE
	REPEAT
	ENDWHILE
	// Operators
	EQ
	PLUS
	MINUS
	ASTERISK
	SLASH
	EQEQ
	NOTEQ
	LT
	LTEQ
	GT
	GTEQ
)

// printable
var TOKENS = map[int]string{
	NUMBER:   "NUMBER",
	STRING:   "STRING",
	IDENT:    "IDENT",
	LABEL:    "LABEL",
	GOTO:     "GOTO",
	PRINT:    "PRINT",
	INPUT:    "INPUT",
	LET:      "LET",
	IF:       "IF",
	THEN:     "THEN",
	ENDIF:    "ENDIF",
	WHILE:    "WHILE",
	REPEAT:   "REPEAT",
	ENDWHILE: "ENDWHILE",
	EQ:       "EQ",
	PLUS:     "+",
	MINUS:    "-",
	ASTERISK: "*",
	SLASH:    "/",
	EQEQ:     "==",
	NOTEQ:    "!=",
	LT:       "<",
	LTEQ:     "<=",
	GT:       ">",
	GTEQ:     ">=",
}

type Token struct {
	kind int
	text string
}

func (t *Token) isComparisonOperator() bool {
	return t.kind == EQEQ || t.kind == NOTEQ || t.kind == LT || t.kind == LTEQ || t.kind == GT || t.kind == GTEQ
}

func NewToken(text string, kind int) *Token {
	return &Token{
		kind: kind,
		text: text,
	}
}

func NewTokenByte(text byte, kind int) *Token {
	return NewToken(string(text), kind)
}

func GetKeywordToken(text string) *Token {
	for number, keywordText := range TOKENS {
		if text == keywordText {
			return NewToken(text, number)
		}
	}
	return nil
}
