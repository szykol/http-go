package http

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
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

	startLine.method = strings.TrimSpace(string(splitted[0]))
	startLine.path = strings.TrimSpace(string(splitted[1]))
	startLine.proto = strings.TrimSpace(string(splitted[2]))

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

	reader := bufio.NewReader(rd)

	startLineStr, err := reader.ReadBytes('\n')
	if err != nil {
		return parsedRequest, fmt.Errorf("error scanning startline: %w", err)
	}

	startLine, err := parseStartLine(startLineStr)
	if err != nil {
		return parsedRequest, fmt.Errorf("error parsing start line: %w", err)
	}

	headers := make(map[string]string)

	for {
		line, err := reader.ReadBytes('\n')
		switch {
		case errors.Is(err, io.EOF):
			break
		case err != nil:
			return parsedRequest, fmt.Errorf("error parsing request: %w", err)
		}
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
	var payload []byte
	if contentLength > 0 {
		buf := make([]byte, contentLength)
		n, err := reader.Read(buf)
		if err != nil {
			return parsedRequest, fmt.Errorf("error parsing content: %w", err)
		}

		payload = buf[:n]
	}

	parsedRequest.Method = startLine.method
	parsedRequest.Proto = startLine.proto
	parsedRequest.Headers = headers
	parsedRequest.Payload = payload
	parsedRequest.ContentLength = contentLength
	parsedRequest.path = startLine.path

	return parsedRequest, nil
}
