package parser

import (
	"aura/src/ast"
	l "aura/src/lexer"
	"fmt"
)

// Signature for functions to parse prefix expressions
type PrefixParsFn func() ast.Expression

// Signature for functions to parse suffix expressions
type SuffixParseFn func(ast.Expression) ast.Expression

// Signature for functions to parse infix expressions
type InfixParseFn func(ast.Expression) ast.Expression

type PrefixParsFns map[l.TokenType]PrefixParsFn
type InfixParseFns map[l.TokenType]InfixParseFn
type SuffixParseFns map[l.TokenType]SuffixParseFn

// represents the precedence of evaluation
type Precedence int

const (
	HeadPrecendence Precedence = iota
	LOWEST                     = 1
	ANDOR                      = 2
	EQUEAL                     = 3
	LESSGRATER                 = 4
	SUM                        = 5
	PRODUCT                    = 6
	PREFIX                     = 7
	CALL                       = 8
)

var precedences = map[l.TokenType]Precedence{
	l.AND:         ANDOR,
	l.EQ:          EQUEAL,
	l.NOT_EQ:      EQUEAL,
	l.LT:          LESSGRATER,
	l.LTOREQ:      LESSGRATER,
	l.GT:          LESSGRATER,
	l.GTOREQ:      LESSGRATER,
	l.PLUS:        SUM,
	l.MINUS:       SUM,
	l.DIVISION:    PRODUCT,
	l.TIMES:       PRODUCT,
	l.MOD:         PRODUCT,
	l.LPAREN:      CALL,
	l.LBRACKET:    CALL,
	l.OR:          ANDOR,
	l.ASSING:      ANDOR,
	l.COLON:       CALL,
	l.PLUSASSING:  PRODUCT,
	l.MINUSASSING: PRODUCT,
	l.DIVASSING:   PRODUCT,
	l.EXPONENT:    PRODUCT,
	l.TIMEASSI:    PRODUCT,
	l.PLUS2:       PRODUCT,
	l.MINUS2:      PRODUCT,
	l.DOT:         PREFIX,
	l.COLONASSING: PREFIX,
}

// Represents the Parser of the programming lenguage
type Parser struct {
	lexer          *l.Lexer       // represents the lexer of the programming lenguage
	currentToken   *l.Token       // represents the current token in the parsing
	peekToken      *l.Token       // represnts the next token in the parsing
	lastToken      *l.Token       // represents the previus token in the parsing
	errors         []string       // represents the error found while parsing
	prefixParsFns  PrefixParsFns  // represents all the functions to parse prefix expressions
	infixParseFns  InfixParseFns  // represents all the functions to parse infix expressions
	suffixParseFns SuffixParseFns // represents all the functions to parse suffix expressions
}

// generates a new parser instance
func NewParser(lexer *l.Lexer) *Parser {
	parser := &Parser{
		lexer:          lexer,
		currentToken:   nil,
		peekToken:      nil,
		prefixParsFns:  make(PrefixParsFns),
		infixParseFns:  make(InfixParseFns),
		suffixParseFns: make(SuffixParseFns),
	}

	// we register all the functions to parse the expressions
	parser.registerPrefixFns()
	parser.registerInfixFns()
	parser.registerSuffixFns()

	// we advance two times tokens to have a not nil first token
	parser.advanceTokens()
	parser.advanceTokens()
	return parser
}

// advance 1 in the tokens generated by the lexer
func (p *Parser) advanceTokens() {
	p.lastToken = p.currentToken
	p.currentToken = p.peekToken
	nextToken := p.lexer.NextToken()
	p.peekToken = &nextToken
}

// check that the current token is not nil
func (p *Parser) checkCurrentTokenIsNotNil() {
	if p.currentToken == nil {
		panic("Error de parseo se esperaba una expression despues de: " + p.lastToken.Literal)
	}
}

// check that the peek token is not nil
func (p *Parser) checkPeekTokenIsNotNil() {
	if p.peekToken == nil {
		panic("Error de parseo se esperaba una expression despues de: " + p.currentToken.Literal)
	}
}

// return the precedence of the current token
func (p *Parser) currentPrecedence() Precedence {
	p.checkCurrentTokenIsNotNil()
	precedence, exists := precedences[p.currentToken.Token_type]
	if !exists {
		return LOWEST
	}

	return precedence
}

// return the error list in the parser
func (p *Parser) Errors() []string {
	return p.errors
}

// parse all the program
func (p *Parser) ParseProgam() ast.Program {
	program := ast.Program{Staments: []ast.Stmt{}}

	for p.currentToken.Token_type != l.EOF {
		if statement := p.parseStament(); statement != nil {
			program.Staments = append(program.Staments, statement)
		}

		p.advanceTokens()
	}
	return program
}

// expectedToken will check if the peek token is the correct type
// based on the parameter
func (p *Parser) expepectedToken(tokenType l.TokenType) bool {
	if p.peekToken.Token_type == tokenType {
		p.advanceTokens()
		return true
	}

	p.expectedTokenError(tokenType)
	return false
}

// add an error to errors list if there is any unexpected token error
func (p *Parser) expectedTokenError(tokenType l.TokenType) {
	p.checkCurrentTokenIsNotNil()
	err := fmt.Sprintf(
		"se esperaba que el siguient token fuera %s pero se obtuvo %s",
		l.Tokens[tokenType],
		l.Tokens[p.peekToken.Token_type],
	)
	p.errors = append(p.errors, err)
}

// parseBlock will parse a block expression
func (p *Parser) parseBlock() *ast.Block {
	p.checkCurrentTokenIsNotNil()
	blockStament := ast.NewBlock(*p.currentToken)
	p.advanceTokens()

	// we iterate until we find a } token
	for !(p.currentToken.Token_type == l.RBRACE) && !(p.currentToken.Token_type == l.EOF) {
		if stament := p.parseStament(); stament != nil {
			blockStament.Staments = append(blockStament.Staments, stament)
		}

		p.advanceTokens()
	}

	return blockStament
}

// ParseArrayValues will parse all the values in array expressions
func (p *Parser) ParseArrayValues() []ast.Expression {
	p.checkCurrentTokenIsNotNil()
	var values []ast.Expression
	if p.peekToken.Token_type == l.RBRACKET {
		p.advanceTokens()
		return values
	}

	p.advanceTokens()
	if expression := p.parseExpression(LOWEST); expression != nil {
		values = append(values, expression)
	}

	for p.peekToken.Token_type == l.COMMA {
		p.advanceTokens()
		p.advanceTokens()
		if expression := p.parseExpression(LOWEST); expression != nil {
			values = append(values, expression)
		}
	}

	if !p.expepectedToken(l.RBRACKET) {
		return make([]ast.Expression, 0)
	}

	return values
}

// parse all the arguments when a function is call
func (p *Parser) parseCallArguments() []ast.Expression {
	var args []ast.Expression
	p.checkPeekTokenIsNotNil()
	if p.peekToken.Token_type == l.RPAREN {
		// there is no arguemnts
		p.advanceTokens()
		return args
	}

	p.advanceTokens()
	if expression := p.parseExpression(LOWEST); expression != nil {
		args = append(args, expression)
	}

	// we loop until we dont have commas. this means whe parse all the values
	for p.peekToken.Token_type == l.COMMA {
		p.advanceTokens()
		p.advanceTokens()
		if expression := p.parseExpression(LOWEST); expression != nil {
			args = append(args, expression)
		}
	}

	if !p.expepectedToken(l.RPAREN) {
		return make([]ast.Expression, 0)
	}

	return args
}

// parse a expression based on the given precedence
func (p *Parser) parseExpression(precedence Precedence) ast.Expression {
	p.checkCurrentTokenIsNotNil()

	// we check if there is any function to parse the current token
	prefixParseFn, exist := p.prefixParsFns[p.currentToken.Token_type]
	if !exist {
		// there is no function to parse the token
		message := fmt.Sprintf("no se encontro ninguna funcion para parsear %s", p.currentToken.Literal)
		p.errors = append(p.errors, message)
		return nil
	}

	leftExpression := prefixParseFn()
	p.checkPeekTokenIsNotNil()

	// we check if there is any suffix expression to be parsed
	if suffixFn, exists := p.suffixParseFns[p.peekToken.Token_type]; exists {
		p.advanceTokens()
		leftExpression = suffixFn(leftExpression)
		p.advanceTokens()
	}

	// we loop until the precedence is lowest than the next precedence
	for !(p.peekToken.Token_type == l.SEMICOLON) && precedence < p.peekPrecedence() {
		// we check if there is any function to parse an infix expression
		infixParseFn, exist := p.infixParseFns[p.peekToken.Token_type]
		if !exist {
			return leftExpression
		}

		p.advanceTokens()
		if leftExpression == nil {
			panic("Error de parseo :(")
		}

		leftExpression = infixParseFn(leftExpression)
	}

	return leftExpression
}

func (p *Parser) parseClassStatement() ast.Stmt {
	p.checkCurrentTokenIsNotNil()
	class := ast.NewClassStatement(*p.currentToken, nil, nil, []*ast.ClassMethodExp{})
	if !p.expepectedToken(l.IDENT) {
		return nil
	}

	class.Name = p.parseIdentifier().(*ast.Identifier)
	if !p.expepectedToken(l.LPAREN) {
		return nil
	}

	class.Params = p.parseFunctionParameters()
	if !p.expepectedToken(l.LBRACE) {
		return nil
	}

	p.advanceTokens()
	for p.currentToken.Token_type != l.RBRACE && p.currentToken.Token_type != l.EOF {
		if expression := p.parseClassMethod(); expression != nil {
			if method, isMethod := expression.(*ast.ClassMethodExp); isMethod {
				class.Methods = append(class.Methods, method)
			}
		}
	}

	return class
}

// parse a expression statement
func (p *Parser) parserExpressionStatement() *ast.ExpressionStament {
	p.checkCurrentTokenIsNotNil()
	expressionStament := ast.NewExpressionStament(*p.currentToken, nil)
	expressionStament.Expression = p.parseExpression(LOWEST)

	p.checkPeekTokenIsNotNil()
	if p.peekToken.Token_type == l.SEMICOLON {
		p.advanceTokens()
	}

	return expressionStament
}

// parse all the parameters of the function expresison
func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	var params []*ast.Identifier
	p.checkPeekTokenIsNotNil()
	if p.peekToken.Token_type == l.RPAREN {
		// there is no parameters
		p.advanceTokens()
		return params
	}

	p.advanceTokens()
	identifier := ast.NewIdentifier(*p.currentToken, p.currentToken.Literal)
	params = append(params, identifier)

	// we loop until we dont have commas. this means we parse all the parameters
	for p.peekToken.Token_type == l.COMMA {
		p.advanceTokens()
		p.advanceTokens()
		identifier = ast.NewIdentifier(*p.currentToken, p.currentToken.Literal)
		params = append(params, identifier)
	}

	if !p.expepectedToken(l.RPAREN) {
		// syntax error
		return make([]*ast.Identifier, 0)
	}

	return params
}

// parse a suffix function
func (p *Parser) parseSuffixFn(left ast.Expression) ast.Expression {
	return ast.NewSuffix(*p.currentToken, left, p.currentToken.Literal)
}

// parse a null expression
func (p *Parser) ParseNull() ast.Expression {
	p.checkCurrentTokenIsNotNil()
	return ast.NewNull(*p.currentToken)
}

// parse a let statement
func (p *Parser) parseLetSatement() ast.Stmt {
	p.checkCurrentTokenIsNotNil()
	stament := ast.NewLetStatement(*p.currentToken, nil, nil)
	if !p.expepectedToken(l.IDENT) {
		return nil
	}

	stament.Name = p.parseIdentifier().(*ast.Identifier)
	if !p.expepectedToken(l.ASSING) {
		// syntax error. we dont allow this -> var name 5;
		return nil
	}

	p.advanceTokens()
	stament.Value = p.parseExpression(LOWEST)
	p.checkPeekTokenIsNotNil()
	if p.peekToken.Token_type == l.SEMICOLON {
		p.advanceTokens()
	}

	return stament
}

// parse a return stament
func (p *Parser) parseReturnStatement() ast.Stmt {
	p.checkCurrentTokenIsNotNil()
	stament := ast.NewReturnStatement(*p.currentToken, nil)
	p.advanceTokens()

	stament.ReturnValue = p.parseExpression(LOWEST)
	p.checkPeekTokenIsNotNil()
	if p.peekToken.Token_type == l.SEMICOLON {
		p.advanceTokens()
	}

	return stament
}

// check current token and parse the token as a expression, let stament or return stament
func (p *Parser) parseStament() ast.Stmt {
	p.checkCurrentTokenIsNotNil()
	switch p.currentToken.Token_type {
	case l.LET:
		return p.parseLetSatement()

	case l.RETURN:
		return p.parseReturnStatement()

	case l.CLASS:
		return p.parseClassStatement()

	case l.IMPORT:
		return p.parseImportStatement()

	default:
		return p.parserExpressionStatement()
	}
}

// return the precedence of the next token
func (p *Parser) peekPrecedence() Precedence {
	p.checkPeekTokenIsNotNil()
	precedence, exists := precedences[p.peekToken.Token_type]
	if !exists {
		return LOWEST
	}

	return precedence
}

// register all the functions to parse infix expressions
func (p *Parser) registerInfixFns() {
	p.infixParseFns[l.PLUS] = p.parseInfixExpression
	p.infixParseFns[l.MINUS] = p.parseInfixExpression
	p.infixParseFns[l.COLON] = p.parseMethod
	p.infixParseFns[l.DIVISION] = p.parseInfixExpression
	p.infixParseFns[l.TIMES] = p.parseInfixExpression
	p.infixParseFns[l.EQ] = p.parseInfixExpression
	p.infixParseFns[l.NOT_EQ] = p.parseInfixExpression
	p.infixParseFns[l.GTOREQ] = p.parseInfixExpression
	p.infixParseFns[l.LTOREQ] = p.parseInfixExpression
	p.infixParseFns[l.LT] = p.parseInfixExpression
	p.infixParseFns[l.IN] = p.parseInfixExpression
	p.infixParseFns[l.GT] = p.parseInfixExpression
	p.infixParseFns[l.PLUSASSING] = p.parseInfixExpression
	p.infixParseFns[l.MINUSASSING] = p.parseInfixExpression
	p.infixParseFns[l.TIMEASSI] = p.parseInfixExpression
	p.infixParseFns[l.DIVASSING] = p.parseInfixExpression
	p.infixParseFns[l.EXPONENT] = p.parseInfixExpression
	p.infixParseFns[l.LPAREN] = p.parseCall
	p.infixParseFns[l.ASSING] = p.parseReassigment
	p.infixParseFns[l.LBRACKET] = p.parseCallList
	p.infixParseFns[l.MOD] = p.parseInfixExpression
	p.infixParseFns[l.AND] = p.parseInfixExpression
	p.infixParseFns[l.OR] = p.parseInfixExpression
	p.infixParseFns[l.DOT] = p.parseClassFieldsCall
	p.infixParseFns[l.COLONASSING] = p.parseAssigmentExp
}

// register all the functions to parse prefix expressions
func (p *Parser) registerPrefixFns() {
	p.prefixParsFns[l.FALSE] = p.parseBoolean
	p.prefixParsFns[l.FOR] = p.parseFor
	p.prefixParsFns[l.FUNCTION] = p.parseFunction
	p.prefixParsFns[l.WHILE] = p.parseWhile
	p.prefixParsFns[l.IDENT] = p.parseIdentifier
	p.prefixParsFns[l.IF] = p.parseIf
	p.prefixParsFns[l.INT] = p.parseInteger
	p.prefixParsFns[l.LPAREN] = p.parseGroupExpression
	p.prefixParsFns[l.MINUS] = p.parsePrefixExpression
	p.prefixParsFns[l.NOT] = p.parsePrefixExpression
	p.prefixParsFns[l.TRUE] = p.parseBoolean
	p.prefixParsFns[l.STRING] = p.parseStringLiteral
	p.prefixParsFns[l.DATASTRCUT] = p.ParseArray
	p.prefixParsFns[l.NULLT] = p.ParseNull
	p.prefixParsFns[l.MAP] = p.parseMap
	p.prefixParsFns[l.FLOAT] = p.parseFloat
	p.prefixParsFns[l.NEW] = p.parseClassCall
}

// register all the functions to parse suffix expressions
func (p *Parser) registerSuffixFns() {
	p.suffixParseFns[l.EXPONENT] = p.parseSuffixFn
	p.suffixParseFns[l.PLUS2] = p.parseSuffixFn
	p.suffixParseFns[l.MINUS2] = p.parseSuffixFn
}
