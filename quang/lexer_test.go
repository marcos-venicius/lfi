package quang

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrimWhitespaces(t *testing.T) {
	lexer := createLexer("     ")

	lexer.trimWhitespaces()

	assert.Equal(t, 5, lexer.cursor)
	assert.Equal(t, 0, lexer.bot)
}

func TestLexParenthesis(t *testing.T) {
	lexer := createLexer("   () (())   ")

	lexer.lex()

	assert.Equal(t, 6, len(lexer.tokens))
	assert.Equal(t, ltk_open_parenthesis, lexer.tokens[0].kind)
	assert.Equal(t, ltk_close_parenthesis, lexer.tokens[1].kind)
	assert.Equal(t, ltk_open_parenthesis, lexer.tokens[2].kind)
	assert.Equal(t, ltk_open_parenthesis, lexer.tokens[3].kind)
	assert.Equal(t, ltk_close_parenthesis, lexer.tokens[4].kind)
	assert.Equal(t, ltk_close_parenthesis, lexer.tokens[5].kind)

	assert.Equal(t, "(", lexer.tokens[0].value)
	assert.Equal(t, ")", lexer.tokens[1].value)
	assert.Equal(t, "(", lexer.tokens[2].value)
	assert.Equal(t, "(", lexer.tokens[3].value)
	assert.Equal(t, ")", lexer.tokens[4].value)
	assert.Equal(t, ")", lexer.tokens[5].value)
}

func TestLexAtoms(t *testing.T) {
	lexer := createLexer(":a :helloWorld")

	err := lexer.lex()

	assert.Nil(t, err)
	assert.Equal(t, 2, len(lexer.tokens))
	assert.Equal(t, ":a", lexer.tokens[0].value)
	assert.Equal(t, ":helloWorld", lexer.tokens[1].value)
	assert.Equal(t, ltk_atom, lexer.tokens[0].kind)
	assert.Equal(t, ltk_atom, lexer.tokens[1].kind)

	lexer = createLexer("   :")

	err = lexer.lex()

	assert.NotNil(t, err)
	assert.Equal(t, "error: unterminated atom at position 4", err.Error())
}

func TestLexIntegers(t *testing.T) {
	l := createLexer("10 20 1234")

	err := l.lex()

	assert.Nil(t, err)
	assert.Equal(t, 3, len(l.tokens))
	assert.Equal(t, "10", l.tokens[0].value)
	assert.Equal(t, "20", l.tokens[1].value)
	assert.Equal(t, "1234", l.tokens[2].value)
	assert.Equal(t, ltk_integer, l.tokens[0].kind)
	assert.Equal(t, ltk_integer, l.tokens[1].kind)
	assert.Equal(t, ltk_integer, l.tokens[2].kind)
}

func TestLexSymbols(t *testing.T) {
	l := createLexer("hello eq ne gt gte lt lte")

	err := l.lex()

	assert.Nil(t, err)
	assert.Equal(t, 7, len(l.tokens))
	assert.Equal(t, "hello", l.tokens[0].value)
	assert.Equal(t, "eq", l.tokens[1].value)
	assert.Equal(t, "ne", l.tokens[2].value)
	assert.Equal(t, "gt", l.tokens[3].value)
	assert.Equal(t, "gte", l.tokens[4].value)
	assert.Equal(t, "lt", l.tokens[5].value)
	assert.Equal(t, "lte", l.tokens[6].value)

	assert.Equal(t, ltk_symbol, l.tokens[0].kind)
	assert.Equal(t, ltk_eq_keyword, l.tokens[1].kind)
	assert.Equal(t, ltk_ne_keyword, l.tokens[2].kind)
	assert.Equal(t, ltk_gt_keyword, l.tokens[3].kind)
	assert.Equal(t, ltk_gte_keyword, l.tokens[4].kind)
	assert.Equal(t, ltk_lt_keyword, l.tokens[5].kind)
	assert.Equal(t, ltk_lte_keyword, l.tokens[6].kind)
}

func TestLex(t *testing.T) {
	l := createLexer("(a eq ne gt lt gte lte 10 :get)")

	err := l.lex()

	assert.Nil(t, err)
	assert.Equal(t, 11, len(l.tokens))

	assert.Equal(t, "(", l.tokens[0].value)
	assert.Equal(t, "a", l.tokens[1].value)
	assert.Equal(t, "eq", l.tokens[2].value)
	assert.Equal(t, "ne", l.tokens[3].value)
	assert.Equal(t, "gt", l.tokens[4].value)
	assert.Equal(t, "lt", l.tokens[5].value)
	assert.Equal(t, "gte", l.tokens[6].value)
	assert.Equal(t, "lte", l.tokens[7].value)
	assert.Equal(t, "10", l.tokens[8].value)
	assert.Equal(t, ":get", l.tokens[9].value)
	assert.Equal(t, ")", l.tokens[10].value)

	assert.Equal(t, ltk_open_parenthesis, l.tokens[0].kind)
	assert.Equal(t, ltk_symbol, l.tokens[1].kind)
	assert.Equal(t, ltk_eq_keyword, l.tokens[2].kind)
	assert.Equal(t, ltk_ne_keyword, l.tokens[3].kind)
	assert.Equal(t, ltk_gt_keyword, l.tokens[4].kind)
	assert.Equal(t, ltk_lt_keyword, l.tokens[5].kind)
	assert.Equal(t, ltk_gte_keyword, l.tokens[6].kind)
	assert.Equal(t, ltk_lte_keyword, l.tokens[7].kind)
	assert.Equal(t, ltk_integer, l.tokens[8].kind)
	assert.Equal(t, ltk_atom, l.tokens[9].kind)
	assert.Equal(t, ltk_close_parenthesis, l.tokens[10].kind)
}
