package http_server_internals

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

type Request struct {
	Method  string
	Path    string
	Proto   string
	Headers map[string]string
	Body    []byte
}

type Response struct {
	StatusCode int
	StatusText string
	Headers    map[string]string
	Body       []byte
	Close      bool
}

var statusText = map[int]string{
	200: "OK",
	400: "Bad Request",
	404: "Not Found",
	500: "Internal Server Error",
}

func ParseRequest(r *bufio.Reader, maxBody int64) (*Request, error) {
	line, err := readLine(r)
	if err != nil {
		return nil, err
	}
	parts := strings.SplitN(line, " ", 3)
	if len(parts) != 3 {
		return nil, errors.New("invalid request line")
	}
	req := &Request{
		Method:  parts[0],
		Path:    parts[1],
		Proto:   parts[2],
		Headers: map[string]string{},
	}

	for {
		h, err := readLine(r)
		if err != nil {
			return nil, err
		}
		if h == "" {
			break
		}
		idx := strings.Index(h, ":")
		if idx <= 0 {
			return nil, errors.New("invalid header line")
		}
		key := strings.TrimSpace(h[:idx])
		val := strings.TrimSpace(h[idx+1:])
		req.Headers[strings.ToLower(key)] = val
	}

	if cl, ok := req.Headers["content-length"]; ok {
		length, err := strconv.ParseInt(cl, 10, 64)
		if err != nil || length < 0 {
			return nil, errors.New("invalid content-length")
		}
		if maxBody > 0 && length > maxBody {
			return nil, errors.New("body too large")
		}
		if length > 0 {
			buf := make([]byte, length)
			if _, err := io.ReadFull(r, buf); err != nil {
				return nil, err
			}
			req.Body = buf
		}
	}

	return req, nil
}

func BuildResponse(resp *Response) []byte {
	var buf bytes.Buffer
	status := resp.StatusText
	if status == "" {
		status = statusText[resp.StatusCode]
		if status == "" {
			status = "Unknown"
		}
	}
	buf.WriteString(fmt.Sprintf("HTTP/1.1 %d %s\r\n", resp.StatusCode, status))

	headers := map[string]string{}
	for k, v := range resp.Headers {
		headers[k] = v
	}
	if _, ok := headers["Date"]; !ok {
		headers["Date"] = time.Now().UTC().Format(time.RFC1123)
	}
	if _, ok := headers["Content-Length"]; !ok {
		headers["Content-Length"] = strconv.Itoa(len(resp.Body))
	}
	if resp.Close {
		headers["Connection"] = "close"
	}
	for k, v := range headers {
		buf.WriteString(k)
		buf.WriteString(": ")
		buf.WriteString(v)
		buf.WriteString("\r\n")
	}
	buf.WriteString("\r\n")
	buf.Write(resp.Body)
	return buf.Bytes()
}

func WriteResponse(w io.Writer, resp *Response) error {
	_, err := w.Write(BuildResponse(resp))
	return err
}

func HandleConn(conn net.Conn, handler func(*Request) *Response, maxBody int64) error {
	defer conn.Close()
	r := bufio.NewReader(conn)
	req, err := ParseRequest(r, maxBody)
	if err != nil {
		return err
	}
	resp := handler(req)
	if resp == nil {
		resp = &Response{StatusCode: 500, Body: []byte("nil response"), Close: true}
	}
	return WriteResponse(conn, resp)
}

func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}
