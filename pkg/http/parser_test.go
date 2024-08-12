package http

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser(t *testing.T) {
	requestInput := "POST / HTTP/1.1\r\n Host: localhost:4221\r\n User-Agent: curl/8.4.0\r\n Accept: */*\r\nContent-Length: 16\r\n Content-Type: application/json\r\n \r\n {\"test\":\"value\"}"

	t.Logf("elo: %s", requestInput)

	reader := strings.NewReader(requestInput)

	request, err := parseRequest(reader)

	expectedRequest := Request{
		Method:        "POST",
		Proto:         "HTTP/1.1",
		ContentLength: 16,
		Headers: map[string]string{
			"host":           "localhost:4221",
			"user-agent":     "curl/8.4.0",
			"accept":         "*/*",
			"content-length": "16",
			"content-type":   "application/json",
		},
		Payload: []byte("{\"test\":\"value\"}"),
	}

	assert.Nil(t, err)
	assert.Equal(t, expectedRequest, request)
}
