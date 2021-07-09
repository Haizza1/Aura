package parser

import (
	"aura/src/ast"
	l "aura/src/lexer"
)

// parse a method expression
func (p *Parser) parseMethod(left ast.Expression) ast.Expression {
	p.checkCurrentTokenIsNotNil()
	method := ast.NewMethodExpression(*p.currentToken, left, nil)
	if !p.expepectedToken(l.IDENT) {
		// syntax error. we dont allow this -> obj:();
		return nil
	}

	method.Method = p.parseExpression(LOWEST)
	return method
}

// parse an infix expressoin
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	p.checkCurrentTokenIsNotNil()
	infix := ast.Newinfix(*p.currentToken, nil, p.currentToken.Literal, left)
	precedence := p.currentPrecedence()
	p.advanceTokens()
	infix.Rigth = p.parseExpression(precedence)
	return infix
}

// parse a function call
func (p *Parser) parseCall(function ast.Expression) ast.Expression {
	p.checkCurrentTokenIsNotNil()
	call := ast.NewCall(*p.currentToken, function)
	call.Arguments = p.parseCallArguments()
	return call
}

// parse a call list expression
func (p *Parser) parseCallList(valueList ast.Expression) ast.Expression {
	p.checkCurrentTokenIsNotNil()
	callList := ast.NewCallList(*p.currentToken, valueList, nil)
	p.advanceTokens()
	callList.Index = p.parseExpression(LOWEST)
	if !p.expepectedToken(l.RBRACKET) {
		// syntax error. we dont allow tihs -> lista[2,3,4,5;
		return nil
	}

	return callList
}

// parse a ressigment expression
func (p *Parser) parseReassigment(ident ast.Expression) ast.Expression {
	p.checkCurrentTokenIsNotNil()
	reassignment := ast.NewReassignment(*p.currentToken, ident, nil)
	p.advanceTokens()
	reassignment.NewVal = p.parseExpression(LOWEST)
	return reassignment
}

// parse a key value expression
func (p *Parser) parseKeyValues() *ast.KeyValue {
	p.checkCurrentTokenIsNotNil()
	keyVal := ast.NewKeyVal(*p.currentToken, nil, nil)
	keyVal.Key = p.parseExpression(LOWEST)
	if !p.expepectedToken(l.ARROW) {
		return nil
	}

	p.advanceTokens()
	keyVal.Value = p.parseExpression(LOWEST)
	return keyVal
}

// parse a range expression
func (p *Parser) parseRangeExpression() ast.Expression {
	rangeExpress := ast.NewRange(*p.currentToken, nil, nil)
	if !p.expepectedToken(l.IDENT) {
		// syntax error. we dont allow this -> por(en rango(10))
		return nil
	}

	rangeExpress.Variable = p.parseIdentifier()
	if !p.expepectedToken(l.IN) {
		// syntax error. we dont allow this -> por(i rango(10))
		return nil
	}

	p.advanceTokens()
	rangeExpress.Range = p.parseExpression(LOWEST)
	return rangeExpress
}

// parse a class field or method call
func (p *Parser) parseClassFieldsCall(left ast.Expression) ast.Expression {
	call := ast.NewClassFieldCall(*p.currentToken, left, nil)
	p.checkPeekTokenIsNotNil()
	p.advanceTokens()
	call.Field = p.parseExpression(LOWEST)
	return call
}

// parse an assigment expression like
//		x := 5;
func (p *Parser) parseAssigmentExp(left ast.Expression) ast.Expression {
	ident, isIdent := left.(*ast.Identifier)
	if !isIdent {
		return nil
	}

	assigment := ast.NewAssigmentExp(*p.currentToken, ident, nil)
	p.advanceTokens()
	assigment.Val = p.parseExpression(LOWEST)
	return assigment
}

func (p *Parser) parseArrowFunc() ast.Expression {
	arrowFunc := ast.NewArrowFunc(*p.currentToken, make([]*ast.Identifier, 0), nil)
	arrowFunc.Params = p.parseArrowValues()
	if !p.expepectedToken(l.ARROW) {
		return nil
	}

	if p.peekToken.Token_type == l.LBRACE {
		p.advanceTokens()
		arrowFunc.Body = p.parseBlock()
	} else {
		p.advanceTokens()
		exp := p.parserExpressionStatement()
		arrowFunc.Body = &ast.Block{Staments: []ast.Stmt{exp}}
	}

	return arrowFunc
}

func (p *Parser) parseArrowValues() []*ast.Identifier {
	var values []*ast.Identifier
	if p.peekToken.Token_type == l.BAR {
		p.advanceTokens()
		return values
	}

	p.advanceTokens()
	if ident := p.parseIdentifier().(*ast.Identifier); ident != nil {
		values = append(values, ident)
	}

	if p.peekToken.Token_type == l.COMMA {
		p.advanceTokens()
		p.advanceTokens()
		if ident := p.parseIdentifier().(*ast.Identifier); ident != nil {
			values = append(values, ident)
		}
	}

	if !p.expepectedToken(l.BAR) {
		return make([]*ast.Identifier, 0)
	}

	return values
}
