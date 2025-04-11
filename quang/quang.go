package quang

import "regexp"

type Quang struct {
	expression *expression_t
	variables  map[string]variable_t
	atoms      map[string]int
	tokens     []lexer_token_t
}

func CreateQuang(expression string) (Quang, error) {
	quang := Quang{
		variables: make(map[string]variable_t),
		atoms:     make(map[string]int),
	}

	l := createLexer(expression)

	if err := l.lex(); err != nil {
		return quang, err
	}

	quang.tokens = l.tokens

	return quang, nil
}

func (q *Quang) AddAtom(name string, value int) {
	if _, ok := q.atoms[name]; ok {
		panic("you already have this atom")
	}

	q.atoms[name] = value
}

func (q *Quang) AddIntVar(name string, value int) *Quang {
	q.variables[name] = variable_t{
		dtype:   type_integer,
		integer: value,
	}

	return q
}

func (q *Quang) AddAtomVar(name string, value int) *Quang {
	q.variables[name] = variable_t{
		dtype: type_atom,
		atom:  value,
	}

	return q
}

func (q *Quang) AddStrVar(name string, value string) *Quang {
	q.variables[name] = variable_t{
		dtype:  type_string,
		string: value,
	}

	return q
}

func (q *Quang) ClearState() *Quang {
	q.variables = make(map[string]variable_t)
	q.atoms = make(map[string]int)

	return q
}

func cmpStrToStr(left *expression_t, op expression_operator_t, right *expression_t) bool {
	l := left.string
	r := right.string

	switch op {
  case eo_reg:
    return regexp.MustCompile(r).MatchString(l)
	case eo_eq:
		return l == r
	case eo_ne:
		return l != r
	case eo_lt:
		return l < r
	case eo_gt:
		return l > r
	case eo_lte:
		return l <= r
	case eo_gte:
		return l >= r
	}

	panic("unreacheable: invalid operator")
}

func cmpAtomToAtom(left *expression_t, op expression_operator_t, right *expression_t) bool {
	l := left.atom
	r := right.atom

	switch op {
  case eo_reg:
    panic("you cannot use regex operator in atoms")
	case eo_eq:
		return l == r
	case eo_ne:
		return l != r
	case eo_lt:
		return l < r
	case eo_gt:
		return l > r
	case eo_lte:
		return l <= r
	case eo_gte:
		return l >= r
	}

	panic("unreacheable: invalid operator")
}

func cmpIntToInt(left *expression_t, op expression_operator_t, right *expression_t) bool {
	l := left.number
	r := right.number

	switch op {
  case eo_reg:
    panic("you cannot use regex operator in integers")
	case eo_eq:
		return l == r
	case eo_ne:
		return l != r
	case eo_lt:
		return l < r
	case eo_gt:
		return l > r
	case eo_lte:
		return l <= r
	case eo_gte:
		return l >= r
	}

	panic("unreacheable: invalid operator")
}

func (q Quang) eval(expr *expression_t) bool {
	if expr.kind == expr_binary {
		if expr.binary.operator == eo_or {
			return q.eval(expr.binary.left) || q.eval(expr.binary.right)
		} else if expr.binary.operator == eo_and {
			return q.eval(expr.binary.left) && q.eval(expr.binary.right)
		}

		left := expr.binary.left
		right := expr.binary.right

		if left.kind == expr_number && right.kind == expr_number {
			return cmpIntToInt(left, expr.binary.operator, right)
		} else if left.kind == expr_atom && right.kind == expr_atom {
			return cmpAtomToAtom(left, expr.binary.operator, right)
		} else if left.kind == expr_string && right.kind == expr_string {
			return cmpStrToStr(left, expr.binary.operator, right)
    }

		panic("cannot compare different types")
	}

	panic("unreacheable")
}

// TODO: implement the parser with lazy eval to parse only once and then, when evaluating
//
//	load the variables and try to validate the types
func (q Quang) Eval() bool {
	p := createParser(q.tokens, q.variables, q.atoms)

	q.expression = p.parseExpression()

	if q.expression == nil {
		return true
	}

	return q.eval(q.expression)
}
