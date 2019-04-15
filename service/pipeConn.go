package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"net"
)

type Salt [aes.BlockSize]byte

func NewSalt() *Salt {
	s := new(Salt)
	if _, err := io.ReadFull(rand.Reader, s[:]); err != nil {
		return nil
	}

	return s
}

type PipeCryptKey [32]byte
type ProducerConn struct {
	IV Salt
	*service
	net.Conn
	cipher.Stream
}

func NewProducerConn(c net.Conn, aesKey PipeCryptKey) *ProducerConn {

	conn := &ProducerConn{
		Conn: c,
	}
	n, err := io.ReadFull(c, conn.IV[:])
	if err != nil {
		logger.Error("read salt for producer connection err:", err)
		return nil
	}

	if n != aes.BlockSize {
		logger.Error("read salt for producer connection length wrong:", n)
		return nil
	}

	block, err := aes.NewCipher(aesKey[:])
	if err != nil {
		logger.Error("create cipher for connection err:", err)
		return nil
	}

	conn.Stream = cipher.NewCFBDecrypter(block, conn.IV[:])

	return conn
}
