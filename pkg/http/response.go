package http

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"

	"go.uber.org/zap/buffer"
)

type ResponseWriter interface {
	io.Writer

	SetStatus(int) error
}

type responseWriter struct {
	conn       net.Conn
	headers    map[string]string
	statusCode int

	buffer *bytes.Buffer
}

func newResponseWriter(conn net.Conn) *responseWriter {
	return &responseWriter{
		conn:    conn,
		headers: make(map[string]string),
		buffer:  &bytes.Buffer{},
	}
}

func (w *responseWriter) Write(message []byte) (int, error) {
	if w.statusCode == 0 {
		if err := w.SetStatus(200); err != nil {
			return 0, fmt.Errorf("could not set status: %w", err)
		}
	}

	contentLength := len(message)

	w.setContentLength(contentLength)
	w.setContentType("application/x-www-form-urlencoded")
	w.setHeader("Connection", "Keep-Alive")
	w.setHeader("Server", "go-simple-server")
	w.write([]byte("\r\n"))
	w.write(message)

	return w.conn.Write(w.buffer.Bytes())
}

func (w *responseWriter) setHeader(key string, value string) {
	buf := bytes.Buffer{}
	_, _ = buf.WriteString(key)
	_, _ = buf.Write([]byte(": "))
	_, _ = buf.WriteString(value)
	_, _ = buf.Write([]byte("\r\n"))

	w.write(buf.Bytes())
}

func (w *responseWriter) setContentType(contentType string) {
	w.setHeader("Content-Type", contentType)
}

func (w *responseWriter) setContentLength(length int) {
	w.setHeader("Content-Length", strconv.Itoa(length))
}

func (w *responseWriter) write(message []byte) {
	_, _ = w.buffer.Write(message)
}

func (w *responseWriter) SetStatus(statusCode int) error {
	w.statusCode = statusCode

	buf := buffer.Buffer{}
	_, _ = buf.Write([]byte("HTTP/1.1 "))

	_, _ = buf.WriteString(strconv.Itoa(statusCode))
	_ = buf.WriteByte(' ')
	_, _ = buf.WriteString(getStatus(statusCode))
	_, _ = buf.Write([]byte("\r\n"))

	_, err := w.conn.Write(buf.Bytes())
	return err
}

func getStatus(statusCode int) string {
	switch statusCode {
	case 200:
		return "OK"
	case 201:
		return "Created"
	case 202:
		return "Accepted"
	case 204:
		return "No Content"
	case 400:
		return "Bad Request"
	case 401:
		return "Unauthorized"
	case 403:
		return "Forbidden"
	case 404:
		return "Not Found"
	case 409:
		return "Conflict"
	case 418:
		return "I'm a teapot"
	default:
		return "UNKNOWN"
	}
}
