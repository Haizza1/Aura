package lpp

type TokenType int

const (
	Head TokenType = iota
	ASSING
	COMMA
	DIVISION
	ELSE
	EOF
	FALSE
	FUNCTION
	GT // grather than
	IDENT
	IF
	ILLEGAL
	INT
	LBRACE
	LET
	LPAREN
	LT    // less than
	MINUS // -
	NOT   // !
	PLUS
	RBRACE
	RETURN
	RPAREN
	SEMICOLON
	TIMES // *
	TRUE
)

type Token struct {
	Token_type TokenType
	Literal    string
}

// verify that given literal is a string
func LookUpTokenType(literal string) TokenType {
	keywords := map[string]TokenType{
		"falso":     FALSE,
		"funcion":   FUNCTION,
		"regresa":   RETURN,
		"si":        IF,
		"si_no":     ELSE,
		"var":       LET,
		"verdadero": TRUE,
	}

	TokenType, exists := keywords[literal]
	if exists {
		return TokenType
	} else {
		return IDENT
	}
}
