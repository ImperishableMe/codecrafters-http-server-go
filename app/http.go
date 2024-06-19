package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Headers map[string]string
type ResponseWriter interface {
	WriteHeader(status int)
	Write([]byte) (int, error)
	Headers() Headers
}
type HandlerFunc func(*Request, ResponseWriter)

type Request struct {
	Path    string
	Method  string
	Headers map[string]string
	Body    []byte
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
	headers     Headers
	writer      io.Writer
	wroteHeader bool
}

func (r *Response) Headers() Headers {
	return r.headers
}

func (r *Response) WriteHeader(status int) {
	if status != 200 && status != 404 {
		status = 500 // fallback status for now. Not needed for the tasks.
	}
	statusDescription := map[int]string{
		200: "OK",
		400: "Bad Request",
		404: "Not Found",
		500: "Unknown",
	}
	fmt.Fprintf(r.writer, "HTTP/1.1 %d %s\r\n", status, statusDescription[status])

	headers := r.headers
	for k, v := range headers {
		fmt.Fprintf(r.writer, "%s: %s\r\n", k, v)
	}
	fmt.Fprintf(r.writer, "\r\n")
	r.wroteHeader = true
}

func (r *Response) Write(p []byte) (int, error) {
	if !r.wroteHeader {
		r.headers["Content-Length"] = strconv.Itoa(len(p))
		_, exist := r.headers["Content-Type"]
		if !exist {
			r.headers["Content-Type"] = "text/plain"
		}
		r.WriteHeader(200)
	}
	return r.writer.Write(p)
}
