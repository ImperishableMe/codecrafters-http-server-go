package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
)

var FILE_ROOT string

func main() {
	fmt.Println("Logs from your program will appear here!")

	fmt.Println("Command line directory", os.Args[2])
	if len(os.Args) < 3 {
		fmt.Println("No directory for files")
		os.Exit(1)
	}
	FILE_ROOT = os.Args[2]

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
	request, err := parseRequest(rwc)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("Got request for path: ", request.Path)
	path := request.Path

	if path == "/" {
		writeResponse(rwc, Response{
			Status: 200,
			Body:   []byte(""),
		})
	} else if path == "/user-agent" {
		userAgent, ok := request.Headers["user-agent"]
		if !ok {
			writeResponse(rwc, Response{
				Status: 400,
				Body:   []byte("User-Agent header is required"),
			})
			return
		}
		writeResponse(rwc, Response{
			Status:  200,
			Body:    []byte(userAgent),
			Headers: map[string]string{"content-type": "text/plain"},
		})
	} else if strings.HasPrefix(path, "/echo/") {
		str, _ := strings.CutPrefix(path, "/echo/")
		fmt.Println("echoing back: ", str)
		writeResponse(rwc, Response{
			Status:  200,
			Body:    []byte(str),
			Headers: map[string]string{"content-type": "text/plain"},
		})
	} else if strings.HasPrefix(path, "/files/") {
		path, _ := strings.CutPrefix(path, "/files/")
		fmt.Println("Asking for file ", path)
		filePath := filepath.Join(FILE_ROOT, path)
		fileBody, err := os.ReadFile(filePath)
		if err != nil {
			writeResponse(rwc, Response{
				Status: 404,
			})
			return
		}
		writeResponse(rwc, Response{
			Status: 200,
			Body: fileBody,
			Headers: map[string]string{
				"content-type": "application/octet-stream",
			},
		})
	} else {
		writeResponse(rwc, Response{
			Status: 404,
		})
	}
}
