package service

import (
	"context"
	"fmt"
	"github.com/youpipe/go-youPipe/thread"
	"net"
)

type MicroPayChannel struct {
	net.Conn
}

func newMicroPayment(c net.Conn) *thread.Thread {
	c.(*net.TCPConn).SetKeepAlive(true)

	rAddr := c.RemoteAddr().String()
	p := &MicroPayChannel{
		Conn: c,
	}

	t := thread.NewThreadWithName(p, rAddr)
	logger.Debugf("new service(%s) coming.", t.Name)
	return t
}

func (cw *MicroPayChannel) CloseCallBack(t *thread.Thread) {
	cw.Close()
	//if len(cw.customerId) == 0 {
	//	return
	//}
	//
	//if u := cw.getCustomer(cw.customerId); u != nil {
	//
	//	u.removePipe(cw.pipeId)
	//
	//	if u.isPipeEmpty() {
	//		cw.removeUser(cw.customerId)
	//	}
	//}
}

func (cw *MicroPayChannel) DebugInfo() string {
	return fmt.Sprintf("\n||||||||||||||||||||||||||||||||||||||||||||||||\n"+
		//"|%s|\n"+
		"|%-15s:%30s|\n"+
		"||||||||||||||||||||||||||||||||||||||||||||||||",
		//cw.RemoteAddr().String(),
		"remoteIP", cw.RemoteAddr().String())
}

func (cw *MicroPayChannel) Run(ctx context.Context) {
}
