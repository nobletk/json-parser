package lexer

import (
	"log/slog"
	"regexp"

	"github.com/nobletk/json-parser/internal/token"
)

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
	line         int
	column       int
	Logger       *slog.Logger
}

func New(logger *slog.Logger, input string) *Lexer {
	l := &Lexer{
		input:  input,
		Logger: logger,
		line:   1,
		column: 0,
	}
	l.readChar()

	return l
}

func newToken(tokenType token.TokenType, ch byte, pos token.Position) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch), Position: pos}
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()
	pos := token.Position{Line: l.line, Column: l.column}

	switch l.ch {
	case '{':
		tok = newToken(token.LBRACE, l.ch, pos)
	case '}':
		tok = newToken(token.RBRACE, l.ch, pos)
	case '[':
		tok = newToken(token.LBRACKET, l.ch, pos)
	case ']':
		tok = newToken(token.RBRACKET, l.ch, pos)
	case ',':
		tok = newToken(token.COMMA, l.ch, pos)
	case ':':
		tok = newToken(token.COLON, l.ch, pos)
	case '"':
		tok = l.readString()
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
		tok.Position = pos
	default:
		if l.isLetter(l.ch) {
			l.Logger.Info("NextToken isLetter default:")
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			tok.Position = pos
			return tok
		}

		if l.ch == '-' || l.isDigit(l.ch) {
			l.Logger.Info("NextToken isDigit default:")
			tok = l.readNumber()
			return tok
		}

		tok = newToken(token.ILLEGAL, l.ch, pos)
	}

	l.readChar()

	return tok
}

func (l *Lexer) readString() token.Token {
	startPos := token.Position{Line: l.line, Column: l.column}
	start := l.position + 1
	l.Logger.Info("Reading String Start:",
		"curChar", string(l.ch),
		"curPosition", l.position,
		"peekChar", string(l.peekChar()),
		"peekCharPosition", l.position+1,
	)

ReadLoop:
	for {
		l.readChar()
		prvCh := l.input[l.position-1]

		l.Logger.Info("Reading String Loop:",
			"prevChar", string(prvCh),
			"prevPos", l.position-1,
			"curChar", string(l.ch),
			"curPosition", l.position,
			"peekChar", string(l.peekChar()),
			"peekCharPosition", l.position+1,
		)

		switch l.ch {
		case '"':
			backslashCount := 0
			for i := l.position - 1; i >= 0 && l.input[i] == '\\'; i-- {
				backslashCount++
			}

			if backslashCount%2 == 0 {
				l.Logger.Info("Reading String Stopped Closing Quote:",
					"prevChar", string(prvCh),
					"curChar", string(l.ch),
					"peekChar", string(l.peekChar()),
				)
				break ReadLoop
			}
		// case '\n', '\r':
		// 	return token.Token{
		// 		Type:     token.ILLEGAL,
		// 		Literal:  l.input[start:l.position],
		// 		Position: startPos,
		// 	}
		// case 0:
		// 	l.Logger.Info("Reading String Stopped EOF:",
		// 		"prevChar", string(prvCh),
		// 		"curChar", string(l.ch),
		// 		"peekChar", string(l.peekChar()),
		// 	)
		// 	if prvCh != '"' {
		// 		return token.Token{
		// 			Type:     token.ILLEGAL,
		// 			Literal:  l.input[start:l.position],
		// 			Position: startPos,
		// 		}
		// 	}
		// 	break ReadLoop
		default:
			if l.ch >= 0 && l.ch <= 31 {
				return token.Token{
					Type:     token.ILLEGAL,
					Literal:  l.input[start:l.position],
					Position: startPos,
				}
			}
		}
	}

	return token.Token{
		Type:     token.STRING,
		Literal:  l.input[start:l.position],
		Position: startPos,
	}
}

func (l *Lexer) readNumber() token.Token {
	startPos := token.Position{Line: l.line, Column: l.column}
	start := l.position
	l.Logger.Info("Reading Number Started:",
		"curChar", string(l.ch),
		"curPosition", l.position,
		"peekChar", string(l.peekChar()),
		"peekCharPosition", start+1,
	)

	for {
		switch l.peekChar() {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.', '-', '+', 'e', 'E':
			l.readChar()
			l.Logger.Info("Reading Number Main Case:",
				"curChar", string(l.ch),
				"curPosition", l.position,
				"peekChar", string(l.peekChar()),
				"peekCharPosition", l.position+1,
			)
		default:
			numberStr := l.input[start : l.position+1]
			l.Logger.Info("numberStr", "start", start, "end", l.position+1, "inputLen",
				len(l.input), "input", l.input, "numberStr", numberStr)
			numberRegex := regexp.MustCompile(`^[-]?(([1-9][0-9]*)|0)(\.[0-9]+)?([eE][-+]?[0-9]+)?$`)
			l.readChar()
			if numberRegex.MatchString(numberStr) {
				l.Logger.Info("Reading Number Completed:",
					"tokenType", token.NUMBER,
					"literal", numberStr,
					"pos", startPos,
				)
				return token.Token{
					Type:     token.NUMBER,
					Literal:  numberStr,
					Position: startPos,
				}
			}

			l.Logger.Info("Reading Number Stopped:",
				"tokenType", token.ILLEGAL,
				"literal", numberStr,
				"pos", startPos,
			)
			return token.Token{
				Type:     token.ILLEGAL,
				Literal:  numberStr,
				Position: startPos,
			}
		}
	}
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}

	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}

	l.Logger.Info("Character Read:",
		"char", string(l.ch),
		"ascii", l.ch,
		"position", l.readPosition,
	)

	l.position = l.readPosition
	l.readPosition++
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		return l.input[l.readPosition]
	}
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) readIdentifier() string {
	position := l.position

	for l.isLetter(l.ch) {
		l.readChar()
	}

	return l.input[position:l.position]
}

func (l *Lexer) isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func (l *Lexer) isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
