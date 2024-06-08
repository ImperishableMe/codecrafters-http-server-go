package main

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

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
		handleConnection(conn)
	}
}

func handleConnection(c net.Conn) {
	defer c.Close()
	request, err := parseRequest(c)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("Got request for path: ", request.Path)
	path := request.Path
	pattern := `^/echo/\w+$`
	re := regexp.MustCompile(pattern)

	if path == "/" {
		writeResponse(c, 200, "")
	} else if path == "/user-agent" {
		userAgent, ok := request.Headers["user-agent"]
		if !ok {
			writeResponse(c, 400, "User-Agent header is required")
			return
		}
		writeResponse(c, 200, userAgent)
	} else if re.MatchString(path) {
		str, _ := strings.CutPrefix(path, "/echo/")
		fmt.Println("Echoing back: ", str)
		writeResponse(c, 200, str)
	} else {
		writeResponse(c, 404, "")
	}
}

func writeResponse(c net.Conn, status int, body string) {
	responseString := map[int]string{
		200: "OK",
		404: "Not Found",
	}
	const LINE_BREAK = "\r\n"

	var respBuilder strings.Builder

	respBuilder.WriteString("HTTP/1.1")
	respBuilder.WriteString(fmt.Sprintf(" %d ", status))
	respBuilder.WriteString(responseString[status])
	respBuilder.WriteString(LINE_BREAK)
	if body != "" {
		respBuilder.WriteString(fmt.Sprintf("Content-Length: %d", len(body)))
		respBuilder.WriteString(LINE_BREAK)
		respBuilder.WriteString("Content-Type: text/plain")
		respBuilder.WriteString(LINE_BREAK)
	}
	respBuilder.WriteString(LINE_BREAK)
	if body != "" {
		respBuilder.WriteString(body)
		// respBuilder.WriteString(LINE_BREAK)
	}

	_, err := c.Write([]byte(respBuilder.String()))
	if err != nil {
		fmt.Println("Failed to write response...", err.Error())
		os.Exit(1)
	}
}
