package http

var gzipper = func(r *Request, w ResponseWriter, f Handler) {
	encodings, _ := r.GetHeader("accept-encoding")
	gZipFound := false
	for _, enc := range encodings {
		if enc == "gzip" {
			gZipFound = true
		}
	}
	if !gZipFound {
		f.ServeHttp(r, w)
		return
	}
	w.Headers()["Content-Encoding"] = "gzip"
	f.ServeHttp(r, w)
}
