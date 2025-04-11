package quang

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type expression_kind_t int
type data_type_t int
type expression_operator_t int

type binary_expression_t struct {
	operator expression_operator_t

	left  *expression_t
	right *expression_t
}

type expression_t struct {
	kind expression_kind_t

	number int

	atom int

	string string

	binary *binary_expression_t
}

type variable_t struct {
	dtype data_type_t

	integer int
	atom    int
	string  string
}

type parser_t struct {
	current_token int
	tokens        []lexer_token_t
	variables     map[string]variable_t
	atoms         map[string]int
}

const (
	expr_number expression_kind_t = iota
	expr_atom
	expr_string
	expr_binary
	expr_comparison
)

const (
	type_integer data_type_t = iota
	type_atom
	type_string
)

const (
	eo_eq expression_operator_t = iota
	eo_reg
	eo_ne
	eo_gt
	eo_lt
	eo_gte
	eo_lte
	eo_and
	eo_or
)

func lexerTokenKindToExpressionOperator(kind lexer_token_kind_t) expression_operator_t {
	switch kind {
	case ltk_reg_keyword:
		return eo_reg
	case ltk_eq_keyword:
		return eo_eq
	case ltk_ne_keyword:
		return eo_ne
	case ltk_gt_keyword:
		return eo_gt
	case ltk_lt_keyword:
		return eo_lt
	case ltk_gte_keyword:
		return eo_gte
	case ltk_lte_keyword:
		return eo_lte
	case ltk_and_keyword:
		return eo_and
	case ltk_or_keyword:
		return eo_or
	}
	panic("invalid token kind")
}

func terror(err error) {
	fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
	os.Exit(1)
}

func parseInteger32(text string) int {
	number, err := strconv.ParseInt(text, 10, 32)

	if err != nil {
		terror(err)
	}

	return int(number)
}

func createParser(tokens []lexer_token_t, variables map[string]variable_t, atoms map[string]int) parser_t {
	return parser_t{
		current_token: 0,
		tokens:        tokens,
		variables:     variables,
		atoms:         atoms,
	}
}

func (p parser_t) isEmpty() bool {
	return p.current_token >= len(p.tokens)-1
}

func (p parser_t) token() lexer_token_t {
	return p.tokens[p.current_token]
}

func (p *parser_t) parsePrimary() *expression_t {
	current := p.token()

	p.current_token++

	if current.kind == ltk_integer {
		return &expression_t{
			kind:   expr_number,
			number: parseInteger32(current.value),
		}
	}

	if current.kind == ltk_symbol {
		if variable, ok := p.variables[current.value]; ok {
			switch variable.dtype {
			case type_integer:
				return &expression_t{
					kind:   expr_number,
					number: variable.integer,
				}
			case type_atom:
				return &expression_t{
					kind: expr_atom,
					atom: variable.atom,
				}
			case type_string:
				return &expression_t{
					kind:   expr_string,
					string: variable.string,
				}
			}
		}

		terror(errors.New(fmt.Sprintf("the variable \"%s\" does not exist", current.value)))
	}

	if current.kind == ltk_atom {
		if value, ok := p.atoms[current.value[1:]]; ok {
			return &expression_t{
				kind: expr_atom,
				atom: value,
			}
		}

		terror(fmt.Errorf("the atom \"%s\" does not exists", current.value))
	}

	if current.kind == ltk_string {
		return &expression_t{
			kind:   expr_string,
			string: current.value,
		}
	}

	terror(fmt.Errorf("invalid syntax. unexpected token kind: %s \"%s\"", current.kind.string(), current.value))

	return nil
}

func (p *parser_t) parseComparison() *expression_t {
	current := p.token()

	left := p.parsePrimary()

	current = p.token()

	switch current.kind {
	case ltk_eq_keyword, ltk_ne_keyword, ltk_gt_keyword, ltk_lt_keyword, ltk_gte_keyword, ltk_lte_keyword, ltk_reg_keyword:
		{
			p.current_token++
			// TODO: validate data types
			right := p.parsePrimary()

			return &expression_t{
				kind: expr_binary,
				binary: &binary_expression_t{
					operator: lexerTokenKindToExpressionOperator(current.kind),
					left:     left,
					right:    right,
				},
			}
		}
	}

	terror(fmt.Errorf("expected comparison operator after expression got \"%s\"", current.value))
	return nil
}

func (p *parser_t) parseFactor() *expression_t {
	current := p.token()

	if current.kind == ltk_open_parenthesis {
		p.current_token++

		expr := p.parseExpression()

		if p.token().kind != ltk_close_parenthesis {
			terror(fmt.Errorf("expected ')', got \"%s\"", p.token().value))
		}

		p.current_token++
		return expr
	}

	return p.parseComparison()
}

func (p *parser_t) parseTerm() *expression_t {
	left := p.parseFactor()

	for !p.isEmpty() {
		current := p.token()

		if current.kind != ltk_and_keyword {
			break
		}

		p.current_token++

		right := p.parseFactor()

		left = &expression_t{
			kind: expr_binary,
			binary: &binary_expression_t{
				operator: eo_and,
				left:     left,
				right:    right,
			},
		}
	}

	return left
}

func (p *parser_t) parseExpression() *expression_t {
	if p.isEmpty() {
		return nil
	}

	left := p.parseTerm()

	for !p.isEmpty() {
		current := p.token()

		if current.kind != ltk_or_keyword {
			break
		}

		p.current_token++

		right := p.parseTerm()

		left = &expression_t{
			kind: expr_binary,
			binary: &binary_expression_t{
				operator: eo_or,
				left:     left,
				right:    right,
			},
		}
	}

	return left
}
