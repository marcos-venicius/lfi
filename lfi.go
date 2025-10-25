package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/marcos-venicius/lfi/formatter"
	"github.com/marcos-venicius/quang"
)

type lfi_t struct {
	formatTokens []string
	verbose      bool

	q *quang.Quang
}

type log_t struct {
	ip         string
	time       string
	method     quang.AtomType
	host       string
	resource   string
	version    string
	statusCode quang.IntegerType
	size       quang.IntegerType
	userAgent  string
}

var wg sync.WaitGroup

var kongLogRegex = regexp.MustCompile(`^(\d{1,3}.\d{1,3}.\d{1,3}.\d{1,3}) .* .* \[(\d{1,2}\/\w+\/\d{4}:\d{2}:\d{2}:\d{2} \+\d{4})\] "(\w+) (.*?) (HTTP\/\d.\d)" (\d+|-) (\d+|-) "(.*?)" "(.*?)"$`)
var ipRegex = regexp.MustCompile(`^\d{1,3}.\d{1,3}.\d{1,3}.\d{1,3}`)
var ipRegexFull = regexp.MustCompile(`^\d{1,3}.\d{1,3}.\d{1,3}.\d{1,3}$`)
var timeRegex = regexp.MustCompile(`\d{1,2}\/\w+\/\d{4}:\d{2}:\d{2}:\d{2} \+\d{4}`)
var stringRegex = regexp.MustCompile(`^".*?"`)
var statusCodeRegex = regexp.MustCompile(`^\d{3}`)
var sizeRegex = regexp.MustCompile(`^\d+`)
var logsTimeout = 0
var lastUpdateTime = time.Now().UnixMilli()

const (
	http_get_atom quang.AtomType = iota
	http_post_atom
	http_delete_atom
	http_patch_atom
	http_put_atom
	http_options_atom
	http_head_atom
)

var atoms = map[string]quang.AtomType{
	":get":     http_get_atom,
	":post":    http_post_atom,
	":delete":  http_delete_atom,
	":patch":   http_patch_atom,
	":put":     http_put_atom,
	":options": http_options_atom,
	":head":    http_head_atom,
}

func methodDisplay(method quang.AtomType) string {
	switch method {
	case http_options_atom:
		return "OPTIONS"
	case http_get_atom:
		return "GET"
	case http_put_atom:
		return "PUT"
	case http_post_atom:
		return "POST"
	case http_delete_atom:
		return "DELETE"
	case http_patch_atom:
		return "PATCH"
	case http_head_atom:
		return "HEAD"
	}

	return "unknown"
}

func stringMethodToType(method string) (quang.AtomType, error) {
	switch method {
	case "OPTIONS", "options":
		return http_options_atom, nil
	case "GET", "get":
		return http_get_atom, nil
	case "PUT", "put":
		return http_put_atom, nil
	case "POST", "post":
		return http_post_atom, nil
	case "DELETE", "delete":
		return http_delete_atom, nil
	case "PATCH", "patch":
		return http_patch_atom, nil
	case "HEAD", "head":
		return http_head_atom, nil
	}

	return 0, errors.New("error: invalid method")
}

func parseKongLogLine(line string, breakParamsOut bool, configs *Configs) (log_t, error) {
	log := log_t{}

	positions := make(map[order_t]int)

	for p, key := range configs.order {
		positions[key] = p + 1
	}

	matches := configs.regex.FindStringSubmatch(line)

	if len(matches) != 10 {
		return log, errors.New(line)
	}

	log.ip = matches[positions[ORDER_IP]]
	log.time = matches[positions[ORDER_TIME]]

	if method, err := stringMethodToType(matches[positions[ORDER_METHOD]]); err == nil {
		log.method = method
	} else {
		return log, err
	}

	if breakParamsOut {
		parsed, err := url.Parse(matches[positions[ORDER_RESOURCE]])

		if err != nil {
			return log, err
		}

		log.resource = parsed.Path
	} else {
		log.resource = matches[positions[ORDER_RESOURCE]]
	}
	log.version = matches[positions[ORDER_HTTP_VERSION]]

	if n, err := strconv.ParseInt(matches[positions[ORDER_STATUS_CODE]], 10, 32); err == nil {
		log.statusCode = quang.IntegerType(n)
	} else {
		log.statusCode = 0
	}

	if n, err := strconv.ParseInt(matches[positions[ORDER_REQUEST_SIZE]], 10, 32); err == nil {
		log.size = quang.IntegerType(n)
	} else {
		log.size = 0
	}

	log.host = matches[positions[ORDER_HOST]]

	log.userAgent = matches[positions[ORDER_USER_AGENT]]

	return log, nil
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

func (l lfi_t) worker(logs chan []byte, breakParamsOut bool, configs *Configs) {
	defer wg.Done()

	for {
		line := <-logs

		if len(line) == 0 {
			break
		}

		lineString := string(line)

		log, err := parseKongLogLine(lineString, breakParamsOut, configs)

		if err != nil {
			if l.verbose {
				fmt.Println(err)
			}
		} else {
			l.q.AddStringVar("time", log.time).
				AddStringVar("ip", log.ip).
				AddAtomVar("method", log.method).
				AddStringVar("resource", log.resource).
				AddStringVar("version", log.version).
				AddIntegerVar("status", log.statusCode).
				AddIntegerVar("size", log.size).
				AddStringVar("host", log.host).
				AddStringVar("agent", log.userAgent)

			show, err := l.q.Eval()

			if err != nil {
				fmt.Println(err.Error())
			}

			if show {
				if logsTimeout != 0 {
					now := time.Now().UnixMilli()
					diff := int(now - lastUpdateTime)

					if diff < logsTimeout {
						remaining := logsTimeout - diff

						time.Sleep(time.Duration(remaining) * time.Millisecond)

						lastUpdateTime = time.Now().UnixMilli()
					}
				}

				displayLogsBasedOnFormatting(l.formatTokens, log)
			}
		}
	}
}

func isFlagParsed(name string) bool {
	found := false

	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})

	return found
}

func main() {
	verbose := flag.Bool("v", false, "when verbose mode is activated all errors will be shown")
	format := flag.String("f", defaultFormatting, "format the log in a specific way")
	timeout := flag.Int("t", 0, "timeout between logs. it's usefull when yours logs are crazingly fast. specify it in milliseconds")
	query := flag.String("q", "", "provide any valid filter using quang syntax https://github.com/marcos-venicius/quang.\navailable variables: time, ip, method, resource, version, status, size, host, agent.\navailable method atoms :get, :post, :delete, :patch, :put, :options.")
	breakParamsOut := flag.Bool("s", false, "strip out params from resource. everything like 'url<?param=value>' is going to be removed")

	flag.Parse()

	logsTimeout = *timeout

	configs, err := LoadConfigs()

	if err != nil {
		fmt.Println(err)

		os.Exit(1)
	}

	logFormatter := formatter.CreateFormatter([]string{"time", "ip", "method", "resource", "version", "status", "size", "host", "agent"})

	var formatting string = configs.format

	if isFlagParsed("f") {
		formatting = *format
	}

	tokens, err := logFormatter.ParseFormatString(formatting)

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(1)
	}

	q, err := quang.Init(*query)

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(1)
	}

	q.SetupAtoms(atoms)

	logs := make(chan []byte, 0)

	lfi := lfi_t{
		formatTokens: tokens,
		verbose:      *verbose,
		q:            q,
	}

	wg.Add(1)
	go lfi.worker(logs, *breakParamsOut, configs)

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
