package http

import (
	"bufio"
	"bytes"
	"io"
)

type Request struct {
	StartLine string
	headers   map[string]string
	payload   []byte
}

func parseRequest(rd io.Reader) (Request, error) {
	var parsedRequest Request

	scanner := bufio.NewScanner(rd)

	scanner.Scan()
	startLine := scanner.Bytes()

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

	scanner.Scan()
	payload := scanner.Bytes()

	parsedRequest.StartLine = string(startLine)
	parsedRequest.headers = headers
	parsedRequest.payload = bytes.TrimSpace(payload)

	return parsedRequest, nil
}
