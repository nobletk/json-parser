package parser

import (
	"fmt"
	"log/slog"
	"regexp"
	"strconv"

	"github.com/nobletk/json-parser/internal/ast"
	"github.com/nobletk/json-parser/internal/lexer"
	"github.com/nobletk/json-parser/internal/token"
)

type (
	parseFn func() (ast.Element, *JSONErr)
)

type JSONErr struct {
	Msg string
	Pos token.Position
}

type Parser struct {
	lexer  *lexer.Lexer
	logger *slog.Logger

	prvToken  token.Token
	curToken  token.Token
	peekToken token.Token

	parseFnMap map[token.TokenType]parseFn

	JSONErr *JSONErr
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		logger:  l.Logger,
		lexer:   l,
		JSONErr: &JSONErr{},
	}

	p.parseFnMap = make(map[token.TokenType]parseFn)
	p.registerElement(token.STRING, p.parseString)
	p.registerElement(token.TRUE, p.parseBoolean)
	p.registerElement(token.FALSE, p.parseBoolean)
	p.registerElement(token.NULL, p.parseNull)
	p.registerElement(token.NUMBER, p.parseNumber)
	p.registerElement(token.LBRACKET, p.parseArray)
	p.registerElement(token.LBRACE, p.parseObject)

	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) ParseFile() (*ast.JSONFile, *JSONErr) {
	p.logger.Info("Parsing File:",
		"currentToken", p.curToken.Literal,
		"currentTokenType", p.curToken.Type,
	)

	jf := &ast.JSONFile{}
	jf.Elements = []ast.Element{}

	if !p.curTokenIs(token.LBRACE) && !p.curTokenIs(token.LBRACKET) {
		msg := fmt.Sprintf("Expected '{' or '[', got '%+v' instead\n", p.curToken.Type)
		return nil, &JSONErr{Msg: msg, Pos: p.curToken.Position}
	}

	for !p.curTokenIs(token.EOF) {
		elem, err := p.parseElement()
		if err != nil {
			p.logger.Info("Parsing File Stopped:", "jsonErr", err)
			return nil, err
		}

		jf.Elements = append(jf.Elements, elem)
		p.logger.Info("Adding Element to Elements", "elem", elem.String())

		p.nextToken()
	}

	p.logger.Info("Parsing File Complete:", "jsonFile", jf.String())
	return jf, nil
}

func (p *Parser) parseElement() (ast.Element, *JSONErr) {
	p.logger.Info("Parsing Element:",
		"currentToken", p.curToken.Literal,
		"currentTokenType", p.curToken.Type,
	)

	switch p.curToken.Type {
	case token.LBRACE:
		return p.parseObject()
	case token.LBRACKET:
		return p.parseArray()
	default:
		msg := fmt.Sprintf("Expected 'EOF', got '%+v' instead\n", p.curToken.Type)
		p.logger.Info("Illegal TokenType:",
			"currentToken", p.curToken.Literal,
			"currentTokenType", p.curToken.Type,
			"jsonError", p.JSONErr,
		)
		return nil, &JSONErr{Msg: msg, Pos: p.peekToken.Position}
	}
}

func (p *Parser) parseObject() (ast.Element, *JSONErr) {
	obj := &ast.Object{Token: p.curToken}
	p.logger.Info("Parsing Object:",
		"currentToken", p.curToken.Literal,
		"currentTokenType", p.curToken.Type,
	)

	obj.Pairs = make(map[ast.Element]ast.Element)

	if !p.peekTokenIs(token.STRING) && !p.peekTokenIs(token.RBRACE) {
		msg := fmt.Sprintf("Expected 'STRING', '}', got '%+v' instead\n", p.peekToken.Type)
		return nil, &JSONErr{Msg: msg, Pos: p.peekToken.Position}
	}

	for !p.peekTokenIs(token.RBRACE) {
		p.nextToken()

		prop, err := p.parseValue()
		if err != nil {
			return nil, err
		}

		if err := p.expectPeek(token.COLON); err != nil {
			return nil, err
		}

		if p.isDuplicateProperty(obj.Pairs, prop) {
			msg := fmt.Sprintf("Duplicate JSON property '%+v'\n", prop)
			return nil, &JSONErr{Msg: msg, Pos: p.curToken.Position}
		}

		p.nextToken()

		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		obj.Pairs[prop] = val

		if err := p.expectPeek(token.COMMA); !p.peekTokenIs(token.RBRACE) && err != nil {
			return nil, err
		}

		if p.curTokenIs(token.COMMA) && !p.peekTokenIs(token.STRING) {
			msg := fmt.Sprintf("Expected 'STRING', got '%v' instead\n", p.peekToken.Type)
			return nil, &JSONErr{Msg: msg, Pos: p.peekToken.Position}
		}
	}

	if err := p.expectPeek(token.RBRACE); err != nil {
		return nil, err
	}

	return obj, nil
}

func (p *Parser) parseValue() (ast.Element, *JSONErr) {
	parseFn := p.parseFnMap[p.curToken.Type]
	p.logger.Info("Parsing Value:",
		"currentToken", p.curToken.Literal,
		"currentTokenType", p.curToken.Type,
		"jsonError", p.JSONErr,
	)

	if parseFn == nil {
		p.logger.Info("Exiting Parsing Value, No Parsing Function Found:")
		err := p.noParseFnError(p.curToken)
		return nil, err
	}

	val, err := parseFn()
	if err != nil {
		p.logger.Info("Parsing Value Stopped:", "jsonErr", err)
		return nil, err
	}

	p.logger.Info("Parsing Value Completed:", "value", val.String())
	return val, nil
}

func (p *Parser) parseString() (ast.Element, *JSONErr) {
	str := p.curToken.Literal
	p.logger.Info("Parsing String:", "string", str)

	for i := 0; i < len(str); i++ {
		r := str[i]

		if r == '\\' && i+1 < len(str) {
			escLen, err := p.checkEscapedSequence(str[i:])
			if escLen > 0 && err == nil {
				i += escLen - 1
				continue
			}
			p.logger.Info("Parsing String Stopped:", "escLen", escLen)
			return nil, err
		}
	}
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}, nil
}

func (p *Parser) parseBoolean() (ast.Element, *JSONErr) {
	return &ast.Boolean{
		Token: p.curToken,
		Value: p.curTokenIs(token.TRUE),
	}, nil
}

func (p *Parser) parseNull() (ast.Element, *JSONErr) {
	return &ast.Null{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}, nil
}

func (p *Parser) parseNumber() (ast.Element, *JSONErr) {
	num := &ast.NumberLiteral{Token: p.curToken}
	p.logger.Info("Parsing Number:", "num", num)

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("Failed parsing %q as a float\n", p.curToken.Literal)
		return nil, &JSONErr{Msg: msg, Pos: p.curToken.Position}
	}

	num.Value = value

	p.logger.Info("Parsing Number Completed:", "num", num)
	return num, nil
}

func (p *Parser) parseArray() (ast.Element, *JSONErr) {
	array := &ast.ArrayLiteral{Token: p.curToken}
	p.logger.Info("Parsing Array Started:")

	var err *JSONErr
	array.Elements, err = p.parseArrayList(token.RBRACKET)
	if err != nil {
		p.logger.Info("Parsing Array Stopped:", "jsonError", err)
		return nil, err
	}

	p.logger.Info("Parsing Array Completed:",
		"array", array.Elements,
		"currentToken", array.Token.Literal,
		"currentTokenType", array.Token.Type,
	)
	return array, nil
}

func (p *Parser) parseArrayList(end token.TokenType) ([]ast.Element, *JSONErr) {
	list := []ast.Element{}
	p.logger.Info("Parsing Array List Started:")

	if p.peekTokenIs(end) {
		p.nextToken()
		p.logger.Info("Empty Array:")
		return list, nil
	}

	p.logger.Info("Parsing Array List:",
		"endTokenType", end,
		"endFound", false,
		"list", list,
	)

	err := p.consumeAndParseValue(&list)
	if err != nil {
		return nil, err
	}

	for p.peekTokenIs(token.COMMA) {
		p.logger.Info("Parsing Array List and peekTokenIs Comma")
		p.nextToken()

		if p.peekTokenIs(end) {
			msg := fmt.Sprintf("Expected 'STRING', 'NUMBER', 'TRUE', 'FALSE', 'NULL'. got '%v' instead\n",
				p.peekToken.Type)
			return nil, &JSONErr{Msg: msg, Pos: p.peekToken.Position}
		}

		err := p.consumeAndParseValue(&list)
		if err != nil {
			return nil, err
		}
	}

	if !p.peekTokenIs(end) && !p.curTokenIs(token.COMMA) {
		msg := fmt.Sprintf("Expected ',', ']'. got '%v' instead\n", p.peekToken.Type)
		return nil, &JSONErr{Msg: msg, Pos: p.peekToken.Position}
	}
	p.nextToken()

	p.logger.Info("Parsing Array List Completed:", "list", list, "jsonError", p.JSONErr)
	return list, nil
}

func (p *Parser) consumeAndParseValue(list *[]ast.Element) *JSONErr {
	p.nextToken()
	val, err := p.parseValue()
	if err != nil {
		return err
	}
	*list = append(*list, val)
	p.logger.Info("ConsumeAndParseValue:", "elem", val.String())
	return nil
}

func (p *Parser) nextToken() {
	p.prvToken = p.curToken
	p.curToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
	p.logger.Info("Fetching New Token:",
		"prevToken", p.prvToken.Literal,
		"prevTokenType", p.prvToken.Type,
		"prevTokenPos", p.prvToken.Position,
		"currentToken", p.curToken.Literal,
		"currentTokenType", p.curToken.Type,
		"currentTokenPos", p.curToken.Position,
		"peekTokenPos", p.peekToken.Position,
		"peekToken", p.peekToken.Literal,
		"peekTokenType", p.peekToken.Type,
	)
}

func (p *Parser) expectPeek(t token.TokenType) *JSONErr {
	if ok := p.peekTokenIs(t); ok {
		p.logger.Info("Checking nextTokenType:", string(t), ok)
		p.nextToken()
		return nil
	} else {
		p.logger.Info("Checking nextTokenType:", string(t), ok)
		return p.peekError(t)
	}
}

func (p *Parser) peekError(t token.TokenType) *JSONErr {
	msg := fmt.Sprintf("Expected '%s', got '%v' instead\n",
		t, p.peekToken.Type)

	return &JSONErr{Msg: msg, Pos: p.peekToken.Position}
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) registerElement(tokenType token.TokenType, fn parseFn) {
	p.parseFnMap[tokenType] = fn
}

func (p *Parser) noParseFnError(t token.Token) *JSONErr {
	msg := ""

	switch p.prvToken.Type {
	case token.COLON:
		msg = fmt.Sprintf("Expected 'STRING', 'NUMBER', 'NULL', 'TRUE', 'FALSE', '{', '[', got '%v' instead\n",
			t.Type)
	case token.LBRACKET:
		msg = fmt.Sprintf("Expected 'STRING', 'NUMBER', 'NULL', 'TRUE', 'FALSE', '{', '[', ']', got '%v' instead\n",
			t.Type)
	case token.COMMA:
		msg = fmt.Sprintf("Expected 'STRING', 'NUMBER', 'NULL', 'TRUE', 'FALSE', '{', '[', got '%v' instead\n",
			t.Type)
	default:
		msg = fmt.Sprintf("Unexpected token found '%s'\n",
			t.Type)
	}
	return &JSONErr{Msg: msg, Pos: p.curToken.Position}
}

func (p *Parser) isDuplicateProperty(propMap map[ast.Element]ast.Element, prop ast.Element) bool {
	for p := range propMap {
		if prop.String() == p.String() {
			return true
		}
	}
	return false
}

func (p *Parser) checkNumberFormat(n ast.Element) (ast.Element, *JSONErr) {
	p.logger.Info("Checking Number Format:", "number", n.String())
	var pattern = regexp.MustCompile(`^[-+]?(([1-9][0-9]*)|0)(\.[0-9]+)?([eE][-+]?[0-9]+)?$`)
	matched := pattern.MatchString(n.String())
	if !matched {
		msg := fmt.Sprintf("Invalid number format '%s'\n", n.String())
		return nil, &JSONErr{Msg: msg, Pos: p.curToken.Position}
	}

	return n, nil
}

func (p *Parser) checkEscapedSequence(str string) (int, *JSONErr) {
	switch str[1] {
	case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
		p.logger.Info("Checking Escapped Sequence in String:", "sequence", string(str[1]))
		return 2, nil
	case 'u':
		p.logger.Info("Checking Unicode Escapped Sequence in String:")
		if len(str) >= 6 && p.isValidHexSequence(str[2:6]) {
			return 6, nil
		}
		msg := "Invalid unicode escape sequence\n"
		p.logger.Info("Failed Checking Escapped Sequence:", "error", p.JSONErr)
		return 0, &JSONErr{Msg: msg, Pos: p.curToken.Position}
	default:
		msg := fmt.Sprintf("Invalid escape sequence\n")
		p.logger.Info("Failed Checking Escapped Sequence:", "error", p.JSONErr)
		return 0, &JSONErr{Msg: msg, Pos: p.curToken.Position}
	}
}

func (p *Parser) isValidHexSequence(seq string) bool {
	if len(seq) != 4 {
		return false
	}
	for _, r := range seq {
		if !p.isHexDigit(r) {
			return false
		}
	}
	return true
}

func (p *Parser) isHexDigit(r rune) bool {
	return 'a' <= r && r <= 'f' || 'A' <= r && r <= 'F' || '0' <= r && r <= '9'
}
