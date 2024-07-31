package http

import (
	"compress/gzip"
	"fmt"
	"io"
	"strconv"
)

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

type GzipResponseWriter struct {
	w           ResponseWriter
	gw          io.WriteCloser
	wroteHeader bool
}

func (g GzipResponseWriter) WriteHeader(status int) {
	g.w.WriteHeader(status)
	g.wroteHeader = true
}

func (g GzipResponseWriter) Write(p []byte) (int, error) {
	if !g.wroteHeader {
		g.WriteHeader(200)
	}
	defer func(gw io.WriteCloser) {
		err := gw.Close()
		if err != nil {
			fmt.Println("Error closing gzip writer:", err)
		}
	}(g.gw)
	return g.gw.Write(p)
}

func (g GzipResponseWriter) Headers() Headers {
	return g.w.Headers()
}

func NewGzipResponseWriter(w ResponseWriter) GzipResponseWriter {
	gw := gzip.NewWriter(w)
	return GzipResponseWriter{w: w, gw: gw}
}
