package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
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

func NewRequest() *Request {
	return &Request{
		Headers: make(map[string]string),
		Body:    make([]byte, 0),
	}
}

func parseRequest(r io.Reader) (*Request, error) {
	scanner := bufio.NewScanner(r)
	request := NewRequest()

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

type Response struct {
	ContentType string
	Headers     map[string]string
	Body        []byte
	Status      int
}

func writeResponse(wc io.WriteCloser, response Response) {
	defer wc.Close()

	statusDescription := map[int]string{
		200: "OK",
		404: "Not Found",
	}
	status := response.Status
	body := response.Body
	if response.Headers == nil {
		response.Headers = make(map[string]string)
	}
	header := response.Headers

	fmt.Fprintf(wc, "HTTP/1.1 %d %s\r\n", status, statusDescription[status])
	header["content-length"] = strconv.Itoa(len(body)) // override the `content-length` header

	for k, v := range header {
		fmt.Fprintf(wc, "%s: %s\r\n", k, v)
	}
	fmt.Fprintf(wc, "\r\n")
	if body != nil {
		wc.Write(body)
	}
}
