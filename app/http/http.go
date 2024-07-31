package http

type Handler interface {
	ServeHttp(*Request, ResponseWriter)
}

type Headers map[string]string

type ResponseWriter interface {
	WriteHeader(status int)
	Write([]byte) (int, error)
	Headers() Headers
}

type HandlerFunc func(*Request, ResponseWriter)

func (h HandlerFunc) ServeHttp(request *Request, writer ResponseWriter) {
	h(request, writer)
}

func gzipMiddleware(h Handler) Handler {
	return HandlerFunc(
		func(r *Request, w ResponseWriter) {
			encodings, _ := r.GetHeader("accept-encoding")
			gZipFound := false
			for _, enc := range encodings {
				if enc == "gzip" {
					gZipFound = true
				}
			}
			if !gZipFound {
				h.ServeHttp(r, w)
				return
			}
			w.Headers()["Content-Encoding"] = "gzip"
			gzipResponseWriter := NewGzipResponseWriter(w)
			h.ServeHttp(r, gzipResponseWriter)
		})
}
