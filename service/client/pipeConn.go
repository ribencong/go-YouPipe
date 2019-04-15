package client

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"github.com/youpipe/go-youPipe/service"
	"net"
)

type ConsumerConn struct {
	IV service.Salt
	*Client
	net.Conn
	cipher.Stream
}

func NewConsumerConn(c net.Conn, iv []byte, cli *Client) *ConsumerConn {

	conn := &ConsumerConn{
		Conn:   c,
		Client: cli,
	}
	if len(iv) != aes.BlockSize {
		return nil
	}

	copy(conn.IV[:], iv)

	if err := c.(*net.TCPConn).SetKeepAlive(true); err != nil {
		fmt.Println("set keepAlive for consumer connection err:", err)
		return nil
	}
	block, err := aes.NewCipher(conn.connKey[:])
	if err != nil {
		fmt.Println("create cipher for connection err:", err)
		return nil
	}

	conn.Stream = cipher.NewCFBDecrypter(block, iv)
	return conn
}
