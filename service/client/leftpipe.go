package client

import (
	"github.com/youpipe/go-youPipe/service"
	"net"
)

type LeftPipe struct {
	*PayChannel
	done        chan error
	requestBuf  []byte
	responseBuf []byte
	proxyConn   net.Conn
	consume     *service.PipeConn
}

func NewPipe(l net.Conn, r *service.PipeConn, pay *PayChannel) *LeftPipe {
	return &LeftPipe{
		done:        make(chan error),
		requestBuf:  make([]byte, service.BuffSize),
		responseBuf: make([]byte, service.BuffSize),
		proxyConn:   l,
		consume:     r,
		PayChannel:  pay,
	}
}

func (p *LeftPipe) collectRequest() {

	for {
		nr, err := p.proxyConn.Read(p.requestBuf)
		if nr > 0 {
			if _, errW := p.consume.WriteCryptData(p.requestBuf[:nr]); errW != nil {
				p.done <- errW
				return
			}
		}
		if err != nil {
			p.done <- err
			return
		}
	}
}

func (p *LeftPipe) pullDataFromServer() {

	for {
		n, err := p.consume.ReadCryptData(p.responseBuf)

		if n > 0 {
			if _, errW := p.proxyConn.Write(p.responseBuf[:n]); errW != nil {
				p.done <- errW
				return
			}
		}

		if err != nil {
			p.done <- err
			return
		}

		p.PayChannel.Consume(n)
	}
}
