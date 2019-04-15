package service

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"net"
)

type reader struct {
	io.Reader
	cipher.Stream
	buf []byte
}

func NewReader(r io.Reader, s cipher.Stream) io.Reader {
	return &reader{Reader: r, Stream: s, buf: make([]byte, BuffSize)}
}

func (r *reader) Read(b []byte) (int, error) {

	n, err := r.Reader.Read(b)
	if err != nil {
		return 0, err
	}
	b = b[:n]
	r.XORKeyStream(b, b)
	return n, nil
}

func (r *reader) WriteTo(w io.Writer) (n int64, err error) {
	for {
		buf := r.buf
		nr, er := r.Read(buf)
		if nr > 0 {
			nw, ew := w.Write(buf[:nr])
			n += int64(nw)

			if ew != nil {
				err = ew
				return
			}
		}

		if er != nil {
			if er != io.EOF {
				err = er
			}
			return
		}
	}
}

type writer struct {
	io.Writer
	cipher.Stream
	buf []byte
}

func NewWriter(w io.Writer, s cipher.Stream) io.Writer {
	return &writer{Writer: w, Stream: s, buf: make([]byte, BuffSize)}
}

func (w *writer) ReadFrom(r io.Reader) (n int64, err error) {
	for {
		buf := w.buf
		nr, er := r.Read(buf)
		if nr > 0 {
			n += int64(nr)
			buf = buf[:nr]
			w.XORKeyStream(buf, buf)
			_, ew := w.Writer.Write(buf)
			if ew != nil {
				err = ew
				return
			}
		}

		if er != nil {
			if er != io.EOF { // ignore EOF as per io.ReaderFrom contract
				err = er
			}
			return
		}
	}
}

func (w *writer) Write(b []byte) (int, error) {
	n, err := w.ReadFrom(bytes.NewBuffer(b))
	return int(n), err
}

type shadowConn struct {
	net.Conn
	b cipher.Block
	r *reader
	w *writer
}

func Shadow(c net.Conn, k [32]byte) (net.Conn, error) {
	block, err := aes.NewCipher(k[:])
	if err != nil {
		logger.Warning("error to create cipher:->", err)
		return nil, err
	}

	return &shadowConn{Conn: c, b: block}, nil
}

func (c *shadowConn) initReader() error {
	if c.r == nil {
		buf := make([]byte, BuffSize)
		iv := buf[:aes.BlockSize]
		if _, err := io.ReadFull(c.Conn, iv); err != nil {
			return err
		}
		c.r = &reader{Reader: c.Conn, Stream: cipher.NewCFBDecrypter(c.b, iv), buf: buf}
	}
	return nil
}

func (c *shadowConn) initWriter() error {
	if c.w == nil {
		buf := make([]byte, BuffSize)
		iv := buf[:aes.BlockSize]
		if _, err := io.ReadFull(rand.Reader, iv); err != nil {
			return err
		}
		if _, err := c.Conn.Write(iv); err != nil {
			return err
		}
		c.w = &writer{Writer: c.Conn, Stream: cipher.NewCFBEncrypter(c.b, iv), buf: buf}
	}
	return nil
}

func (c *shadowConn) Read(b []byte) (int, error) {
	if c.r == nil {
		if err := c.initReader(); err != nil {
			return 0, err
		}
	}
	return c.r.Read(b)
}

func (c *shadowConn) ReadFrom(r io.Reader) (int64, error) {
	if c.w == nil {
		if err := c.initWriter(); err != nil {
			return 0, err
		}
	}
	return c.w.ReadFrom(r)
}

func (c *shadowConn) Write(b []byte) (int, error) {
	if c.w == nil {
		if err := c.initWriter(); err != nil {
			return 0, err
		}
	}
	return c.w.Write(b)
}

func (c *shadowConn) WriteTo(w io.Writer) (int64, error) {
	if c.r == nil {
		if err := c.initReader(); err != nil {
			return 0, err
		}
	}
	return c.r.WriteTo(w)
}
