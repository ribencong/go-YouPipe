package client

import (
	"fmt"
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
		_, errW := p.consume.WriteCryptData(p.requestBuf[:nr])
		if errW != nil || err != nil {
			p.done <- fmt.Errorf("errW:%v, err:%v", errW, err)
			return
		}
	}
}

func (p *LeftPipe) pullDataFromServer() {

	for {
		n, err := p.consume.ReadCryptData(p.responseBuf)

		if n > 0 {
			_, errW := p.proxyConn.Write(p.responseBuf[:n])
			p.done <- errW
			return
		}

		if err != nil {
			p.done <- err
		}
		p.consume(n)
	}
}
