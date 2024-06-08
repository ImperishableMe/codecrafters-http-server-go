package main

import "fmt"

type Request struct {
	Path    string
	Method  string
	Headers map[string]string
	Body    []byte
}

func (r Request) string() string {
	return fmt.Sprintf("Request{Path: %s, Method: %s, Headers: %v, Body: %s}", r.Path, r.Method, r.Headers, r.Body)
}
