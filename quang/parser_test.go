package quang

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseComparison(t *testing.T) {
	data := "size eq 0"

	l := createLexer(data)
	err := l.lex()

	assert.Nil(t, err)

	variables := map[string]variable_t{
		"size": {dtype: type_integer, integer: 10},
	}

	p := createParser(l.tokens, variables)
	expr := p.parseComparison()

	assert.Equal(t, expr_binary, expr.kind)
	assert.Equal(t, eo_eq, expr.binary.operator)
	assert.Equal(t, expr_number, expr.binary.left.kind)
	assert.Equal(t, expr_number, expr.binary.right.kind)
}

func TestParseTerm(t *testing.T) {
	data := "size gte 0 and size lte 1024"

	l := createLexer(data)
	err := l.lex()

	assert.Nil(t, err)

	variables := map[string]variable_t{
		"size": {dtype: type_integer, integer: 10},
	}

	p := createParser(l.tokens, variables)
	expr := p.parseTerm()

	assert.Equal(t, expr_binary, expr.kind)
	assert.Equal(t, eo_and, expr.binary.operator)

	left := expr.binary.left

	assert.Equal(t, expr_binary, left.kind)

	assert.Equal(t, expr_number, left.binary.left.kind)
	assert.Equal(t, eo_gte, left.binary.operator)
	assert.Equal(t, expr_number, left.binary.right.kind)

	right := expr.binary.right

	assert.Equal(t, expr_binary, right.kind)

	assert.Equal(t, expr_number, right.binary.right.kind)
	assert.Equal(t, eo_lte, right.binary.operator)
	assert.Equal(t, expr_number, right.binary.right.kind)
}

func TestParseExpression(t *testing.T) {
  data := "(method eq :get and size gt 0 and size lte 1024) or (method eq :post and status ne 204)"

	l := createLexer(data)
	err := l.lex()

	assert.Nil(t, err)

	variables := map[string]variable_t{
		"size":   {dtype: type_integer, integer: 10},
		"method": {dtype: type_atom, atom: "post"},
		"status": {dtype: type_integer, integer: 204},
	}

	p := createParser(l.tokens, variables)
	expr := p.parseExpression()

	assert.NotNil(t, expr)
}
