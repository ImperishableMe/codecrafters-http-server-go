package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type Request struct {
	Path    string
	Method  string
	Headers map[string]string
	Body    []byte
}

func (r Request) string() string {
	return fmt.Sprintf("Request{Path: %s, Method: %s, Headers: %v, Body: %s}", r.Path, r.Method, r.Headers, r.Body)
}

func newRequest() *Request {
	return &Request{
		Headers: make(map[string]string),
		Body:    make([]byte, 0),
	}
}

func parseRequest(c net.Conn) (*Request, error) {
	scanner := bufio.NewScanner(c)
	request := newRequest()

	// parse request line
	if scanner.Scan() {
		splits := strings.Split(scanner.Text(), " ")
		if len(splits) < 2 {
			return request, fmt.Errorf("invalid request")
		}
		request.Method = splits[0]
		request.Path = splits[1]
	}
	// CLRF
	ok := scanner.Scan()
	if !ok {
		return request, scanner.Err()
	}

	// parse headers
	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			break
		}

		splits := strings.Split(line, ": ")
		header := strings.ToLower(splits[0])
		value := splits[1]
		fmt.Println("Adding header: ", header, " with value: ", value)
		request.Headers[header] = value
	}
	// CLRF
	// ok = scanner.Scan()
	// if !ok {
	// 	return request, nil
	// }

	return request, scanner.Err()
}
