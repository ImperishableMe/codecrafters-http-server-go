package http

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type Request struct {
	Path    string
	Method  string
	Headers map[string][]string
	Body    io.Reader
}

func NewRequest() *Request {
	return &Request{
		Headers: make(map[string][]string),
	}
}

func (r *Request) GetHeader(key string) ([]string, bool) {
	val, ok := r.Headers[strings.ToLower(key)]
	return val, ok
}

func parseRequest(r io.Reader) (*Request, error) {
	bufR := bufio.NewReader(r)
	request := NewRequest()

	// parse request line
	requestLine, err := bufR.ReadString('\n')
	if err != nil {
		return request, err
	}
	splits := strings.Split(strings.Trim(requestLine, "\r\n"), " ")
	if len(splits) < 2 {
		return request, fmt.Errorf("invalid request")
	}
	request.Method = splits[0]
	request.Path = splits[1]

	// parse headers
	for {
		line, err := bufR.ReadString('\n')
		if err != nil {
			return request, err
		}
		line = strings.Trim(line, "\r\n")
		if line == "" {
			break
		}
		splits := strings.Split(line, ": ")
		header := strings.ToLower(splits[0])
		value := strings.Split(splits[1], ", ")
		fmt.Println("Adding header ", header, ": ", value)
		request.Headers[header] = value
	}

	request.Body = bufR
	return request, nil
}
