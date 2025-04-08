package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"sync"

	"github.com/marcos-venicius/lfi/formatter"
)

var wg sync.WaitGroup

var kongLogRegex = regexp.MustCompile(`^(\d{1,3}.\d{1,3}.\d{1,3}.\d{1,3}) - - \[(\d{1,2}\/\w+\/\d{4}:\d{2}:\d{2}:\d{2} \+\d{4})\] "(\w+) (.*?) (HTTP\/\d.\d)" (\d+) (\d+) "(.*?)" "(.*?)"$`)
var ipRegex = regexp.MustCompile(`^\d{1,3}.\d{1,3}.\d{1,3}.\d{1,3}`)
var ipRegexFull = regexp.MustCompile(`^\d{1,3}.\d{1,3}.\d{1,3}.\d{1,3}$`)
var timeRegex = regexp.MustCompile(`\d{1,2}\/\w+\/\d{4}:\d{2}:\d{2}:\d{2} \+\d{4}`)
var stringRegex = regexp.MustCompile(`^".*?"`)
var statusCodeRegex = regexp.MustCompile(`^\d{3}`)
var sizeRegex = regexp.MustCompile(`^\d+`)

type method_t int

const (
	http_get     method_t = iota
	http_post    method_t = iota
	http_delete  method_t = iota
	http_patch   method_t = iota
	http_put     method_t = iota
	http_options method_t = iota
)

type log_t struct {
	ip         string
	time       string
	method     method_t
	host       string
	resource   string
	version    string
	statusCode int
	size       int
	userAgent  string
}

func methodDisplay(method method_t) string {
	switch method {
	case http_options:
		return "OPTIONS"
	case http_get:
		return "GET"
	case http_put:
		return "PUT"
	case http_post:
		return "POST"
	case http_delete:
		return "DELETE"
	case http_patch:
		return "PATCH"
	}

	return "unknown"
}

func stringMethodToType(method string) (method_t, error) {
	switch method {
	case "OPTIONS", "options":
		return http_options, nil
	case "GET", "get":
		return http_get, nil
	case "PUT", "put":
		return http_put, nil
	case "POST", "post":
		return http_post, nil
	case "DELETE", "delete":
		return http_delete, nil
	case "PATCH", "patch":
		return http_patch, nil
	}

	return 0, errors.New("invalid status code")
}

func isValidIP(ip string) bool {
	return ipRegexFull.MatchString(ip)
}

func parseKongLogLine(line string) (log_t, error) {
	log := log_t{}

	matches := kongLogRegex.FindStringSubmatch(line)

	if len(matches) != 10 {
		return log, errors.New(line)
	}

	log.ip = matches[1]
	log.time = matches[2]

	if method, err := stringMethodToType(matches[3]); err == nil {
		log.method = method
	} else {
		return log, err
	}

	log.resource = matches[4]
	log.version = matches[5]

	if n, err := strconv.ParseInt(matches[6], 10, 32); err == nil {
		log.statusCode = int(n)
	} else {
		return log, err
	}

	if n, err := strconv.ParseInt(matches[7], 10, 32); err == nil {
		log.size = int(n)
	} else {
		return log, err
	}

	log.host = matches[8]

	log.userAgent = matches[9]

	return log, nil
}

func filterLog(log log_t, filters filters_t) bool {
	if filters.statusCode.active {
		switch filters.statusCode.mode {
		case filter_mode_eq:
			if log.statusCode != filters.statusCode.value {
				return false
			}
		case filter_mode_ne:
			if log.statusCode == filters.statusCode.value {
				return false
			}
		}
	}

	if filters.method.active {
		switch filters.method.mode {
		case filter_mode_eq:
			if log.method != filters.method.value {
				return false
			}
		case filter_mode_ne:
			if log.method == filters.method.value {
				return false
			}
		}
	}

	if filters.ip.active {
		switch filters.ip.mode {
		case filter_mode_eq:
			if log.ip != filters.ip.value {
				return false
			}
		case filter_mode_ne:
			if log.ip == filters.ip.value {
				return false
			}
		}
	}

	return true
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func displayLogsBasedOnFormatting(tokens []string, log log_t) {
	for _, token := range tokens {
		if token[0] == '\'' {
			fmt.Print(token[1 : len(token)-1])
		} else if token[0] == ' ' {
			fmt.Print(token)
		} else {
			switch token {
			case "%time":
				fmt.Print(log.time)
			case "%ip":
				fmt.Print(log.ip)
			case "%method":
				fmt.Print(methodDisplay(log.method))
			case "%resource":
				fmt.Print(log.resource)
			case "%version":
				fmt.Print(log.version)
			case "%status":
				fmt.Print(log.statusCode)
			case "%size":
				fmt.Print(log.size)
			case "%host":
				fmt.Print(log.host)
			case "%agent":
				fmt.Print(log.userAgent)
			default:
				fmt.Print(token)
			}
		}
	}

	fmt.Println()
}

func worker(logs chan []byte, flags flags_t, filters filters_t) {
	defer wg.Done()

	for {
		line := <-logs

		if len(line) == 0 {
			break
		}

		lineString := string(line)

		log, err := parseKongLogLine(lineString)

		if err != nil {
			if flags.displayErrorLines {
				fmt.Println(err)
			}
		} else if filterLog(log, filters) {
			displayLogsBasedOnFormatting(flags.formatTokens, log)
		}
	}
}

type filter_mode_t int

const (
	filter_mode_eq filter_mode_t = iota
	filter_mode_ne filter_mode_t = iota
)

type filter_t[T int | method_t | string] struct {
	active bool
	value  T
	mode   filter_mode_t
}

type filters_t struct {
	statusCode filter_t[int]
	method     filter_t[method_t]
	ip         filter_t[string]
}

type flags_t struct {
	formatTokens      []string
	displayErrorLines bool
}

func main() {
	filters := filters_t{}

	status := flag.Int("fs", -1, "filter by a specific status code. when -1 the filter is not used")
	statusNe := flag.Int("nefs", -1, "filter by logs where the status code is not equal to the provided one. when -1 the filter is not used")
	method := flag.String("fm", "", "filter by a specific method")
	methodNe := flag.String("nefm", "", "filter by logs where the method is not equal to the provided one")
	ip := flag.String("fi", "", "filter by a specific ip")
	displayErrors := flag.Bool("de", false, "disable error lines output")

	format := flag.String("f", "%time %ip %method %resource %version %status %size %host %agent", "format the log in a specific way")

	flag.Parse()

	if *status != -1 {
		filters.statusCode.active = true
		filters.statusCode.value = *status
		filters.statusCode.mode = filter_mode_eq
	}

	if *statusNe != -1 {
		filters.statusCode.active = true
		filters.statusCode.value = *statusNe
		filters.statusCode.mode = filter_mode_ne
	}

	if *method != "" {
		if parsedMethod, err := stringMethodToType(*method); err == nil {
			filters.method.active = true
			filters.method.value = parsedMethod
			filters.method.mode = filter_mode_eq
		} else {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			os.Exit(1)
		}
	}

	if *methodNe != "" {
		if parsedMethod, err := stringMethodToType(*methodNe); err == nil {
			filters.method.active = true
			filters.method.value = parsedMethod
			filters.method.mode = filter_mode_ne
		} else {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			os.Exit(1)
		}
	}

	if *ip != "" {
		if isValidIP(*ip) {
			filters.ip.active = true
			filters.ip.value = *ip
			filters.ip.mode = filter_mode_eq
		} else {
			fmt.Fprintf(os.Stderr, "error: invalid ip format\n")
			os.Exit(1)
		}
	}

	logFormatter := formatter.CreateFormatter([]string{"time", "ip", "method", "resource", "version", "status", "size", "host", "agent"})

	tokens, err := logFormatter.ParseFormatString(*format)

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(1)
	}

	logs := make(chan []byte, 0)

	user_flags := flags_t{
		formatTokens:      tokens,
		displayErrorLines: !*displayErrors,
	}

	wg.Add(1)
	go worker(logs, user_flags, filters)

	bytes := make([]byte, 256)

	line := make([]byte, 0, 1024)
	lineSize := 0

	for {
		n, err := os.Stdin.Read(bytes)

		if err == io.EOF {
			break
		}

		for i := range n {
			if bytes[i] == '\n' {
				logs <- line[:lineSize]

				line = make([]byte, 0, 1024)
				lineSize = 0
			} else {
				line = append(line, bytes[i])
				lineSize += 1
			}
		}

		bytes = make([]byte, 256)
	}

	close(logs)
	wg.Wait()
}
