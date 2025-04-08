package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var wg sync.WaitGroup

var kongLogRegex = regexp.MustCompile(`^\d{1,3}.\d{1,3}.\d{1,3}.\d{1,3} - - \[\d{1,2}\/\w+\/\d{4}:\d{2}:\d{2}:\d{2} \+\d{4}\] "\w+ .*? HTTP\/\d.\d" \d+ \d+ ".*?" ".*?"$`)
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

	if !kongLogRegex.MatchString(line) {
		return log, errors.New(line)
	}

	loc := ipRegex.FindStringIndex(line)

	log.ip = line[loc[0]:loc[1]]
	line = line[loc[1]+5:]

	loc = timeRegex.FindStringIndex(line)

	log.time = line[loc[0]:loc[1]]

	line = line[loc[1]+2:]

	loc = stringRegex.FindStringIndex(line)

	split := strings.Split(line[loc[0]+1:loc[1]-1], " ")

	if statusCode, err := stringMethodToType(split[0]); err == nil {
		log.method = statusCode
	} else {
		return log, err
	}

	log.resource = split[1]
	log.version = split[2]

	line = line[loc[1]+1:]

	loc = statusCodeRegex.FindStringIndex(line)

	if n, err := strconv.ParseInt(line[loc[0]:loc[1]], 10, 32); err == nil {
		log.statusCode = int(n)
	} else {
		return log, err
	}

	line = line[loc[1]+1:]

	loc = sizeRegex.FindStringIndex(line)

	if n, err := strconv.ParseInt(line[loc[0]:loc[1]], 10, 32); err == nil {
		log.size = int(n)
	} else {
		return log, err
	}

	line = line[loc[1]+1:]

	loc = stringRegex.FindStringIndex(line)

	log.host = line[loc[0]+1 : loc[1]-1]

	line = line[loc[1]+1:]

	loc = stringRegex.FindStringIndex(line)

	log.userAgent = line[loc[0]+1 : loc[1]-1]

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

func parseFormatString(format string, log log_t) {
	for i := 0; i < len(format); i += 1 {
		if format[i] == '%' {
			j := i + 1

			for ; j < len(format); j += 1 {
				if !isAlpha(format[j]) {
					break
				}
			}

			label := format[i+1 : j]

			switch label {
			case "time":
				fmt.Printf(log.time)
			case "ip":
				fmt.Printf(log.ip)
			case "method":
				fmt.Printf(methodDisplay(log.method))
			case "resource":
				fmt.Printf(log.resource)
			case "version":
				fmt.Printf(log.version)
			case "status":
				fmt.Printf("%d", log.statusCode)
			case "size":
				fmt.Printf("%d", log.size)
			case "host":
				fmt.Printf(log.host)
			case "agent":
				fmt.Printf(log.userAgent)
			default:
				fmt.Fprintf(os.Stderr, "error: invalid formatting string at label %%%s\n", label)
				os.Exit(1)
			}

			i = j - 1
		} else if format[i] == '\\' {
			if i+1 >= len(format) {
				fmt.Fprintf(os.Stderr, "error: untermiated back slashed char\n")
				os.Exit(1)
			}

			switch format[i+1] {
			case 't':
				fmt.Printf("\t")
			default:
				fmt.Fprintf(os.Stderr, "error: invalid special char \\%c\n", format[i+1])
				os.Exit(1)
			}

			i += 1
		} else if format[i] == ' ' {
			fmt.Printf(" ")
		} else {
			fmt.Fprintf(os.Stderr, "error: invalid formatting string at char \"%c\"\n", format[i])
			os.Exit(1)
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
			parseFormatString(flags.format, log)
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
	format            string
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
			fmt.Fprintf(os.Stderr, "error: invalid ip format")
			os.Exit(1)
		}
	}

	logs := make(chan []byte, 0)

	user_flags := flags_t{
		format:            *format,
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
