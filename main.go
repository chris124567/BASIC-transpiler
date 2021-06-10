package main

import (
	"log"
)

func main() {
	// source := `
	// LET foobar = 123
	// `

	source := `

LET welcomeMessage = "Fibonacci Series Program"
PRINT welcomeMessage

PRINT "How many fibonacci numbers do you want?"
INPUT nums
IF nums >= 15 THEN
	PRINT "large number inputted"
ELSE
	PRINT "small number inputted"
ENDIF

LET a = 0
LET b = 1
WHILE nums > 0 REPEAT
    PRINT a
    LET c = a + b
    LET a = b
    LET b = c
    LET nums = nums - 1
ENDWHILE

LET i = 0
FOR i < 10 REPEAT
	PRINT i
	LET i = i + 1
ENDFOR

`

	lexer := NewLexer([]byte(source))
	emitter, err := NewEmitter("output/output.go")
	if err != nil {
		panic(err)
	}
	parser := NewParser(lexer, emitter)
	if err != nil {
		panic(err)
	}
	parser.program()
	err = emitter.write()
	if err != nil {
		panic(err)
	}
	log.Print("Parsing completed")
}
