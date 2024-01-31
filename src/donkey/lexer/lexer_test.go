package lexer

import (
	"testing"

	"donkey/token"
)

func TestNextToken(t *testing.T) {
	input := `
let 🥰 = 5;
let ten = 10;

let add = fn(x, y) {
  x + y;
};

let result = add(five, ten);
!-/*5;
5 < 10 > 5;

if (5 < 10) {
	return true;
} else {
	return false;
}

10 == 10;
10 != 9;
"foobar"
"foo bar"
[1, 2];
{"foo": "bar"}
macro(x, y) { x + y; };
import "file.dk";
`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.LET, "let"},
		{token.IDENT, "🥰"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "ten"},
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "add"},
		{token.ASSIGN, "="},
		{token.FUNCTION, "fn"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.IDENT, "y"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "result"},
		{token.ASSIGN, "="},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.IDENT, "five"},
		{token.COMMA, ","},
		{token.IDENT, "ten"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},
		{token.BANG, "!"},
		{token.MINUS, "-"},
		{token.SLASH, "/"},
		{token.ASTERISK, "*"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.GT, ">"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.TRUE, "true"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.ELSE, "else"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.FALSE, "false"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.INT, "10"},
		{token.EQ, "=="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.INT, "10"},
		{token.NOT_EQ, "!="},
		{token.INT, "9"},
		{token.SEMICOLON, ";"},
		{token.STRING, "foobar"},
		{token.STRING, "foo bar"},
		{token.LBRACKET, "["},
		{token.INT, "1"},
		{token.COMMA, ","},
		{token.INT, "2"},
		{token.RBRACKET, "]"},
		{token.SEMICOLON, ";"},
		{token.LBRACE, "{"},
		{token.STRING, "foo"},
		{token.COLON, ":"},
		{token.STRING, "bar"},
		{token.RBRACE, "}"},
		{token.MACRO, "macro"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.IDENT, "y"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},
		{token.IMPORT, "import"},
		{token.STRING, "file.dk"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestLexerLinesAndColumns(t *testing.T) {
	input := `
let 🥰 = 5;
let ten = 10;

let add = fn(x, y) {
x + y;
};

let result = add(five, ten);`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
		expectedLine    int
		expectedColumn  int
	}{
		{token.LET, "let", 2, 1},
		{token.IDENT, "🥰", 2, 5},
		{token.ASSIGN, "=", 2, 7},
		{token.INT, "5", 2, 9},
		{token.SEMICOLON, ";", 2, 10},
		{token.LET, "let", 3, 1},
		{token.IDENT, "ten", 3, 5},
		{token.ASSIGN, "=", 3, 9},
		{token.INT, "10", 3, 11},
		{token.SEMICOLON, ";", 3, 13},
		{token.LET, "let", 5, 1},
		{token.IDENT, "add", 5, 5},
		{token.ASSIGN, "=", 5, 9},
		{token.FUNCTION, "fn", 5, 11},
		{token.LPAREN, "(", 5, 13},
		{token.IDENT, "x", 5, 14},
		{token.COMMA, ",", 5, 15},
		{token.IDENT, "y", 5, 17},
		{token.RPAREN, ")", 5, 18},
		{token.LBRACE, "{", 5, 20},
		{token.IDENT, "x", 6, 1},
		{token.PLUS, "+", 6, 3},
		{token.IDENT, "y", 6, 5},
		{token.SEMICOLON, ";", 6, 6},
		{token.RBRACE, "}", 7, 1},
		{token.SEMICOLON, ";", 7, 2},
		{token.LET, "let", 9, 1},
		{token.IDENT, "result", 9, 5},
		{token.ASSIGN, "=", 9, 12},
		{token.IDENT, "add", 9, 14},
		{token.LPAREN, "(", 9, 17},
		{token.IDENT, "five", 9, 18},
		{token.COMMA, ",", 9, 22},
		{token.IDENT, "ten", 9, 24},
		{token.RPAREN, ")", 9, 27},
		{token.SEMICOLON, ";", 9, 28},
		{token.EOF, "", 9, 29},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}

		line := tok.Location.Line
		col := tok.Location.Column
		if line != tt.expectedLine {
			t.Fatalf("tests[%d] - line number wrong. expected=%d, got=%d",
				i, tt.expectedLine, line)
		}

		if col != tt.expectedColumn {
			t.Fatalf("tests[%d] - column number wrong. expected=%d, got=%d",
				i, tt.expectedColumn, col)
		}
	}
}

func TestAsyncToken(t *testing.T) {
	input := `
let add = async fn(x, y) {
x + y;
};

async fn test() {}
`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.LET, "let"},
		{token.IDENT, "add"},
		{token.ASSIGN, "="},
		{token.ASYNC, "async"},
		{token.FUNCTION, "fn"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.IDENT, "y"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},
		{token.ASYNC, "async"},
		{token.FUNCTION, "fn"},
		{token.IDENT, "test"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}
