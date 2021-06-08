package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

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

	toWrite := []string{"package main\n", importString, "func main() {", e.function, "}"}

	for _, w := range toWrite {
		_, err = file.WriteString(w)
		if err != nil {
			return err
		}
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

func NewParser(lexer *Lexer, emitter *Emitter) (*Parser, error) {
	parser := &Parser{
		lexer:          lexer,
		emitter:        emitter,
		symbols:        make(map[string]bool),
		labelsDeclared: make(map[string]bool),
		labelsGotoed:   make(map[string]bool),
	}
	err := parser.nextToken()
	if err != nil {
		return nil, err
	}
	err = parser.nextToken()
	if err != nil {
		return nil, err
	}
	return parser, nil
}

func (p *Parser) checkToken(kind int) bool {
	return p.currentToken.kind == kind
}

func (p *Parser) checkPeekToken(kind int) bool {
	return p.peekToken.kind == kind
}

func (p *Parser) match(kind int) error {
	if !p.checkToken(kind) {
		return fmt.Errorf("Expected %d, got %d", kind, p.currentToken.kind)
	}
	return p.nextToken()
}

func (p *Parser) nextToken() error {
	p.currentToken = p.peekToken
	token, err := p.lexer.getToken()
	if err != nil {
		return err
	}
	p.peekToken = token
	return nil
}

// program ::= {statement}
func (p *Parser) program() error {
	log.Print("PROGRAM")

	err := p.nl()
	if err != nil {
		return err
	}

	for !p.checkToken(EOF) {
		err := p.statement()
		if err != nil {
			return err
		}
	}

	for label := range p.labelsGotoed {
		if !p.labelsDeclared[label] {
			return fmt.Errorf("attempting to GOTO undeclared label: %s", label)
		}
	}
	return nil
}

func (p *Parser) statement() error {
	var err error
	switch {
	case p.checkToken(PRINT):
		log.Print("STATEMENT-PRINT")

		err = p.nextToken()
		if err != nil {
			return err
		}
		if p.checkToken(STRING) {
			p.emitter.addImport("fmt")
			p.emitter.emitLine(fmt.Sprintf(`fmt.Println("%s");`, p.currentToken.text))
			err = p.nextToken()
			if err != nil {
				return err
			}
		} else {
			p.emitter.emit(`fmt.Printf("%v\n",`)

			err = p.expression()
			if err != nil {
				return err
			}
			p.emitter.emit(")")
		}
	case p.checkToken(IF):
		log.Print("STATEMENT-IF")
		err = p.nextToken()
		if err != nil {
			return err
		}
		p.emitter.emit("if ")

		err = p.comparison()
		if err != nil {
			return err
		}
		p.emitter.emitLine("{")

		err = p.match(THEN)
		if err != nil {
			return err
		}
		for !p.checkToken(ENDIF) {
			err = p.statement()
			if err != nil {
				return err
			}
		}
		err = p.match(ENDIF)
		if err != nil {
			return err
		}
		p.emitter.emitLine("}")
	case p.checkToken(WHILE):
		log.Print("STATEMENT-WHILE")
		err = p.nextToken()
		if err != nil {
			return err
		}
		p.emitter.emit("for ")

		err = p.comparison()
		if err != nil {
			return err
		}
		p.emitter.emitLine("{")

		err = p.match(REPEAT)
		if err != nil {
			return err
		}
		err = p.nl()
		if err != nil {
			return err
		}

		for !p.checkToken(ENDWHILE) {
			err = p.statement()
			if err != nil {
				return err
			}
		}
		err = p.match(ENDWHILE)
		if err != nil {
			return err
		}
		p.emitter.emitLine("}")
	case p.checkToken(LABEL):
		log.Print("STATEMENT-LABEL")
		err = p.nextToken()
		if err != nil {
			return err
		}
		if p.labelsDeclared[p.currentToken.text] {
			return fmt.Errorf("label %s already exists", p.currentToken.text)
		}
		p.labelsDeclared[p.currentToken.text] = true

		p.emitter.emitLine(p.currentToken.text + ":")
		err = p.match(IDENT)
		if err != nil {
			return err
		}
	case p.checkToken(GOTO):
		log.Print("STATEMENT-GOTO")
		p.nextToken()
		if err != nil {
			return err
		}
		p.labelsGotoed[p.currentToken.text] = true
		p.emitter.emitLine("goto " + p.currentToken.text)
		err = p.match(IDENT)
		if err != nil {
			return err
		}
	case p.checkToken(LET):
		log.Print("STATEMENT-LET")
		err = p.nextToken()
		if err != nil {
			return err
		}

		seen := p.symbols[p.currentToken.text]
		if !seen {
			p.symbols[p.currentToken.text] = true
		}
		name := p.currentToken.text

		err = p.match(IDENT)
		if err != nil {
			return err
		}

		err = p.match(EQ)
		if err != nil {
			return err
		}

		if seen {
			p.emitter.emit(name + " = ")
		} else {
			p.emitter.emit(name + " := ")
		}

		err = p.expression()
		if err != nil {
			return err
		}
		p.emitter.emit(";")

	case p.checkToken(INPUT):
		log.Print("STATEMENT-INPUT")
		p.emitter.addImport("fmt")

		err = p.nextToken()
		if err != nil {
			return err
		}

		name := p.currentToken.text
		seen := p.symbols[name]
		if !seen {
			p.symbols[name] = true
		}

		err = p.match(IDENT)
		if err != nil {
			return err
		}

		if seen {
			p.emitter.emitLine(fmt.Sprintf(`fmt.Scanf("%%f", &%s);`, name))
		} else {
			p.emitter.emitLine(fmt.Sprintf(`var %s float64; fmt.Scanf("%%f", &%s);`, name, name))

		}
	default:
		return fmt.Errorf("Invalid statement at %s (%s)", p.currentToken.text, TOKENS[p.currentToken.kind])
	}
	return p.nl()
}

// expression ::= term {( "-" | "+" ) term}
func (p *Parser) expression() error {
	log.Print("EXPRESSION")
	err := p.term()
	if err != nil {
		return err
	}

	for p.checkToken(PLUS) || p.checkToken(MINUS) {
		p.emitter.emit(p.currentToken.text)
		err = p.nextToken()
		if err != nil {
			return err
		}
		err = p.term()
		if err != nil {
			return err
		}
	}
	return nil
}

// comparison ::= expression (("==" | "!=" | ">" | ">=" | "<" | "<=") expression)+
func (p *Parser) comparison() error {
	log.Print("COMPARISON")
	err := p.expression()
	if err != nil {
		return err
	}
	if p.currentToken.isComparisonOperator() {
		p.emitter.emit(p.currentToken.text)
		err = p.nextToken()
		if err != nil {
			return err
		}
		err = p.expression()
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Expected comparison operator at: %s", p.currentToken.text)
	}

	for p.currentToken.isComparisonOperator() {
		p.emitter.emit(p.currentToken.text)
		err = p.nextToken()
		if err != nil {
			return err
		}
		err = p.expression()
		if err != nil {
			return err
		}
	}

	return nil
}

// term ::= unary {( "/" | "*" ) unary}
func (p *Parser) term() error {
	log.Print("TERM")
	err := p.unary()
	if err != nil {
		return err
	}
	for p.checkToken(ASTERISK) || p.checkToken(SLASH) {
		p.emitter.emit(p.currentToken.text)
		err = p.nextToken()
		if err != nil {
			return err
		}
		err = p.unary()
		if err != nil {
			return err
		}
	}
	return nil
}

// unary ::= ["+" | "-"] primary
func (p *Parser) unary() error {
	log.Print("UNARY")

	if p.checkToken(PLUS) || p.checkToken(MINUS) {
		p.emitter.emit(p.currentToken.text)
		err := p.nextToken()
		if err != nil {
			return err
		}
	}
	return p.primary()
}

func (p *Parser) primary() error {
	log.Printf("Primary (%s)", p.currentToken.text)
	if p.checkToken(NUMBER) {
		p.emitter.emit("float64(" + p.currentToken.text + ")")
		return p.nextToken()
	} else if p.checkToken(IDENT) {
		if !p.symbols[p.currentToken.text] {
			return fmt.Errorf("referencing variable before assignment: %s", p.currentToken.text)
		}
		p.emitter.emit(p.currentToken.text)
		return p.nextToken()
	} else {
		return fmt.Errorf("unexpected token at %s", p.currentToken.text)
	}
}

// nl ::= '\n'+
func (p *Parser) nl() error {
	log.Print("NEWLINE")

	err := p.match(NEWLINE)
	if err == nil {
		return nil
	}
	for p.checkToken(NEWLINE) {
		err = p.nextToken()
		if err != nil {
			return err
		}
	}
	return nil
}
