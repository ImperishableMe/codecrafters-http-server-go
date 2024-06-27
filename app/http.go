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
	Body    io.Reader
}

func NewRequest() *Request {
	return &Request{
		Headers: make(map[string]string),
	}
}

func (r *Request) GetHeader(key string) (string, bool) {
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
		value := splits[1]
		fmt.Println("Adding header ", header, ": ", value)
		request.Headers[header] = value
	}

	request.Body = bufR
	return request, nil
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
	statusDescription := map[int]string{
		200: "OK",
		201: "Created",
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
