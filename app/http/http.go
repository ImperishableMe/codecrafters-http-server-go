package http

import (
	"fmt"
	"io"
	"strconv"
)

type Headers map[string]string
type ResponseWriter interface {
	WriteHeader(status int)
	Write([]byte) (int, error)
	Headers() Headers
}
type HandlerFunc func(*Request, ResponseWriter)

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
