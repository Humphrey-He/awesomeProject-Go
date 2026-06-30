package websocket_practices

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"io"
)

const websocketGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

func ComputeAcceptKey(clientKey string) string {
	h := sha1.Sum([]byte(clientKey + websocketGUID))
	return base64.StdEncoding.EncodeToString(h[:])
}

type Frame struct {
	Fin     bool
	Opcode  byte
	Masked  bool
	MaskKey [4]byte
	Payload []byte
}

func EncodeFrame(f Frame) ([]byte, error) {
	if len(f.Payload) > (1 << 31) {
		return nil, errors.New("payload too large")
	}
	first := f.Opcode
	if f.Fin {
		first |= 0x80
	}
	payloadLen := len(f.Payload)
	var header []byte
	var second byte
	if f.Masked {
		second |= 0x80
	}
	switch {
	case payloadLen <= 125:
		second |= byte(payloadLen)
		header = []byte{first, second}
	case payloadLen <= 65535:
		second |= 126
		header = []byte{first, second, byte(payloadLen >> 8), byte(payloadLen)}
	default:
		second |= 127
		header = []byte{first, second,
			0, 0, 0, 0,
			byte(payloadLen >> 24), byte(payloadLen >> 16), byte(payloadLen >> 8), byte(payloadLen),
		}
	}

	out := make([]byte, 0, len(header)+4+payloadLen)
	out = append(out, header...)
	payload := make([]byte, payloadLen)
	copy(payload, f.Payload)
	if f.Masked {
		out = append(out, f.MaskKey[:]...)
		for i := range payload {
			payload[i] ^= f.MaskKey[i%4]
		}
	}
	out = append(out, payload...)
	return out, nil
}

func DecodeFrame(r io.Reader) (Frame, error) {
	var f Frame
	header := make([]byte, 2)
	if _, err := io.ReadFull(r, header); err != nil {
		return f, err
	}
	f.Fin = header[0]&0x80 != 0
	f.Opcode = header[0] & 0x0f
	f.Masked = header[1]&0x80 != 0
	length := int(header[1] & 0x7f)
	if length == 126 {
		ext := make([]byte, 2)
		if _, err := io.ReadFull(r, ext); err != nil {
			return f, err
		}
		length = int(ext[0])<<8 | int(ext[1])
	} else if length == 127 {
		ext := make([]byte, 8)
		if _, err := io.ReadFull(r, ext); err != nil {
			return f, err
		}
		length = int(ext[4])<<24 | int(ext[5])<<16 | int(ext[6])<<8 | int(ext[7])
	}
	if f.Masked {
		if _, err := io.ReadFull(r, f.MaskKey[:]); err != nil {
			return f, err
		}
	}
	if length > 0 {
		f.Payload = make([]byte, length)
		if _, err := io.ReadFull(r, f.Payload); err != nil {
			return f, err
		}
		if f.Masked {
			for i := range f.Payload {
				f.Payload[i] ^= f.MaskKey[i%4]
			}
		}
	}
	return f, nil
}
