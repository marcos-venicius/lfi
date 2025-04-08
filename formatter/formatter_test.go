package formatter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnterminatedLabel(t *testing.T) {
	var fmt = CreateFormatter([]string{})

	str := "testing test %"
	_, _, err := fmt.parseLabel(str, strings.Index(str, "%"))

	assert.NotNil(t, err)

	str = "testing % test"
	_, _, err = fmt.parseLabel(str, strings.Index(str, "%"))

	assert.NotNil(t, err)

	str = "% testing test"
	_, _, err = fmt.parseLabel(str, strings.Index(str, "%"))

	assert.NotNil(t, err)
}

func TestInvalidLabel(t *testing.T) {
	var fmt = CreateFormatter([]string{})

	str := "%testing test"
	_, _, err := fmt.parseLabel(str, strings.Index(str, "%"))

	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "invalid label %testing")
}

func TestValidLabel(t *testing.T) {
	var fmt = CreateFormatter([]string{"testing"})

	str := "%testing test"
	label, nextIndex, err := fmt.parseLabel(str, strings.Index(str, "%"))

	assert.Nil(t, err)
	assert.Equal(t, "%testing", label)
	assert.Equal(t, 8, nextIndex)
}

func TestNextIndex(t *testing.T) {
	var fmt = CreateFormatter([]string{"testing", "test"})

	str := "%testing%test hello"
	label, nextIndex, err := fmt.parseLabel(str, strings.Index(str, "%"))

	assert.Nil(t, err)
	assert.Equal(t, "%testing", label)
	assert.Equal(t, 8, nextIndex)

	label, nextIndex, err = fmt.parseLabel(str, nextIndex)

	assert.Nil(t, err)
	assert.Equal(t, "%test", label)
	assert.Equal(t, 13, nextIndex)
}

func TestParseFormatString(t *testing.T) {
	var fmt = CreateFormatter([]string{"testing", "test"})

	str := "%testing    %test 'this is \\'my string\\''"

	tokens, err := fmt.ParseFormatString(str)

	assert.Nil(t, err)
	assert.Equal(t, 5, len(tokens))
	assert.Equal(t, "%testing", tokens[0])
	assert.Equal(t, "    ", tokens[1])
	assert.Equal(t, "%test", tokens[2])
	assert.Equal(t, " ", tokens[3])
	assert.Equal(t, "'this is \\'my string\\''", tokens[4])
}

func TestParseStringUnterminatedStrings(t *testing.T) {
	str := "'this is an unterminated string"

	_, _, err := parseString(str, strings.Index(str, "'"))

	assert.NotNil(t, err)
	assert.Equal(t, "format has unterminated string at position 1", err.Error())

	str = "this is an unterminated string'"

	_, _, err = parseString(str, strings.Index(str, "'"))

	assert.NotNil(t, err)
	assert.Equal(t, "format has unterminated string at position 31", err.Error())
}

func TestParseStringScapeSingleQuote(t *testing.T) {
	str := "'this is an \\'unterminated string\\''"

	text, _, err := parseString(str, strings.Index(str, "'"))

	assert.Nil(t, err)
	assert.Equal(t, "'this is an \\'unterminated string\\''", text)
}

func TestParseStringInvalidScapeSequence(t *testing.T) {
	str := "'testing \\a'"

	_, _, err := parseString(str, strings.Index(str, "'"))

	assert.NotNil(t, err)
	assert.Equal(t, "format has an invalid scape sequence at position 10", err.Error())

	str = "'testing \\"

	_, _, err = parseString(str, strings.Index(str, "'"))

	assert.NotNil(t, err)
	assert.Equal(t, "format has an invalid scape sequence at position 10", err.Error())
}

func TestParseScapeSequence(t *testing.T) {
	str := "this: test\\s"

	_, _, err := parseScapeSequence(str, strings.Index(str, "\\"))

	assert.NotNil(t, err)
	assert.Equal(t, "format has an invalid scape sequence at position 11", err.Error())

	str = "this: test\\"

	_, _, err = parseScapeSequence(str, strings.Index(str, "\\"))

	assert.NotNil(t, err)
	assert.Equal(t, "format has an invalid scape sequence at position 11", err.Error())

	str = "\\n"

	text, _, err := parseScapeSequence(str, strings.Index(str, "\\"))

	assert.Nil(t, err)
	assert.Equal(t, "\n", text)

	str = "\\t"

	text, _, err = parseScapeSequence(str, strings.Index(str, "\\"))

	assert.Nil(t, err)
	assert.Equal(t, "\t", text)
}
