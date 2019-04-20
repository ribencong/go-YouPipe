package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"github.com/youpipe/go-youPipe/account"
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

type PipeConn struct {
	IV *Salt
	net.Conn
	Coder   cipher.Stream
	Decoder cipher.Stream
}

func NewProducerConn(c net.Conn, key account.PipeCryptKey) *PipeConn {
	salt := new(Salt)
	_, err := io.ReadFull(c, salt[:])
	if err != nil {
		logger.Error("read salt for producer connection err:", err)
		return nil
	}

	logger.Debugf("read salt:%0x", salt[:])

	return newConn(c, key, salt)
}

func NewConsumerConn(c net.Conn, key account.PipeCryptKey) *PipeConn {
	salt := NewSalt()
	if salt == nil {
		logger.Error("read salt for consumer connection failed:")
		return nil
	}

	if _, err := c.Write(salt[:]); err != nil {
		logger.Error("Send salt to peer failed:", err)
		return nil
	}

	logger.Debugf("send salt:%0x", salt)

	return newConn(c, key, salt)
}

func newConn(c net.Conn, key account.PipeCryptKey, salt *Salt) *PipeConn {
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
		IV:      salt,
		Conn:    c,
		Coder:   cipher.NewCFBEncrypter(block, salt[:]),
		Decoder: cipher.NewCFBDecrypter(block, salt[:]),
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
