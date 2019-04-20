package service

import (
	"encoding/json"
	"fmt"
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
		mineBuf:     make([]byte, BuffSize),
		serverBuf:   make([]byte, BuffSize),
		serverConn:  r,
		chargeConn:  l,
		bandCharger: charger,
	}
}

type RightPipe struct {
	mineBuf    []byte
	serverBuf  []byte
	serverConn net.Conn
	chargeConn *PipeConn
	*bandCharger
}

func (p *RightPipe) listenRequest() {
	defer p.expire()

	for {
		n, err := p.chargeConn.ReadCryptData(p.serverBuf)
		logger.Debugf("request from(%s) to %s", p.peerID, p.serverConn.RemoteAddr().String())
		if n > 0 {
			if _, errW := p.serverConn.Write(p.serverBuf[:n]); errW != nil {
				logger.Warning("forward request err:", errW, p.peerID)
				return
			}
		}
		if err != nil {
			return
		}
	}
}

func (p *RightPipe) pushBackToClient() {
	defer p.expire()

	for {
		n, err := p.serverConn.Read(p.serverBuf)
		logger.Debugf("pull data(no:%d, err:%v) for client:%s", n, err, p.peerID)
		if n > 0 {
			if _, errW := p.chargeConn.WriteCryptData(p.serverBuf[:n]); errW != nil {
				logger.Warning("forward response err:", errW, p.peerID)
				return
			}
		}

		if err != nil {
			return
		}

		if err := p.bandCharger.Charge(n); err != nil {
			return
		}
	}
}

func (p *RightPipe) expire() {
	p.chargeConn.SetDeadline(time.Now())
	p.serverConn.SetDeadline(time.Now())
}

func (p *RightPipe) String() string {
	return fmt.Sprintf("%s<->%s",
		p.chargeConn.RemoteAddr().String(),
		p.serverConn.RemoteAddr().String())
}
