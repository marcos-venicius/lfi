package main

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"
)

type order_t int

const (
	ORDER_IP order_t = iota
	ORDER_TIME
	ORDER_METHOD
	ORDER_RESOURCE
	ORDER_HTTP_VERSION
	ORDER_STATUS_CODE
	ORDER_REQUEST_SIZE
	ORDER_HOST
	ORDER_USER_AGENT
	ORDER_COUNT
)

const MAX_CONFIG_FILE_SIZE = 1024

type Configs struct {
	regex  *regexp.Regexp
	order  []order_t
	format string
}

var configFileName = ".lfi"
var defaultLogRegex = regexp.MustCompile(`^(\d{1,3}.\d{1,3}.\d{1,3}.\d{1,3}) .* .* \[(\d{1,2}\/\w+\/\d{4}:\d{2}:\d{2}:\d{2} \+\d{4})\] "(\w+) (.*?) (HTTP\/\d.\d)" (\d+|-) (\d+|-) "(.*?)" "(.*?)"$`)
var defaultFormatting = "%time %ip %method %resource %version %status %size %host %agent"
var defaultOrder = []order_t{
	ORDER_IP,
	ORDER_TIME,
	ORDER_METHOD,
	ORDER_RESOURCE,
	ORDER_HTTP_VERSION,
	ORDER_STATUS_CODE,
	ORDER_REQUEST_SIZE,
	ORDER_HOST,
	ORDER_USER_AGENT,
}

func (o order_t) String() string {
	switch o {
	case ORDER_IP:
		return ":ip"
	case ORDER_TIME:
		return ":time"
	case ORDER_METHOD:
		return ":method"
	case ORDER_RESOURCE:
		return ":resource"
	case ORDER_HTTP_VERSION:
		return ":http_version"
	case ORDER_STATUS_CODE:
		return ":status_code"
	case ORDER_REQUEST_SIZE:
		return ":request_size"
	case ORDER_HOST:
		return ":host"
	case ORDER_USER_AGENT:
		return ":user_agent"
	}

	return ":unknown"
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)

	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func (c Configs) writeToConfigFile(configFilePath string) error {
	file, err := os.OpenFile(configFilePath, os.O_CREATE|os.O_WRONLY, 0600)

	if err != nil {
		return err
	}

	file.WriteString("regex = ")
	file.WriteString(c.regex.String())
	file.WriteString("\n")
	file.WriteString("order = [")

	for i, item := range c.order {
		if i > 0 {
			file.WriteString(", ")
		}

		switch item {
		case ORDER_IP:
			file.WriteString(":ip")
		case ORDER_TIME:
			file.WriteString(":time")
		case ORDER_METHOD:
			file.WriteString(":method")
		case ORDER_RESOURCE:
			file.WriteString(":resource")
		case ORDER_HTTP_VERSION:
			file.WriteString(":http_version")
		case ORDER_STATUS_CODE:
			file.WriteString(":status_code")
		case ORDER_REQUEST_SIZE:
			file.WriteString(":request_size")
		case ORDER_HOST:
			file.WriteString(":host")
		case ORDER_USER_AGENT:
			file.WriteString(":user_agent")
		}
	}

	file.WriteString("]\n")

	file.WriteString("format = ")
	file.WriteString(c.format)
	file.WriteString("\n")

	return file.Close()
}

func getFileSize(filePath string) (int64, error) {
	stat, err := os.Stat(filePath)

	if err != nil {
		return 0, err
	}

	return stat.Size(), nil
}

func LoadConfigs() (*Configs, error) {
	configs := Configs{
		regex:  defaultLogRegex,
		order:  defaultOrder,
		format: defaultFormatting,
	}

	userHomeDir, err := os.UserHomeDir()

	if err != nil {
		return nil, err
	}

	configFilePath := path.Join(userHomeDir, configFileName)

	exists, err := pathExists(configFilePath)

	if err != nil {
		return nil, err
	}

	if !exists {
		configs.writeToConfigFile(configFilePath)

		return &configs, nil
	}

	fileSize, err := getFileSize(configFilePath)

	if err != nil {
		return nil, err
	}

	if fileSize > MAX_CONFIG_FILE_SIZE {
		return nil, fmt.Errorf("config file is too big. it's impossible this file to exceed %d bytes", MAX_CONFIG_FILE_SIZE)
	}

	data, err := os.ReadFile(configFilePath)

	if err != nil {
		return nil, err
	}

	for number, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)

		if len(line) == 0 {
			continue
		}

		equalIndex := strings.Index(line, "=")

		if equalIndex == -1 {
			return nil, fmt.Errorf("%s:%d error: invalid config line", configFilePath, number+1)
		}

		key := strings.TrimSpace(line[:equalIndex])

		if len(key) == 0 {
			return nil, fmt.Errorf("%s:%d error: missing key", configFilePath, number+1)
		}

		value := strings.TrimSpace(line[equalIndex+1:])

		switch key {
		case "regex":
			if len(value) == 0 {
				return nil, fmt.Errorf("%s:%d error: missing value for regex", configFilePath, number+1)
			}

			configs.regex = regexp.MustCompile(value)
		case "order":
			order, err := parseOrderArray(configFilePath, number+1, value)

			if err != nil {
				return nil, err
			}

			configs.order = order
		case "format":
			configs.format = value
		default:
			return nil, fmt.Errorf("%s:%d error: invalid config key \"%s\"", configFilePath, number+1, key)
		}
	}

	return &configs, nil
}

func parseOrderArray(configFilePath string, lineNumber int, content string) ([]order_t, error) {
	order := make([]order_t, 0, ORDER_COUNT)

	if len(content) == 0 {
		return []order_t{}, fmt.Errorf("%s:%d error: please specify a value to the order", configFilePath, lineNumber)
	}

	cursor := 0
	bot := 0

	if content[cursor] != '[' {
		return []order_t{}, fmt.Errorf("%s:%d error: expected \"[\" but received \"%c\"", configFilePath, lineNumber, content[cursor])
	}

	cursor++
	bot = cursor
	hasComma := false

	for cursor < len(content) {
		for cursor < len(content) && content[cursor] == ' ' || content[cursor] == '\n' {
			cursor++
		}

		bot = cursor

		if content[cursor] == ']' {
			break
		}

		if content[cursor] == ':' {
			cursor++

			for cursor < len(content) && ((content[cursor] >= 'a' && content[cursor] <= 'z') || content[cursor] == '_') {
				cursor++
			}

			key := content[bot:cursor]

			switch key {
			case ":ip":
				order = append(order, ORDER_IP)
			case ":time":
				order = append(order, ORDER_TIME)
			case ":method":
				order = append(order, ORDER_METHOD)
			case ":resource":
				order = append(order, ORDER_RESOURCE)
			case ":http_version":
				order = append(order, ORDER_HTTP_VERSION)
			case ":status_code":
				order = append(order, ORDER_STATUS_CODE)
			case ":request_size":
				order = append(order, ORDER_REQUEST_SIZE)
			case ":host":
				order = append(order, ORDER_HOST)
			case ":user_agent":
				order = append(order, ORDER_USER_AGENT)
			default:
				return []order_t{}, fmt.Errorf("%s:%d error: invalid order item \"%s\"", configFilePath, lineNumber, key)
			}

			if cursor < len(content) && content[cursor] == ',' {
				hasComma = true
				cursor++
			} else {
				hasComma = false
			}
		} else {
			return []order_t{}, fmt.Errorf("%s:%d error: unexpected \"%c\"", configFilePath, lineNumber, content[cursor])
		}
	}

	if content[cursor] != ']' {
		return []order_t{}, fmt.Errorf("%s:%d error: expected \"]\" but received \"%c\"", configFilePath, lineNumber, content[cursor])
	}

	if hasComma {
		return []order_t{}, fmt.Errorf("%s:%d error: unexpected \",\" at the end of the array", configFilePath, lineNumber)
	}

	if len(order) < int(ORDER_COUNT) {
		missing := make([]string, 0, ORDER_COUNT)

		for _, a := range defaultOrder {
			if !slices.Contains(order, a) {
				missing = append(missing, a.String())
			}
		}

		missingString := strings.Join(missing, ", ")

		return []order_t{}, fmt.Errorf("%s:%d error: missing order values: %s", configFilePath, lineNumber, missingString)
	}

	empty := struct{}{}
	uniq := make(map[order_t]struct{})

	for _, n := range order {
		if _, ok := uniq[n]; ok {
			return []order_t{}, fmt.Errorf("%s:%d error: duplicated \"%s\" inside enum", configFilePath, lineNumber, n.String())
		}

		uniq[n] = empty
	}

	return order, nil
}
