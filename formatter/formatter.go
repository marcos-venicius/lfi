package formatter

import (
	"errors"
	"fmt"
)

type Formatter struct {
	labels map[string]struct{}
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func (f Formatter) parseLabel(format string, index int) (string, int, error) {
	j := index

	for i := index + 1; i < len(format); i += 1 {
		if !isAlpha(format[i]) {
			break
		}

		j = i
	}

	if j == index {
		return "", 0, errors.New(fmt.Sprintf("format has unterminated label at position %d", index+1))
	}

	label := format[index+1 : j+1]

	if _, ok := f.labels[label]; !ok {
		return "", 0, errors.New(fmt.Sprintf("invalid label %%%s", label))
	}

	return format[index : j+1], j + 1, nil
}

func parseSpaces(format string, index int) (string, int) {
	j := index

	for i := index + 1; i < len(format); i += 1 {
		if format[i] != ' ' {
			break
		}

		j = i
	}

	return format[index : j+1], j + 1
}

func parseString(format string, index int) (string, int, error) {
	end := index

	for i := index + 1; i < len(format); i += 1 {
		end = i

		if format[i] == '\\' {
			if i >= len(format)-1 {
				return "", 0, errors.New(fmt.Sprintf("format has an invalid scape sequence at position %d", i+1))
			}

			switch format[i+1] {
			case '\'':
				i += 1
				continue
			default:
				return "", 0, errors.New(fmt.Sprintf("format has an invalid scape sequence at position %d", i+1))
			}
		}

		if format[i] == '\'' {
			break
		}
	}

	if end-1 <= index {
		return "", 0, errors.New(fmt.Sprintf("format has unterminated string at position %d", index+1))
	}

	if format[end] != '\'' {
		return "", 0, errors.New(fmt.Sprintf("format has unterminated string at position %d", index+1))
	}

	return format[index : end+1], end + 1, nil
}

func parseScapeSequence(format string, index int) (string, int, error) {
	if index >= len(format)-1 {
		return "", 0, errors.New(fmt.Sprintf("format has an invalid scape sequence at position %d", index+1))
	}

	switch format[index+1] {
	case 'n':
		return "\n", index + 2, nil
	case 't':
		return "\t", index + 2, nil
	}

	return "", 0, errors.New(fmt.Sprintf("format has an invalid scape sequence at position %d", index+1))
}

func (f Formatter) ParseFormatString(format string) ([]string, error) {
	tokens := make([]string, 0)

	index := 0

	for index < len(format) {
		switch format[index] {
		case '%':
			label, nextIndex, err := f.parseLabel(format, index)

			if err != nil {
				return nil, err
			}

			tokens = append(tokens, label)

			index = nextIndex
		case ' ':
			spaces, nextIndex := parseSpaces(format, index)

			tokens = append(tokens, spaces)

			index = nextIndex
		case '\'':
			text, nextIndex, err := parseString(format, index)

			if err != nil {
				return nil, err
			}

			tokens = append(tokens, text)

			index = nextIndex
		case '\\':
			text, nextIndex, err := parseScapeSequence(format, index)

			if err != nil {
				return nil, err
			}

			tokens = append(tokens, text)
			index = nextIndex
		default:
			return nil, errors.New(fmt.Sprintf("unrecognized char: \"%c\" at position %d", format[index], index+1))
		}
	}

	return tokens, nil
}

func CreateFormatter(labels []string) Formatter {
	l := make(map[string]struct{})

	for _, label := range labels {
		l[label] = struct{}{}
	}

	return Formatter{
		labels: l,
	}
}
