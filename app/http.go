package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func parseRequest(c net.Conn) (Request, error) {
	/** Request line
		GET /user-agent HTTP/1.1
		// <EMPTY_LINE> Headers start next
		Host: localhost:4221
		User-Agent: foobar/1.2.3  // Read this value
		Accept: ****
		// <EMPTY_LINE> Body starts next
		// Request body (empty)
	**/

	scanner := bufio.NewScanner(c)
	request := Request{
		Headers: make(map[string]string),
		Body:    make([]byte, 0),
	}

	// parse request line
	if scanner.Scan() {
		splits := strings.Split(scanner.Text(), " ")
		if len(splits) < 2 {
			return request, fmt.Errorf("invalid request")
		}
		request.Method = splits[0]
		request.Path = splits[1]
	}
	// CLRF
	ok := scanner.Scan()
	if !ok {
		return request, nil
	}

	// parse headers
	for scanner.Scan() {
		line := scanner.Text()

		fmt.Println(line)

		if line == "" {
			break
		}
		splits := strings.Split(line, ": ")
		header := strings.ToLower(splits[0])
		value := splits[1]
		fmt.Println("Adding header: ", header, " with value: ", value)
		request.Headers[header] = value
	}
	// CLRF
	// ok = scanner.Scan()
	// if !ok {
	// 	return request, nil
	// }

	return request, nil
}
