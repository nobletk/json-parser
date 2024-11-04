package token

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	STRING = "STRING"
	NUMBER = "NUMBER"

	TRUE  = "TRUE"
	FALSE = "FALSE"
	NULL  = "NULL"

	COMMA = ","
	COLON = ":"

	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"
)

type Position struct {
	Line   int
	Column int
}

type TokenType string

type Token struct {
	Type     TokenType
	Literal  string
	Position Position
}

var keywords = map[string]TokenType{
	"true":  TRUE,
	"false": FALSE,
	"null":  NULL,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return ILLEGAL
}
