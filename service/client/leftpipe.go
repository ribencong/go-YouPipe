package client

import (
	"fmt"
	"github.com/youpipe/go-youPipe/service"
	"net"
	"time"
)

type LeftPipe struct {
	*PayChannel
	target      string
	requestBuf  []byte
	responseBuf []byte
	proxyConn   net.Conn
	consume     *service.PipeConn
}

func NewPipe(l net.Conn, r *service.PipeConn, pay *PayChannel, tgt string) *LeftPipe {
	return &LeftPipe{
		target:      tgt,
		requestBuf:  make([]byte, service.BuffSize),
		responseBuf: make([]byte, service.BuffSize),
		proxyConn:   l,
		consume:     r,
		PayChannel:  pay,
	}
}

func (p *LeftPipe) collectRequest() {
	defer p.expire()
	for {
		nr, err := p.proxyConn.Read(p.requestBuf)
		if nr > 0 {
			if _, errW := p.consume.WriteCryptData(p.requestBuf[:nr]); errW != nil {
				return
			}
		}
		if err != nil {
			return
		}
	}
}

func (p *LeftPipe) pullDataFromServer() {
	defer p.expire()
	for {
		n, err := p.consume.ReadCryptData(p.responseBuf)

		fmt.Printf("\n\n Pull data(no:%d, err:%v) from server:%s\n", n, err,
			p.consume.RemoteAddr().String())

		if n > 0 {
			if _, errW := p.proxyConn.Write(p.responseBuf[:n]); errW != nil {
				return
			}
		}

		if err != nil {
			return
		}

		p.PayChannel.Consume(n)
	}
}

func (p *LeftPipe) expire() {
	p.consume.SetDeadline(time.Now())
	p.proxyConn.SetDeadline(time.Now())
}

func (p *LeftPipe) String() string {
	return fmt.Sprintf("%s<->%s",
		p.proxyConn.RemoteAddr().String(),
		p.consume.RemoteAddr().String())
}
