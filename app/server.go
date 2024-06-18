package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
)

var FILE_ROOT = "/tmp"

func main() {
	fmt.Println("Logs from your program will appear here!")

	if len(os.Args) < 3 {
		fmt.Println("No directory for files. Using default directory /tmp")
	} else {
		FILE_ROOT = os.Args[2]
	}

	l, err := net.Listen("tcp", "0.0.0.0:4221")
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
		go handleConnection(conn)
	}
}

func handleConnection(rwc io.ReadWriteCloser) {
	defer rwc.Close()
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
	handler := getHandler(request.Path)
	handler(request, response)

}

func getHandler(path string) Handler {
	if path == "/" {
		return func(r *Request, w ResponseWriter) {
			w.Write([]byte(""))
		}
	} else if path == "/user-agent" {
		return func(r *Request, w ResponseWriter) {
			userAgent, ok := r.Headers["user-agent"]
			if !ok {
				w.WriteHeader(400)
				w.Write([]byte("User-Agent header is required"))
				return
			}
			w.Write([]byte(userAgent))
		}
	} else if strings.HasPrefix(path, "/echo/") {
		str, _ := strings.CutPrefix(path, "/echo/")
		fmt.Println("echoing back: ", str)

		return func(r *Request, w ResponseWriter) {
			w.Write([]byte(str))
		}
	} else if strings.HasPrefix(path, "/files/") {
		path, _ := strings.CutPrefix(path, "/files/")
		fmt.Println("Asking for file ", path)
		return func(r *Request, w ResponseWriter) {
			filePath := filepath.Join(FILE_ROOT, path)
			fileBody, err := os.ReadFile(filePath)
			if err != nil {
				w.WriteHeader(404)
				w.Write(nil)
				return
			}
			w.Headers()["Content-Type"] = "application/octet-stream"
			w.Write(fileBody)
		}
	} else {
		return func(r *Request, w ResponseWriter) {
			w.WriteHeader(404)
		}
	}
}
