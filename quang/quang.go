package quang

type QuangVariables struct {
	ip         string
	time       string
	method     string
	host       string
	resource   string
	version    string
	statusCode int
	size       int
	userAgent  string
}

type Quang struct {
	expression any // AST ready to eval
}

func CreateQuang(expression string) (Quang, error) {
	quang := Quang{}

	l := createLexer(expression)

	if err := l.lex(); err != nil {
		return quang, err
	}

	// parse
	// save the parsed structure to the `quange`

	return quang, nil
}

// TODO: implement the parser
func (q Quang) Eval(variables QuangVariables) bool {
	return false
}
