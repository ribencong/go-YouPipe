package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
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

type PipeConn struct {
	IV *Salt
	*JsonConn
	Coder   cipher.Stream
	Decoder cipher.Stream
}

func NewProducerConn(c net.Conn, key PipeCryptKey) *PipeConn {
	salt := new(Salt)
	_, err := io.ReadFull(c, salt[:])
	if err != nil {
		logger.Error("read salt for producer connection err:", err)
		return nil
	}

	return newConn(c, key, salt)
}

func NewConsumerConn(c net.Conn, key PipeCryptKey) *PipeConn {
	salt := NewSalt()
	if salt == nil {
		logger.Error("read salt for consumer connection failed:")
		return nil
	}

	return newConn(c, key, salt)
}

func newConn(c net.Conn, key PipeCryptKey, salt *Salt) *PipeConn {
	if err := c.(*net.TCPConn).SetKeepAlive(true); err != nil {
		fmt.Println("set keepAlive for consumer connection err:", err)
		return nil
	}
	block, err := aes.NewCipher(key[:])
	if err != nil {
		fmt.Println("create cipher for connection err:", err)
		return nil
	}

	return &PipeConn{
		IV:       salt,
		JsonConn: &JsonConn{c},
		Coder:    cipher.NewCFBEncrypter(block, salt[:]),
		Decoder:  cipher.NewCFBDecrypter(block, salt[:]),
	}
}

func (c *PipeConn) WriteCryptData(buf []byte) (n int, err error) {
	if len(buf) == 0 {
		return
	}
	c.Coder.XORKeyStream(buf, buf)
	n, err = c.Write(buf)
	return
}

func (c *PipeConn) ReadCryptData(buf []byte) (n int, err error) {
	n, err = c.Read(buf)
	if err != nil {
		return
	}
	buf = buf[:n]
	c.Decoder.XORKeyStream(buf, buf)
	return
}
