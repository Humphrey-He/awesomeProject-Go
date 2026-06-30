package grpc_practices

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"io"
)

// gRPC wire framing: 1 byte compression flag + 4 bytes big-endian length + payload.
func WriteMessage(w io.Writer, compressed bool, payload []byte) error {
	flag := byte(0)
	if compressed {
		flag = 1
	}
	header := make([]byte, 5)
	header[0] = flag
	binary.BigEndian.PutUint32(header[1:], uint32(len(payload)))
	if _, err := w.Write(header); err != nil {
		return err
	}
	_, err := w.Write(payload)
	return err
}

func ReadMessage(r io.Reader, maxSize int) (bool, []byte, error) {
	header := make([]byte, 5)
	if _, err := io.ReadFull(r, header); err != nil {
		return false, nil, err
	}
	compressed := header[0] == 1
	length := int(binary.BigEndian.Uint32(header[1:]))
	if length < 0 || (maxSize > 0 && length > maxSize) {
		return false, nil, errors.New("message too large")
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return false, nil, err
	}
	return compressed, payload, nil
}

// Unary handler and interceptor chain (conceptual gRPC server pipeline).

type UnaryHandler func(ctx context.Context, req []byte) ([]byte, error)

type UnaryInterceptor func(ctx context.Context, req []byte, next UnaryHandler) ([]byte, error)

func ChainUnaryInterceptors(handler UnaryHandler, interceptors ...UnaryInterceptor) UnaryHandler {
	if len(interceptors) == 0 {
		return handler
	}
	wrapped := handler
	for i := len(interceptors) - 1; i >= 0; i-- {
		next := wrapped
		current := interceptors[i]
		wrapped = func(ctx context.Context, req []byte) ([]byte, error) {
			return current(ctx, req, next)
		}
	}
	return wrapped
}

// Metadata helpers.

type Metadata map[string]string

type metadataKey struct{}

func WithMetadata(ctx context.Context, md Metadata) context.Context {
	return context.WithValue(ctx, metadataKey{}, md)
}

func GetMetadata(ctx context.Context) Metadata {
	if md, ok := ctx.Value(metadataKey{}).(Metadata); ok {
		return md
	}
	return Metadata{}
}

// Convenience framer for buffered IO.

type Framer struct {
	Reader *bufio.Reader
	Writer io.Writer
}

func (f *Framer) Read(max int) (bool, []byte, error) {
	return ReadMessage(f.Reader, max)
}

func (f *Framer) Write(compressed bool, payload []byte) error {
	return WriteMessage(f.Writer, compressed, payload)
}
