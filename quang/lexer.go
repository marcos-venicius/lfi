package quang

import (
	"errors"
	"fmt"
)

type lexer_token_t struct {
	value string
	kind  lexer_token_kind_t
}

type lexer_t struct {
	content     string
	cursor, bot int
	tokens      []lexer_token_t
}

type lexer_token_kind_t int

const (
	ltk_open_parenthesis  lexer_token_kind_t = iota
	ltk_close_parenthesis lexer_token_kind_t = iota
	ltk_and_keyword       lexer_token_kind_t = iota
	ltk_or_keyword        lexer_token_kind_t = iota
	ltk_eq_keyword        lexer_token_kind_t = iota
	ltk_ne_keyword        lexer_token_kind_t = iota
	ltk_gt_keyword        lexer_token_kind_t = iota
	ltk_lt_keyword        lexer_token_kind_t = iota
	ltk_gte_keyword       lexer_token_kind_t = iota
	ltk_lte_keyword       lexer_token_kind_t = iota
	ltk_symbol            lexer_token_kind_t = iota
	ltk_integer           lexer_token_kind_t = iota
	ltk_atom              lexer_token_kind_t = iota
)

var keywords = map[string]lexer_token_kind_t{
	"and": ltk_and_keyword,
	"or":  ltk_or_keyword,
	"eq":  ltk_eq_keyword,
	"ne":  ltk_ne_keyword,
	"gt":  ltk_gt_keyword,
	"lt":  ltk_lt_keyword,
	"gte": ltk_gte_keyword,
	"lte": ltk_lte_keyword,
}

func createLexer(content string) lexer_t {
	return lexer_t{
		content: content,
		cursor:  0,
		bot:     0,
		tokens:  make([]lexer_token_t, 0, 512),
	}
}

func (l *lexer_t) trimWhitespaces() {
	for l.cursor < len(l.content) && l.content[l.cursor] == ' ' {
		l.cursor++
	}
}

func (l *lexer_t) lexSingleChar(kind lexer_token_kind_t) {
	l.cursor++

	token := lexer_token_t{
		value: l.content[l.bot:l.cursor],
		kind:  kind,
	}

	l.tokens = append(l.tokens, token)
}

func (l *lexer_t) lexAtom() error {
	l.cursor++

	atomSize := 0

	for l.cursor < len(l.content) && isLetter(l.content[l.cursor]) {
		l.cursor++
		atomSize++
	}

	if atomSize == 0 {
		return errors.New(fmt.Sprintf("error: unterminated atom at position %d", l.cursor))
	}

	token := lexer_token_t{
		kind:  ltk_atom,
		value: l.content[l.bot:l.cursor],
	}

	l.tokens = append(l.tokens, token)

	return nil
}

func (l *lexer_t) lexIntegers() {
	for l.cursor < len(l.content) && isDigit(l.content[l.cursor]) {
		l.cursor++
	}

	token := lexer_token_t{
		kind:  ltk_integer,
		value: l.content[l.bot:l.cursor],
	}

	l.tokens = append(l.tokens, token)
}

func (l *lexer_t) lexSymbol() {
	for l.cursor < len(l.content) && isLetter(l.content[l.cursor]) {
		l.cursor++
	}

	token := lexer_token_t{
		value: l.content[l.bot:l.cursor],
		kind:  ltk_symbol,
	}

	if kind, ok := keywords[token.value]; ok {
		token.kind = kind
	}

	l.tokens = append(l.tokens, token)
}

func (l *lexer_t) lex() error {
	for l.cursor < len(l.content) {
		l.trimWhitespaces()

		if l.cursor >= len(l.content) {
			break
		}

		l.bot = l.cursor

		char := l.content[l.cursor]

		switch char {
		case '(':
			l.lexSingleChar(ltk_open_parenthesis)
		case ')':
			l.lexSingleChar(ltk_close_parenthesis)
		case ':':
			if err := l.lexAtom(); err != nil {
				return err
			}
		default:
			if isDigit(char) {
				l.lexIntegers()
			} else if isLetter(char) {
				l.lexSymbol()
			} else {
				return errors.New(fmt.Sprintf("error: unrecognized character \"%c\" at position %d", char, l.cursor+1))
			}
		}
	}

	return nil
}
