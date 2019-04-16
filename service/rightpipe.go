package service

import (
	"encoding/json"
	"github.com/youpipe/go-youPipe/account"
	"golang.org/x/crypto/ed25519"
	"net"
	"time"
)

type PipeReqData struct {
	Addr   string
	Target string
}

func (s *PipeReqData) Verify(sig []byte) bool {
	msg, err := json.Marshal(s)
	if err != nil {
		return false
	}

	pid, err := account.ConvertToID(s.Addr)
	if err != nil {
		return false
	}
	return ed25519.Verify(pid.ToPubKey(), msg, sig)
}

func NewPipe(l *PipeConn, r net.Conn, charger *bandCharger) *RightPipe {

	return &RightPipe{
		done:        make(chan error),
		mineBuf:     make([]byte, BuffSize),
		serverBuf:   make([]byte, BuffSize),
		serverConn:  r,
		chargeConn:  l,
		bandCharger: charger,
	}
}

type RightPipe struct {
	done       chan error
	mineBuf    []byte
	serverBuf  []byte
	serverConn net.Conn
	chargeConn *PipeConn
	*bandCharger
}

func (p *RightPipe) listenRequest() {
	for {

		n, err := p.chargeConn.ReadCryptData(p.serverBuf)
		if n > 0 {
			if _, errW := p.serverConn.Write(p.serverBuf[:n]); errW != nil {
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

func (p *RightPipe) pushBackToClient() {
	for {

		n, err := p.serverConn.Read(p.serverBuf)
		if n > 0 {
			if _, errW := p.chargeConn.WriteCryptData(p.serverBuf[:n]); errW != nil {
				p.done <- errW
				return
			}
		}

		if err != nil {
			p.done <- err
			return
		}

		if err := p.bandCharger.Charge(n); err != nil {
			p.done <- err
			return
		}
	}
}

func (p *RightPipe) expire() {
	p.chargeConn.SetDeadline(time.Now())
	p.serverConn.SetDeadline(time.Now())
}
