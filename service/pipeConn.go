package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"github.com/ribencong/go-youPipe/account"
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

	logger.Debugf("read salt:%02x", salt[:])

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
		err = fmt.Errorf("write empty data to sock client")
		logger.Warning(err)
		return
	}

	dataLen := uint32(len(buf))
	logger.Debugf("WriteCryptData before[%d]:%02x", dataLen, buf)
	c.Coder.XORKeyStream(buf, buf)

	headerBuf := UintToByte(dataLen)
	buf = append(headerBuf, buf...)

	logger.Debugf("WriteCryptData after[%d]:%02x", len(buf), buf)
	n, err = c.Write(buf)
	return
}

func UintToByte(val uint32) []byte {
	lenBuf := make([]byte, 4, 4)
	binary.BigEndian.PutUint32(lenBuf, val)
	return lenBuf
}

func ByteToUint(buff []byte) uint32 {
	return binary.BigEndian.Uint32(buff)
}

func (c *PipeConn) ReadCryptData(buf []byte) (n int, err error) {

	lenBuf := make([]byte, 4)
	if _, err = io.ReadFull(c, lenBuf); err != nil {
		logger.Warning(err)
		return
	}

	dataLen := ByteToUint(lenBuf)
	if dataLen == 0 || dataLen >= BuffSize {
		err = fmt.Errorf("wrong buffer size:%d", dataLen)
		logger.Warning(err)
		return
	}

	buf = buf[:dataLen]
	if _, err = io.ReadFull(c, buf); err != nil {
		logger.Warning(err)
		return
	}

	logger.Debugf("ReadCryptData before[%d]:%02x", dataLen, buf)
	c.Decoder.XORKeyStream(buf, buf)
	logger.Debugf("ReadCryptData after[%d]:%02x", dataLen, buf)
	n = int(dataLen)
	return
}
