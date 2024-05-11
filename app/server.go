package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	fmt.Printf("Listening on: %v", l.Addr().String())

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	fmt.Println("Accepted connection from: ", conn.RemoteAddr().String())

	handleConnection(conn)
}

func handleConnection(c net.Conn) {
	path, err := readPath(c)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Println("Got request for path: ", path)

	if path == "/" {
		writeResponse(c, 200)
	} else {
		writeResponse(c, 404)
	}

	err = c.Close()
	if err != nil {
		fmt.Println("Failed to close connection...", err.Error())
		os.Exit(1)
	}
}

func readPath(c net.Conn) (string, error) {
	scanner := bufio.NewScanner(c)
	if scanner.Scan() {
		splits := strings.Split(scanner.Text(), " ")
		if len(splits) < 2 {
			return "", fmt.Errorf("Invalid request")
		}
		path := splits[1]
		// read the rest of the request
		scanner.Scan()
		scanner.Scan()
		return path, nil
	}
	return "", fmt.Errorf("Failed to read request")
}

func writeResponse(c net.Conn, status int) {
	responseString := map[int]string{
		200: "OK",
		404: "Not Found",
	}
	const LINE_BREAK = "\r\n"
	_, err := c.Write([]byte(fmt.Sprintf("HTTP/1.1 %d %s%s%s", status, responseString[status], LINE_BREAK, LINE_BREAK)))
	if err != nil {
		fmt.Println("Failed to write response...", err.Error())
		os.Exit(1)
	}
}
