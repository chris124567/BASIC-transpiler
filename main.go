package main

import (
	"log"
)

func main() {
	// source := `
	// LET foobar = 123
	// `

	source := `
PRINT "How many fibonacci numbers do you want?"
INPUT nums

LET a = 0
LET b = 1
WHILE nums > 0 REPEAT
    PRINT a
    LET c = a + b
    LET a = b
    LET b = c
    LET nums = nums - 1
ENDWHILE	
	`

	lexer := NewLexer([]byte(source))
	emitter, err := NewEmitter("output/output.go")
	if err != nil {
		panic(err)
	}
	parser, err := NewParser(lexer, emitter)
	if err != nil {
		panic(err)
	}
	err = parser.program()
	if err != nil {
		panic(err)
	}
	err = emitter.write()
	if err != nil {
		panic(err)
	}
	log.Print("Parsing completed")
}
