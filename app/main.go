package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/codecrafters-io/http-server-starter-go/app/http"
)

func main() {
	fmt.Println("Logs from your program will appear here!")
	serverOptions := http.ServerOptions{}
	if len(os.Args) >= 3 {
		serverOptions.FileServerRoot = os.Args[2]
	}
	s := http.NewServer(serverOptions)
	registerHandlers(&s)
	// blocks indefinitely
	s.ListenAndServe()
}

func registerHandlers(s *http.Server) {

	// Register the handlers
	s.Register(
		"GET /", http.HandlerFunc(
			func(r *http.Request, w http.ResponseWriter) {
				w.Write([]byte(""))
			}))
	s.Register(
		"GET /user-agent", http.HandlerFunc(
			func(r *http.Request, w http.ResponseWriter) {
				userAgent, ok := r.Headers["user-agent"]
				if !ok || len(userAgent) != 1 {
					w.WriteHeader(400)
					w.Write([]byte("User-Agent header is required"))
					return
				}
				w.Write([]byte(userAgent[0]))
			}))
	s.Register(
		"GET /echo/{str}", http.HandlerFunc(
			func(r *http.Request, w http.ResponseWriter) {
				str, _ := strings.CutPrefix(r.Path, "/echo/")
				fmt.Println("echoing back: ", str)
				w.Write([]byte(str))
			}))
	s.Register(
		"GET /files/{path}", http.HandlerFunc(
			func(r *http.Request, w http.ResponseWriter) {
				path, _ := strings.CutPrefix(r.Path, "/files/")
				fmt.Println("Asking for file ", path)
				filePath := filepath.Join(s.FileServerRoot, path)
				fileBody, err := os.ReadFile(filePath)
				if err != nil {
					w.WriteHeader(404)
					w.Write(nil)
					return
				}
				w.Headers()["Content-Type"] = "application/octet-stream"
				w.Write(fileBody)
			}))
	s.Register(
		"POST /files/{path}", http.HandlerFunc(
			func(r *http.Request, w http.ResponseWriter) {
				fmt.Println(r.Headers)
				path, _ := strings.CutPrefix(r.Path, "/files/")
				filePath := filepath.Join(s.FileServerRoot, path)
				file, err := os.Create(filePath)
				if err != nil {
					w.WriteHeader(400)
					return
				}
				defer file.Close()

				lenStr, ok := r.GetHeader("Content-Length")
				if !ok || len(lenStr) != 1 {
					w.WriteHeader(400)
					w.Write([]byte("missing Content-Length header."))
					return
				}
				fileSize, _ := strconv.Atoi(lenStr[0])
				fmt.Println("Reading total file size of ", fileSize)

				written, err := io.CopyN(file, r.Body, int64(fileSize))
				fmt.Println("Read ", written, " bytes")
				if err != nil || written != int64(fileSize) {
					w.WriteHeader(400)
					w.Write(
						[]byte(fmt.Sprintf(
							"couldn't write the whole content. Wrote %d, expected %d", written, fileSize)))
					return
				}
				w.WriteHeader(201)
			}))
}
