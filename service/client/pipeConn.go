package client

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"github.com/youpipe/go-youPipe/service"
	"net"
)

type ConsumerConn struct {
	IV *service.Salt
	*service.JsonConn
	cipher.Stream
}

func NewConsumerConn(c net.Conn, key []byte) *ConsumerConn {

	salt := service.NewSalt()
	if salt == nil {
		return nil
	}

	if err := c.(*net.TCPConn).SetKeepAlive(true); err != nil {
		fmt.Println("set keepAlive for consumer connection err:", err)
		return nil
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println("create cipher for connection err:", err)
		return nil
	}

	return &ConsumerConn{
		IV:       salt,
		JsonConn: &service.JsonConn{Conn: c},
		Stream:   cipher.NewCFBDecrypter(block, salt[:]),
	}
}

func (c *ConsumerConn) WriteCryptData(buf []byte) (n int, err error) {
	if len(buf) == 0 {
		return
	}
	c.XORKeyStream(buf, buf)
	n, err = c.Write(buf)
	return
}

func (c *ConsumerConn) ReadCryptData(buf []byte) (n int, err error) {
	n, err = c.Read(buf)
	if err != nil {
		return
	}
	buf = buf[:n]
	c.XORKeyStream(buf, buf)
	return
}
