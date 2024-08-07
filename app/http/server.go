package http

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

type HandlerNode struct {
	pat     pattern
	handler Handler
}

func isWild(s string) bool {
	return len(s) >= 2 && string(s[0]) == "{" && string(s[len(s)-1]) == "}"
}

func (h HandlerNode) matches(pat pattern) bool {
	if h.pat.method != pat.method {
		return false
	}
	segs1, segs2 := h.pat.segments, pat.segments
	i, j := 0, 0
	for ; i < len(segs1) && j < len(segs2); i, j = i+1, j+1 {
		if segs1[i] == segs2[j] || isWild(segs1[i]) {
			continue
		}
		return false
	}
	return i == j
}

type Server struct {
	FileServerRoot string
	Handlers       []HandlerNode
	Port           int
}

func NewServer(options ServerOptions) Server {
	if options.FileServerRoot == "" {
		options.FileServerRoot = "/tmp/"
	}
	// I know, it's ugly! Anyway, life goes on!
	if options.Port == 0 {
		options.Port = 4221
	}
	return Server{
		FileServerRoot: options.FileServerRoot,
		Handlers:       make([]HandlerNode, 0),
		Port:           options.Port,
	}
}

func (s *Server) Register(pat string, h Handler) {
	p, err := fromString(pat)
	fmt.Println("Added pattern ", len(p.segments), p.method, p.segments)
	if err != nil {
		fmt.Println(err)
		return
	}
	s.Handlers = append(s.Handlers, HandlerNode{pat: p, handler: h})
}

func (s *Server) ListenAndServe() {
	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", s.Port))
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	fmt.Printf("Listening on: %v\n", l.Addr().String())

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		fmt.Println("Accepted connection from: ", conn.RemoteAddr().String())
		go s.Serve(conn)
	}
}

func (s *Server) Serve(rwc io.ReadWriteCloser) {
	defer func() {
		fmt.Println("Closing down the request")
		rwc.Close()
	}()
	request, err := parseRequest(rwc)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Println("Got request for path: ", request.Path)

	response := &Response{
		headers: make(Headers),
		writer:  rwc,
	}
	handler := s.findHandler(request.Method + " " + request.Path)
	gzipMiddleware(handler).ServeHttp(request, response)
}

func (s *Server) findHandler(path string) Handler {
	pat, err := fromString(path)
	if err != nil {
		return NotFoundHandler()
	}
	fmt.Println("trying to match pattern..", pat)
	for _, h := range s.Handlers {
		if h.matches(pat) {
			return h.handler
		}
	}
	return NotFoundHandler()
}

var NotFound = func(r *Request, w ResponseWriter) {
	w.WriteHeader(404)
}

func NotFoundHandler() Handler {
	return HandlerFunc(NotFound)
}

type ServerOptions struct {
	FileServerRoot string
	Port           int
}

type pattern struct {
	method   string
	segments []string
}

func fromString(pat string) (pattern, error) {
	splits := strings.Split(pat, " ")
	if len(splits) < 2 {
		return pattern{}, fmt.Errorf("pattern: %s does not have <VERB PATH> form", pat)
	}
	verb, path := splits[0], splits[1]
	return pattern{
		method:   verb,
		segments: strings.Split(path, "/"),
	}, nil
}
