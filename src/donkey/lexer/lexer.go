package lexer

import (
	"donkey/token"
	"unicode"
	"unicode/utf8"
)

type Lexer struct {
	input   string
	pos     int  // current position in input (points to current char)
	readPos int  // current reading position in input (after current char)
	char    rune // current character under examination
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	stepSize := 1

	if l.readPos >= len(l.input) {
		l.char = 0
	} else {
		char, size := utf8.DecodeRuneInString(l.input[l.readPos:])
		l.char = char
		stepSize = size
	}
	l.pos = l.readPos
	l.readPos += stepSize
}

func (l *Lexer) peekChar() rune {
	if l.readPos >= len(l.input) {
		return 0
	} else {
		r, _ := utf8.DecodeRuneInString(l.input[l.readPos:])
		return r
	}
}

func newToken(tokenType token.TokenType, char rune) token.Token {
	return token.Token{Type: tokenType, Literal: string(char)}
}

func (l *Lexer) newTwoCharToken(tokenType token.TokenType) token.Token {
	char := l.char
	l.readChar()
	lit := string(char) + string(l.char)
	return token.Token{Type: tokenType, Literal: lit}
}

func (l *Lexer) readIdentifier() string {
	pos := l.pos
	for isValidIdentifierChar(l.char) {
		l.readChar()
	}
	return l.input[pos:l.pos]
}

func (l *Lexer) readNumber() string {
	pos := l.pos
	for unicode.IsDigit(l.char) {
		l.readChar()
	}
	return l.input[pos:l.pos]
}

// TODO add support for char escaping "hello \" test" "hello\t\n\ttest"
func (l *Lexer) readString() string {
	pos := l.pos + 1
	for {
		l.readChar()

		if l.char == '"' || l.char == 0 {
			break
		}
	}
	return l.input[pos:l.pos]
}

func (l *Lexer) skipWhitespace() {
	for unicode.IsSpace(l.char) {
		l.readChar()
	}
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()
	switch l.char {
	case '=':
		if l.peekChar() == '=' {
			tok = l.newTwoCharToken(token.EQ)
		} else {
			tok = newToken(token.ASSIGN, l.char)
		}
	case '+':
		tok = newToken(token.PLUS, l.char)
	case '-':
		tok = newToken(token.MINUS, l.char)
	case '!':
		if l.peekChar() == '=' {
			tok = l.newTwoCharToken(token.NOT_EQ)
		} else {
			tok = newToken(token.BANG, l.char)
		}
	case '/':
		tok = newToken(token.SLASH, l.char)
	case '*':
		tok = newToken(token.ASTERISK, l.char)
	case '<':
		if l.peekChar() == '=' {
			tok = l.newTwoCharToken(token.LT_EQ)
		} else {
			tok = newToken(token.LT, l.char)
		}
	case '>':
		if l.peekChar() == '=' {
			tok = l.newTwoCharToken(token.GT_EQ)
		} else {
			tok = newToken(token.GT, l.char)
		}
	case ';':
		tok = newToken(token.SEMICOLON, l.char)
	case ':':
		tok = newToken(token.COLON, l.char)
	case ',':
		tok = newToken(token.COMMA, l.char)
	case '(':
		tok = newToken(token.LPAREN, l.char)
	case ')':
		tok = newToken(token.RPAREN, l.char)
	case '{':
		tok = newToken(token.LBRACE, l.char)
	case '}':
		tok = newToken(token.RBRACE, l.char)
	case '[':
		tok = newToken(token.LBRACKET, l.char)
	case ']':
		tok = newToken(token.RBRACKET, l.char)
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isValidIdentifierChar(l.char) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if unicode.IsDigit(l.char) {
			tok.Literal = l.readNumber()
			tok.Type = token.INT
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.char)
		}
	}

	l.readChar()
	return tok
}

func isValidIdentifierChar(char rune) bool {
	return unicode.IsLetter(char) || hasEmoji(int(char))
}

func hasEmoji(charCode int) bool {
	rangeMin := '\U0001F300'
	rangeMax := '\U0001FAF6'
	rangeMin2 := 126980
	rangeMax2 := 127569
	rangeMin3 := 169
	rangeMax3 := 174
	rangeMin4 := 8205
	rangeMax4 := 12953

	if (charCode >= int(rangeMin) && charCode <= int(rangeMax)) ||
		(charCode >= rangeMin2 && charCode <= rangeMax2) ||
		(charCode >= rangeMin3 && charCode <= rangeMax3) ||
		(charCode >= rangeMin4 && charCode <= rangeMax4) {
		return true
	}

	return false
}
