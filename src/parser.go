package src

import (
	"fmt"
	"strconv"
)

type PrefixParsFn func() Expression
type InfixParseFn func(Expression) Expression

type PrefixParsFns map[TokenType]PrefixParsFn
type InfixParseFns map[TokenType]InfixParseFn

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

var PRECEDENCES = map[TokenType]Precedence{
	AND:      ANDOR,
	EQ:       EQUEAL,
	NOT_EQ:   EQUEAL,
	LT:       LESSGRATER,
	LTOREQ:   LESSGRATER,
	GT:       LESSGRATER,
	GTOREQ:   LESSGRATER,
	PLUS:     SUM,
	MINUS:    SUM,
	DIVISION: PRODUCT,
	TIMES:    PRODUCT,
	MOD:      PRODUCT,
	LPAREN:   CALL,
	LBRACKET: CALL,
	OR:       ANDOR,
	ASSING:   ANDOR,
	COLON:    CALL,
}

// parser handle the parsing of the program staments and syntax of the program
type Parser struct {
	lexer         *Lexer
	currentToken  *Token
	peekToken     *Token
	errors        []string
	prefixParsFns PrefixParsFns
	infixParseFns InfixParseFns
}

// generates a new parser instance
func NewParser(lexer *Lexer) *Parser {
	parser := &Parser{
		lexer:        lexer,
		currentToken: nil,
		peekToken:    nil,
	}

	parser.prefixParsFns = parser.registerPrefixFns()
	parser.infixParseFns = parser.registerInfixFns()
	parser.advanceTokens()
	parser.advanceTokens()
	return parser
}

// advance 1 in the tokens generated by the lexer
func (p *Parser) advanceTokens() {
	p.currentToken = p.peekToken
	nextToken := p.lexer.NextToken()
	p.peekToken = &nextToken
}

// check that the current token is not nil
func (p *Parser) checkCurrentTokenIsNotNil() {
	defer p.handlePeekTokenPanic()
	if p.currentToken == nil {
		panic("current token cannot be nil")
	}
}

// check that the peek token is not nil
func (p *Parser) checkPeekTokenIsNotNil() {
	defer p.handlePeekTokenPanic()
	if p.peekToken == nil {
		panic("peek token cannot be nil")
	}
}

func (p *Parser) handlePeekTokenPanic() {
	if r := recover(); r != nil {
		fmt.Println("Syntax error: ", r)
	}
}

// check precedence of the current token
func (p *Parser) currentPrecedence() Precedence {
	p.checkCurrentTokenIsNotNil()
	precedence, exists := PRECEDENCES[p.currentToken.Token_type]
	if !exists {
		return LOWEST
	}

	return precedence
}

// return the error list in the parser
func (p *Parser) Errors() []string {
	return p.errors
}

// parse all program staments
func (p *Parser) ParseProgam() Program {
	program := Program{Staments: []Stmt{}}

	for p.currentToken.Token_type != EOF {
		statement := p.parseStament()
		if statement != nil {
			program.Staments = append(program.Staments, statement)
		}

		p.advanceTokens()
	}
	return program
}

// expectedToken will check if the peek token is the correct type
// based on the parameter
func (p *Parser) expepectedToken(tokenType TokenType) bool {
	if p.peekToken.Token_type == tokenType {
		p.advanceTokens()
		return true
	}

	p.expectedTokenError(tokenType)
	return false
}

// add a error to errors list if there is any unexpected token error
func (p *Parser) expectedTokenError(tokenType TokenType) {
	p.checkCurrentTokenIsNotNil()
	err := fmt.Sprintf(
		"se esperaba que el siguient token fuera %s pero se obtuvo %s",
		tokens[tokenType],
		tokens[p.peekToken.Token_type],
	)
	p.errors = append(p.errors, err)
}

// parse boolean expression and check if true or false
func (p *Parser) parseBoolean() Expression {
	p.checkCurrentTokenIsNotNil()
	var value bool
	if p.currentToken.Token_type == TRUE {
		value = true
		return NewBoolean(*p.currentToken, &value)
	}

	value = false
	return NewBoolean(*p.currentToken, &value)
}

// parse a block of staments
func (p *Parser) parseBlock() *Block {
	p.checkCurrentTokenIsNotNil()
	blockStament := NewBlock(*p.currentToken)
	p.advanceTokens()

	for !(p.currentToken.Token_type == RBRACE) && !(p.currentToken.Token_type == EOF) {
		stament := p.parseStament()
		if stament != nil {
			blockStament.Staments = append(blockStament.Staments, stament)
		}

		p.advanceTokens()
	}

	return blockStament
}

// parse function calls
func (p *Parser) parseCall(function Expression) Expression {
	p.checkCurrentTokenIsNotNil()
	call := NewCall(*p.currentToken, function)
	call.Arguments = p.parseCallArguments()
	return call
}

func (p *Parser) ParseArray() Expression {
	p.checkCurrentTokenIsNotNil()
	arr := NewArray(*p.currentToken, nil)
	if !p.expepectedToken(LBRACKET) {
		return nil
	}

	arr.Values = p.ParseArrayValues()
	return arr
}

func (p *Parser) ParseArrayValues() []Expression {
	p.checkCurrentTokenIsNotNil()
	var values []Expression
	if p.peekToken.Token_type == RBRACKET {
		p.advanceTokens()
		return values
	}

	p.advanceTokens()
	if expression := p.parseExpression(LOWEST); expression != nil {
		values = append(values, expression)
	}

	for p.peekToken.Token_type == COMMA {
		p.advanceTokens()
		p.advanceTokens()
		if expression := p.parseExpression(LOWEST); expression != nil {
			values = append(values, expression)
		}
	}

	if !p.expepectedToken(RBRACKET) {
		return nil
	}

	return values
}

// parse args in function calls
func (p *Parser) parseCallArguments() []Expression {
	var args []Expression
	p.checkPeekTokenIsNotNil()
	if p.peekToken.Token_type == RPAREN {
		p.advanceTokens()
		return args
	}

	p.advanceTokens()
	if expression := p.parseExpression(LOWEST); expression != nil {
		args = append(args, expression)
	}

	for p.peekToken.Token_type == COMMA {
		p.advanceTokens()
		p.advanceTokens()
		if expression := p.parseExpression(LOWEST); expression != nil {
			args = append(args, expression)
		}
	}

	if !p.expepectedToken(RPAREN) {
		return nil
	}

	return args
}

// parse a expression and check if there is a valid expression
func (p *Parser) parseExpression(precedence Precedence) Expression {
	defer p.handlePeekTokenPanic()
	p.checkCurrentTokenIsNotNil()
	prefixParseFn, exist := p.prefixParsFns[p.currentToken.Token_type]
	if !exist {
		message := fmt.Sprintf("no se encontro ninguna funcion para parsear %s", p.currentToken.Literal)
		p.errors = append(p.errors, message)
		return nil
	}

	leftExpression := prefixParseFn()
	p.checkPeekTokenIsNotNil()

	for !(p.peekToken.Token_type == SEMICOLON) && precedence < p.peekPrecedence() {
		infixParseFn, exist := p.infixParseFns[p.peekToken.Token_type]
		if !exist {
			return leftExpression
		}

		p.advanceTokens()
		if leftExpression == nil {
			panic("left expression cannot be nil while parsing a expression")
		}

		leftExpression = infixParseFn(leftExpression)
	}

	return leftExpression
}

// parse a expression statement
func (p *Parser) parserExpressionStatement() *ExpressionStament {
	p.checkCurrentTokenIsNotNil()
	expressionStament := NewExpressionStament(*p.currentToken, nil)
	expressionStament.Expression = p.parseExpression(LOWEST)

	if p.peekToken == nil {
		panic("peek token cannot be bil")
	}
	if p.peekToken.Token_type == SEMICOLON {
		p.advanceTokens()
	}

	return expressionStament
}

// parse group expression like (5 + 5) / 2
func (p *Parser) parseGroupExpression() Expression {
	p.advanceTokens()
	expression := p.parseExpression(LOWEST)
	if !p.expepectedToken(RPAREN) {
		return nil
	}

	return expression
}

func (p *Parser) parseReassigment(ident Expression) Expression {
	p.checkCurrentTokenIsNotNil()
	reassignment := NewReassignment(*p.currentToken, ident, nil)
	p.advanceTokens()
	reassignment.NewVal = p.parseExpression(LOWEST)
	return reassignment
}

// parse a function declaration
func (p *Parser) parseFunction() Expression {
	p.checkCurrentTokenIsNotNil()
	function := NewFunction(*p.currentToken, nil)
	if !p.expepectedToken(LPAREN) {
		return nil
	}

	function.Parameters = p.parseFunctionParameters()
	if !p.expepectedToken(LBRACE) {
		return nil
	}

	function.Body = p.parseBlock()
	return function
}

// parse function parameters and check the syntax
func (p *Parser) parseFunctionParameters() []*Identifier {
	var params []*Identifier
	p.checkPeekTokenIsNotNil()
	if p.peekToken.Token_type == RPAREN {
		p.advanceTokens()
		return params
	}

	p.advanceTokens()
	identifier := NewIdentifier(*p.currentToken, p.currentToken.Literal)
	params = append(params, identifier)

	for p.peekToken.Token_type == COMMA {
		p.advanceTokens()
		p.advanceTokens()
		identifier = NewIdentifier(*p.currentToken, p.currentToken.Literal)
		params = append(params, identifier)
	}

	if !p.expepectedToken(RPAREN) {
		return make([]*Identifier, 0)
	}

	return params
}

// parse a identifier
func (p *Parser) parseIdentifier() Expression {
	p.checkCurrentTokenIsNotNil()
	return &Identifier{token: *p.currentToken, value: p.currentToken.Literal}
}

// parse infix expressoins
func (p *Parser) parseInfixExpression(left Expression) Expression {
	p.checkCurrentTokenIsNotNil()
	infix := Newinfix(*p.currentToken, nil, p.currentToken.Literal, left)
	precedence := p.currentPrecedence()
	p.advanceTokens()
	infix.Rigth = p.parseExpression(precedence)
	return infix
}

func (p *Parser) parseRangeExpression() Expression {
	rangeExpress := NewRange(*p.currentToken, nil, nil)
	if !p.expepectedToken(IDENT) {
		return nil
	}

	rangeExpress.Variable = p.parseIdentifier()
	if !p.expepectedToken(IN) {
		return nil
	}

	p.advanceTokens()
	rangeExpress.Range = p.parseExpression(LOWEST)
	return rangeExpress
}

func (p *Parser) parseFor() Expression {
	p.checkCurrentTokenIsNotNil()
	forExpression := NewFor(*p.currentToken, nil, nil)
	if !p.expepectedToken(LPAREN) {
		return nil
	}

	forExpression.Condition = p.parseRangeExpression()
	if !p.expepectedToken(RPAREN) {
		return nil
	}
	if !p.expepectedToken(LBRACE) {
		return nil
	}

	forExpression.Body = p.parseBlock()
	return forExpression
}

func (p *Parser) parseCallList(valueList Expression) Expression {
	p.checkCurrentTokenIsNotNil()
	callList := NewCallList(*p.currentToken, valueList, nil)
	p.advanceTokens()
	callList.Index = p.parseExpression(LOWEST)
	if !p.expepectedToken(RBRACKET) {
		return nil
	}

	return callList
}

func (p *Parser) parseWhile() Expression {
	p.checkCurrentTokenIsNotNil()
	whileExpression := NewWhile(*p.currentToken, nil, nil)
	if !p.expepectedToken(LPAREN) {
		return nil
	}

	p.advanceTokens()
	whileExpression.Condition = p.parseExpression(LOWEST)
	if !p.expepectedToken(RPAREN) {
		return nil
	}

	if !p.expepectedToken(LBRACE) {
		return nil
	}

	whileExpression.Body = p.parseBlock()
	return whileExpression
}

// parse if expressions, check sintax and if there is an else in the expression
func (p *Parser) parseIf() Expression {
	p.checkCurrentTokenIsNotNil()
	ifExpression := NewIf(*p.currentToken, nil, nil, nil)
	if !p.expepectedToken(LPAREN) {
		return nil
	}

	p.advanceTokens()
	ifExpression.Condition = p.parseExpression(LOWEST)
	if !p.expepectedToken(RPAREN) {
		return nil
	}

	if !p.expepectedToken(LBRACE) {
		return nil
	}

	ifExpression.Consequence = p.parseBlock()

	p.checkPeekTokenIsNotNil()
	if p.peekToken.Token_type == ELSE {
		p.advanceTokens()
		if !p.expepectedToken(LBRACE) {
			return nil
		}

		ifExpression.Alternative = p.parseBlock()
	}

	return ifExpression
}

func (p *Parser) parseMethod(left Expression) Expression {
	p.checkCurrentTokenIsNotNil()
	method := NewMethodExpression(*p.currentToken, left, nil)
	if !p.expepectedToken(IDENT) {
		return nil
	}

	method.Method = p.parseExpression(LOWEST)
	return method
}

// parse integer expressions
func (p *Parser) parseInteger() Expression {
	p.checkCurrentTokenIsNotNil()
	integer := NewInteger(*p.currentToken, nil)

	val, err := strconv.Atoi(p.currentToken.Literal)
	if err != nil {
		message := fmt.Sprintf("no se pudo parsear %s como entero", p.currentToken.Literal)
		p.errors = append(p.errors, message)
		return nil
	}

	integer.Value = &val
	return integer
}

// parse given let stament and check sintax
func (p *Parser) parseLetSatement() Stmt {
	p.checkCurrentTokenIsNotNil()
	stament := NewLetStatement(*p.currentToken, nil, nil)
	if !p.expepectedToken(IDENT) {
		return nil
	}

	stament.Name = p.parseIdentifier().(*Identifier)
	if !p.expepectedToken(ASSING) {
		return nil
	}

	p.advanceTokens()
	stament.Value = p.parseExpression(LOWEST)
	p.checkPeekTokenIsNotNil()
	if p.peekToken.Token_type == SEMICOLON {
		p.advanceTokens()
	}

	return stament
}

// parse a prefix expression
func (p *Parser) parsePrefixExpression() Expression {
	p.checkCurrentTokenIsNotNil()
	prefixExpression := NewPrefix(*p.currentToken, p.currentToken.Literal, nil)
	p.advanceTokens()
	prefixExpression.Rigth = p.parseExpression(PREFIX)
	return prefixExpression
}

// parse given return stament
func (p *Parser) parseReturnStatement() Stmt {
	p.checkCurrentTokenIsNotNil()
	stament := NewReturnStatement(*p.currentToken, nil)
	p.advanceTokens()

	stament.ReturnValue = p.parseExpression(LOWEST)
	p.checkPeekTokenIsNotNil()
	if p.peekToken.Token_type == SEMICOLON {
		p.advanceTokens()
	}

	return stament
}

// check current token and parse the token as a expression, let stament or return stament
func (p *Parser) parseStament() Stmt {
	p.checkCurrentTokenIsNotNil()
	if p.currentToken.Token_type == LET {
		return p.parseLetSatement()
	} else if p.currentToken.Token_type == RETURN {
		return p.parseReturnStatement()
	}

	return p.parserExpressionStatement()
}

func (p *Parser) parseStringLiteral() Expression {
	p.checkCurrentTokenIsNotNil()
	return NewStringLiteral(*p.currentToken, p.currentToken.Literal)
}

// check the precedence of the current token
func (p *Parser) peekPrecedence() Precedence {
	p.checkPeekTokenIsNotNil()
	precedence, exists := PRECEDENCES[p.peekToken.Token_type]
	if !exists {
		return LOWEST
	}

	return precedence
}

// register all infix functions for the different token types
func (p *Parser) registerInfixFns() InfixParseFns {
	inFixFns := make(InfixParseFns)
	inFixFns[PLUS] = p.parseInfixExpression
	inFixFns[MINUS] = p.parseInfixExpression
	inFixFns[COLON] = p.parseMethod
	inFixFns[DIVISION] = p.parseInfixExpression
	inFixFns[TIMES] = p.parseInfixExpression
	inFixFns[EQ] = p.parseInfixExpression
	inFixFns[NOT_EQ] = p.parseInfixExpression
	inFixFns[GTOREQ] = p.parseInfixExpression
	inFixFns[LTOREQ] = p.parseInfixExpression
	inFixFns[LT] = p.parseInfixExpression
	inFixFns[IN] = p.parseInfixExpression
	inFixFns[GT] = p.parseInfixExpression
	inFixFns[LPAREN] = p.parseCall
	inFixFns[ASSING] = p.parseReassigment
	inFixFns[LBRACKET] = p.parseCallList
	inFixFns[MOD] = p.parseInfixExpression
	inFixFns[AND] = p.parseInfixExpression
	inFixFns[OR] = p.parseInfixExpression
	return inFixFns
}

// register all prefix functions for the different token types
func (p *Parser) registerPrefixFns() PrefixParsFns {
	prefixFns := make(PrefixParsFns)
	prefixFns[FALSE] = p.parseBoolean
	prefixFns[FOR] = p.parseFor
	prefixFns[FUNCTION] = p.parseFunction
	prefixFns[WHILE] = p.parseWhile
	prefixFns[IDENT] = p.parseIdentifier
	prefixFns[IF] = p.parseIf
	prefixFns[INT] = p.parseInteger
	prefixFns[LPAREN] = p.parseGroupExpression
	prefixFns[MINUS] = p.parsePrefixExpression
	prefixFns[NOT] = p.parsePrefixExpression
	prefixFns[TRUE] = p.parseBoolean
	prefixFns[STRING] = p.parseStringLiteral
	prefixFns[DATASTRCUT] = p.ParseArray
	return prefixFns
}
