package http

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
)

type Request struct {
	Method        string
	Proto         string
	ContentLength int

	Headers map[string]string
	Payload []byte

	path string
}

type startLine struct {
	method string
	path   string
	proto  string
}

func parseStartLine(line []byte) (startLine, error) {
	var startLine startLine
	splitted := bytes.Split(line, []byte(" "))
	if len(splitted) < 3 {
		return startLine, fmt.Errorf("invalid length of values in http start line")
	}

	startLine.method = string(splitted[0])
	startLine.path = string(splitted[1])
	startLine.proto = string(splitted[2])

	return startLine, nil
}

func getContentLength(headers map[string]string) int {
	length, ok := headers["content-length"]
	if !ok {
		return 0
	}

	value, err := strconv.Atoi(length)
	if err != nil {
		return 0
	}

	return value
}

func parseRequest(rd io.Reader) (Request, error) {
	var parsedRequest Request

	scanner := bufio.NewScanner(rd)

	if ok := scanner.Scan(); !ok {
		return parsedRequest, fmt.Errorf("error scanning startline: %w", scanner.Err())
	}

	startLine, err := parseStartLine(scanner.Bytes())
	if err != nil {
		return parsedRequest, fmt.Errorf("error parsing start line: %w", err)
	}

	headers := make(map[string]string)

	for scanner.Scan() {
		line := scanner.Bytes()
		line = bytes.TrimSpace(line)

		if len(line) == 0 {
			// stop on crlf
			break
		}

		splitted := bytes.SplitN(line, []byte(":"), 2)

		if len(splitted) != 2 {
			// ommit malformed headers
			continue
		}

		key := string(bytes.ToLower(bytes.TrimSpace(splitted[0])))
		value := string(bytes.ToLower(bytes.TrimSpace(splitted[1])))

		headers[key] = value
	}

	contentLength := getContentLength(headers)
	if contentLength > 0 {
		if ok := scanner.Scan(); !ok {
			return parsedRequest, fmt.Errorf("error scanning content: %w", scanner.Err())
		}
	}
	payload := bytes.TrimSpace(scanner.Bytes())

	parsedRequest.Method = startLine.method
	parsedRequest.Proto = startLine.proto
	parsedRequest.Headers = headers
	parsedRequest.Payload = payload
	parsedRequest.ContentLength = contentLength
	parsedRequest.path = startLine.path

	return parsedRequest, nil
}
