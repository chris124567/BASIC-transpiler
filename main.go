package main

import (
	"log"
)

func main() {
	// source := `
	// LET foobar = 123
	// `

	source := `
LET a = 0
WHILE a < 1 REPEAT
    PRINT "Enter number of scores: "
    INPUT a
ENDWHILE

LET b = 0
LET s = 0
PRINT "Enter one value at a time: "
WHILE b < a REPEAT
    INPUT c
    LET s = s + c
    LET b = b + 1
ENDWHILE

PRINT "Average: "
PRINT s / a
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
