package lexer

import (
	"donkey/token"
	"unicode"
	"unicode/utf8"
)

type Lexer struct {
	input       string
	pos         int  // current position in input (points to current char)
	readPos     int  // current reading position in input (after current char)
	char        rune // current character under examination
	Line        int  // current reading line
	Column      int  // current column in the current line
	StartColumn int  // current column in the current line
}

func New(input string) *Lexer {
	l := &Lexer{input: input, Line: 1}
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
	l.Column += 1
}

func (l *Lexer) peekChar() rune {
	if l.readPos >= len(l.input) {
		return 0
	} else {
		r, _ := utf8.DecodeRuneInString(l.input[l.readPos:])
		return r
	}
}

func (l *Lexer) newTok(tokenType token.TokenType, literal string) token.Token {
	return token.Token{Type: tokenType, Literal: literal, Line: l.Line, Column: l.StartColumn}
}

func (l *Lexer) newToken(tokenType token.TokenType, char rune) token.Token {
	return l.newTok(tokenType, string(char))
}

func (l *Lexer) newTwoCharToken(tokenType token.TokenType) token.Token {
	char := l.char
	l.readChar()
	lit := string(char) + string(l.char)
	return l.newTok(tokenType, lit)
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
		if l.char == '\n' {
			// reset new line
			l.Line = l.Line + 1
			l.Column = 0
		}

		l.readChar()
	}
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()
	l.StartColumn = l.Column

	switch l.char {
	case '=':
		if l.peekChar() == '=' {
			tok = l.newTwoCharToken(token.EQ)
		} else {
			tok = l.newToken(token.ASSIGN, l.char)
		}
	case '+':
		tok = l.newToken(token.PLUS, l.char)
	case '-':
		tok = l.newToken(token.MINUS, l.char)
	case '!':
		if l.peekChar() == '=' {
			tok = l.newTwoCharToken(token.NOT_EQ)
		} else {
			tok = l.newToken(token.BANG, l.char)
		}
	case '/':
		tok = l.newToken(token.SLASH, l.char)
	case '*':
		tok = l.newToken(token.ASTERISK, l.char)
	case '<':
		if l.peekChar() == '=' {
			tok = l.newTwoCharToken(token.LT_EQ)
		} else {
			tok = l.newToken(token.LT, l.char)
		}
	case '>':
		if l.peekChar() == '=' {
			tok = l.newTwoCharToken(token.GT_EQ)
		} else {
			tok = l.newToken(token.GT, l.char)
		}
	case ';':
		tok = l.newToken(token.SEMICOLON, l.char)
	case ':':
		tok = l.newToken(token.COLON, l.char)
	case ',':
		tok = l.newToken(token.COMMA, l.char)
	case '(':
		tok = l.newToken(token.LPAREN, l.char)
	case ')':
		tok = l.newToken(token.RPAREN, l.char)
	case '{':
		tok = l.newToken(token.LBRACE, l.char)
	case '}':
		tok = l.newToken(token.RBRACE, l.char)
	case '[':
		tok = l.newToken(token.LBRACKET, l.char)
	case ']':
		tok = l.newToken(token.RBRACKET, l.char)
	case '"':
		tok = l.newTok(token.STRING, l.readString())
	case 0:
		tok = l.newTok(token.EOF, "")
	default:
		if isValidIdentifierChar(l.char) {
			ident := l.readIdentifier()
			tT := token.LookupIdent(ident)
			return l.newTok(tT, ident)
		} else if unicode.IsDigit(l.char) {
			tL := l.readNumber()
			return l.newTok(token.INT, tL)
		} else {
			tok = l.newToken(token.ILLEGAL, l.char)
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
