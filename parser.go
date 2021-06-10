package main

import (
	"fmt"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var _ = log.Print

type Emitter struct {
	fullPath string
	imports  map[string]bool
	function string
}

func (e *Emitter) emit(code string) {
	e.function += code
}

func (e *Emitter) emitLine(code string) {
	e.function += code + "\n"
}

func (e *Emitter) addImport(pkg string) {
	if e.imports[pkg] {
		return
	}
	e.imports[pkg] = true

}

func (e *Emitter) write() error {
	file, err := os.Create(e.fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	importString := "import ("
	for imp := range e.imports {
		importString += fmt.Sprintf(`"%s";`, imp)
	}
	importString += ");"

	code := strings.Join([]string{"package main\n", importString, "func main() {", e.function, "}"}, "")
	formatted, err := format.Source([]byte(code))
	if err != nil {
		return err
	}

	_, err = file.Write(formatted)
	if err != nil {
		return err
	}
	return nil
}

func NewEmitter(path string) (*Emitter, error) {
	fullPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	return &Emitter{
		fullPath: fullPath,
		imports:  make(map[string]bool),
	}, nil
}

type Parser struct {
	lexer          *Lexer
	emitter        *Emitter
	peekToken      *Token
	currentToken   *Token
	symbols        map[string]bool
	labelsDeclared map[string]bool
	labelsGotoed   map[string]bool
}

func NewParser(lexer *Lexer, emitter *Emitter) *Parser {
	parser := &Parser{
		lexer:          lexer,
		emitter:        emitter,
		symbols:        make(map[string]bool),
		labelsDeclared: make(map[string]bool),
		labelsGotoed:   make(map[string]bool),
	}
	parser.nextToken()
	parser.nextToken()
	return parser
}

func (p *Parser) checkToken(kind int) bool {
	return p.currentToken.kind == kind
}

func (p *Parser) checkPeekToken(kind int) bool {
	return p.peekToken.kind == kind
}

func (p *Parser) match(kind int) {
	if !p.checkToken(kind) {
		log.Panicf("Expected %d, got %d", kind, p.currentToken.kind)
	}
	p.nextToken()
}

func (p *Parser) nextToken() {
	p.currentToken = p.peekToken
	token, err := p.lexer.getToken()
	if err != nil {
		panic(err)
	}
	p.peekToken = token
}

// program ::= {statement}
func (p *Parser) program() {
	// p.nl()

	for !p.checkToken(EOF) {
		p.statement()
	}

	for label := range p.labelsGotoed {
		if !p.labelsDeclared[label] {
			log.Panicf("attempting to GOTO undeclared label: %s", label)
		}
	}
}

func (p *Parser) statement() {
	switch {
	case p.checkToken(PRINT):
		log.Print("STATEMENT-PRINT")

		p.nextToken()
		if p.checkToken(STRING) {
			p.emitter.addImport("fmt")
			p.emitter.emitLine(fmt.Sprintf(`fmt.Println("%s");`, p.currentToken.text))
			p.nextToken()
		} else {
			p.emitter.emit(`fmt.Printf("%v\n",`)

			p.expression()
			p.emitter.emit(");")
		}
	case p.checkToken(IF):
		log.Print("STATEMENT-IF")

		p.nextToken()
		p.emitter.emit("if ")

		p.comparison()
		p.emitter.emitLine("{")

		p.match(THEN)
		for !p.checkToken(ENDIF) && !p.checkToken(ELSE) {
			p.statement()
		}
		// no else
		if p.checkToken(ENDIF) {
			p.emitter.emitLine("}")
		} else if p.checkToken(ELSE) {
			p.nextToken()
			p.emitter.emitLine("\n}else {")
			for !p.checkToken(ENDIF) {
				p.statement()
			}
			p.match(ENDIF)
			p.emitter.emitLine("}")

		} else {
			log.Panicf("expected endif or else but got: %v", p.currentToken)
		}
	case p.checkToken(FOR):
		log.Print("STATEMENT-FOR")
		// todo: FOR i := 0; i < x; i++ support
		fallthrough
	case p.checkToken(WHILE):
		log.Print("STATEMENT-WHILE")

		isFor := p.checkToken(FOR)
		p.nextToken()
		p.emitter.emit("for ")

		p.comparison()
		p.emitter.emitLine("{")

		p.match(REPEAT)
		p.nl()
		for (!isFor && !p.checkToken(ENDWHILE)) || (isFor && !p.checkToken(ENDFOR)) {
			p.statement()
		}

		p.emitter.emitLine("}")
		if p.checkToken(ENDWHILE) || p.checkToken(ENDFOR) {
			p.nextToken()
		} else {
			log.Panicf("expected endwhile or endfor but got: %v", p.currentToken)
		}
	case p.checkToken(LABEL):
		log.Print("STATEMENT-LABEL")

		p.nextToken()
		if p.labelsDeclared[p.currentToken.text] {
			log.Panicf("label %s already exists", p.currentToken.text)
		}
		p.labelsDeclared[p.currentToken.text] = true

		p.emitter.emitLine(p.currentToken.text + ":")
		p.match(IDENT)
	case p.checkToken(GOTO):
		log.Print("STATEMENT-GOTO")

		p.nextToken()
		p.labelsGotoed[p.currentToken.text] = true
		p.emitter.emitLine("goto " + p.currentToken.text)
		p.match(IDENT)
	case p.checkToken(LET):
		log.Print("STATEMENT-LET")

		p.nextToken()

		seen := p.symbols[p.currentToken.text]
		if !seen {
			p.symbols[p.currentToken.text] = true
		}
		name := p.currentToken.text

		p.match(IDENT)

		p.match(EQ)

		if seen {
			p.emitter.emit(name + " = ")
		} else {
			p.emitter.emit(name + " := ")
		}

		p.expression()
		p.emitter.emit(";")

	case p.checkToken(INPUT):
		log.Print("STATEMENT-INPUT")

		p.emitter.addImport("fmt")

		p.nextToken()

		name := p.currentToken.text
		seen := p.symbols[name]
		if !seen {
			p.symbols[name] = true
		}

		p.match(IDENT)

		if seen {
			p.emitter.emitLine(fmt.Sprintf(`fmt.Scanf("%%f", &%s);`, name))
		} else {
			p.emitter.emitLine(fmt.Sprintf(`var %s float64; fmt.Scanf("%%f", &%s);`, name, name))

		}
	default:
		log.Panicf("Invalid statement at %s (%s)", p.currentToken.text, TOKENS[p.currentToken.kind])
	}
	p.nl()
}

// expression ::= term {( "-" | "+" ) term}
func (p *Parser) expression() {
	p.term()

	for p.checkToken(PLUS) || p.checkToken(MINUS) {
		p.emitter.emit(p.currentToken.text)
		p.nextToken()
		p.term()
	}
}

// comparison ::= expression (("==" | "!=" | ">" | ">=" | "<" | "<=") expression)+
func (p *Parser) comparison() {
	p.expression()
	if p.currentToken.isComparisonOperator() {
		p.emitter.emit(p.currentToken.text)
		p.nextToken()
		p.expression()
	} else {
		log.Panicf("Expected comparison operator at: %s", p.currentToken.text)
	}

	for p.currentToken.isComparisonOperator() {
		p.emitter.emit(p.currentToken.text)
		p.nextToken()
		p.expression()
	}

}

// term ::= unary {( "/" | "*" ) unary}
func (p *Parser) term() {
	p.unary()
	for p.checkToken(ASTERISK) || p.checkToken(SLASH) {
		p.emitter.emit(p.currentToken.text)
		p.nextToken()
		p.unary()
	}
}

// unary ::= ["+" | "-"] primary
func (p *Parser) unary() {
	if p.checkToken(PLUS) || p.checkToken(MINUS) {
		p.emitter.emit(p.currentToken.text)
		p.nextToken()
	}
	p.primary()
}

func (p *Parser) primary() {
	if p.checkToken(NUMBER) {
		p.emitter.emit("float64(" + p.currentToken.text + ")")
		p.nextToken()
	} else if p.checkToken(STRING) {
		p.emitter.emit(`"` + p.currentToken.text + `"`)
		p.nextToken()
	} else if p.checkToken(IDENT) {
		if !p.symbols[p.currentToken.text] {
			log.Panicf("referencing variable before assignment: %s", p.currentToken.text)
		}
		p.emitter.emit(p.currentToken.text)
		p.nextToken()
	} else {
		log.Panicf("unexpected token at %s", p.currentToken.text)
	}
}

// nl ::= '\n'+
func (p *Parser) nl() {
	for p.checkToken(NEWLINE) {
		p.nextToken()
	}
}
