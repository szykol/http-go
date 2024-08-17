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
		w.SetStatus(200)
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

func (w *responseWriter) setHeader(key string, value string) (int, error) {
	buf := bytes.Buffer{}
	buf.WriteString(key)
	buf.Write([]byte(": "))
	buf.WriteString(value)
	buf.Write([]byte("\r\n"))

	return w.write(buf.Bytes())
}

func (w *responseWriter) setContentType(contentType string) (int, error) {
	return w.setHeader("Content-Type", contentType)
}

func (w *responseWriter) setContentLength(length int) (int, error) {
	return w.setHeader("Content-Length", strconv.Itoa(length))
}

func (w *responseWriter) write(message []byte) (int, error) {
	n, err := w.buffer.Write(message)
	if err != nil {
		return n, fmt.Errorf("Could not write response: %w", err)
	}

	return n, nil
}

func (w *responseWriter) SetStatus(statusCode int) error {
	w.statusCode = statusCode

	buf := buffer.Buffer{}
	buf.Write([]byte("HTTP/1.1 "))

	buf.WriteString(strconv.Itoa(statusCode))
	buf.WriteByte(' ')
	buf.WriteString(getStatus(statusCode))
	buf.Write([]byte("\r\n"))

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
