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
	"github.com/marcos-venicius/lfi/quang"
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

func worker(logs chan []byte, q quang.Quang, flags flags_t) {
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
		} else {
			q.AddStrVar("time", log.time)
			q.AddStrVar("ip", log.ip)
      q.AddAtomVar("method", int(log.method))
			q.AddStrVar("resource", log.resource)
			q.AddStrVar("version", log.version)
			q.AddIntVar("status", log.statusCode)
			q.AddIntVar("size", log.size)
			q.AddStrVar("host", log.host)
			q.AddStrVar("agent", log.userAgent)

			if q.Eval() {
				displayLogsBasedOnFormatting(flags.formatTokens, log)
			}
		}
	}
}

type flags_t struct {
	formatTokens      []string
	displayErrorLines bool
}

// TODO: create a built-in language to accept "and", "or" and queries
// today, we only accept and operations to all filters
func main() {
	query := flag.String("q", "", "query")

	displayErrors := flag.Bool("de", false, "disable error lines output")

	format := flag.String("f", "%time %ip %method %resource %version %status %size %host %agent", "format the log in a specific way")

	flag.Parse()

	logFormatter := formatter.CreateFormatter([]string{"time", "ip", "method", "resource", "version", "status", "size", "host", "agent"})

	tokens, err := logFormatter.ParseFormatString(*format)

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(1)
	}

	q, err := quang.CreateQuang(*query)

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(1)
	}

  q.AddAtom("get", int(http_get))
  q.AddAtom("GET", int(http_get))
  q.AddAtom("post", int(http_post))
  q.AddAtom("POST", int(http_post))
  q.AddAtom("delete", int(http_delete))
  q.AddAtom("DELETE", int(http_delete))
  q.AddAtom("patch", int(http_patch))
  q.AddAtom("PATCH", int(http_patch))
  q.AddAtom("put", int(http_put))
  q.AddAtom("PUT", int(http_put))
  q.AddAtom("options", int(http_options))
  q.AddAtom("OPTIONS", int(http_options))

	logs := make(chan []byte, 0)

	user_flags := flags_t{
		formatTokens:      tokens,
		displayErrorLines: !*displayErrors,
	}

	wg.Add(1)
	go worker(logs, q, user_flags)

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
