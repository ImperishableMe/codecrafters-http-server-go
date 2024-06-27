package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	serverOptions := ServerOptions{}
	if len(os.Args) >= 3 {
		serverOptions.FileServerRoot = os.Args[2]
	}
	s := NewServer(serverOptions)

	// Register the handlers
	s.Register("GET /", func(r *Request, w ResponseWriter) {
		w.Write([]byte(""))
	})
	s.Register("GET /user-agent", func(r *Request, w ResponseWriter) {
		userAgent, ok := r.Headers["user-agent"]
		if !ok {
			w.WriteHeader(400)
			w.Write([]byte("User-Agent header is required"))
			return
		}
		w.Write([]byte(userAgent))
	})
	s.Register("GET /echo/{str}", func(r *Request, w ResponseWriter) {
		str, _ := strings.CutPrefix(r.Path, "/echo/")
		fmt.Println("echoing back: ", str)
		w.Write([]byte(str))
	})
	s.Register("GET /files/{path}", func(r *Request, w ResponseWriter) {
		path, _ := strings.CutPrefix(r.Path, "/files/")
		fmt.Println("Asking for file ", path)
		filePath := filepath.Join(s.fileServerRoot, path)
		fileBody, err := os.ReadFile(filePath)
		if err != nil {
			w.WriteHeader(404)
			w.Write(nil)
			return
		}
		w.Headers()["Content-Type"] = "application/octet-stream"
		w.Write(fileBody)
	})
	s.Register("POST /files/{path}", func(r *Request, w ResponseWriter) {
		fmt.Println(r.Headers)
		path, _ := strings.CutPrefix(r.Path, "/files/")
		filePath := filepath.Join(s.fileServerRoot, path)
		file, err := os.Create(filePath)
		if err != nil {
			w.WriteHeader(400)
			return
		}
		defer file.Close()

		lenStr, ok := r.GetHeader("Content-Length")
		if !ok {
			w.WriteHeader(400)
			w.Write([]byte("missing Content-Length header."))
			return
		}
		fileSize, _ := strconv.Atoi(lenStr)
		fmt.Println("Reading total file size of ", fileSize)

		written, err := io.CopyN(file, r.Body, int64(fileSize))
		fmt.Println("Read ", written, " bytes")
		if err != nil || written != int64(fileSize) {
			w.WriteHeader(400)
			w.Write([]byte(fmt.Sprintf("couldn't write the whole content. Wrote %d, expected %d", written, fileSize)))
			return
		}
		w.WriteHeader(201)
	})

	s.ListenAndServe()
}

type Handler struct {
	pat     pattern
	handler HandlerFunc
}

func isWild(s string) bool {
	return len(s) >= 2 && string(s[0]) == "{" && string(s[len(s)-1]) == "}"
}

func (h Handler) matches(pat pattern) bool {
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
	fileServerRoot string
	handlers       []Handler
	port           int
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
		fileServerRoot: options.FileServerRoot,
		handlers:       make([]Handler, 0),
		port:           options.Port,
	}
}

func (s *Server) Register(pat string, h HandlerFunc) error {
	p, err := fromString(pat)
	fmt.Println("Added pattern ", len(p.segments), p.method, p.segments)
	if err != nil {
		return err
	}
	s.handlers = append(s.handlers, Handler{pat: p, handler: h})
	return nil
}

func (s *Server) ListenAndServe() {
	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", s.port))
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
	handler(request, response)
}

func (s *Server) findHandler(path string) HandlerFunc {
	pat, err := fromString(path)
	if err != nil {
		return NotFoundHandler
	}
	fmt.Println("trying to match pattern..", pat)
	for _, h := range s.handlers {
		if h.matches(pat) {
			return h.handler
		}
	}
	return NotFoundHandler
}

var NotFoundHandler = func(r *Request, w ResponseWriter) {
	w.WriteHeader(404)
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
