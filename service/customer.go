package service

import (
	"github.com/youpipe/go-youPipe/thread"
)

type customer struct {
	address    string
	onlineSig  chan struct{}
	license    *License
	pipeMng    *PipeAdmin
	payChannel *thread.Thread
}

func (node *SNode) newCustomer(conn *CtrlConn) {
	user, err := initCustomer(conn, node)

	if err != nil {
		logger.Warning(err.Error())
		conn.writeAck(err)
		conn.Close()
		return
	}

	user.working()

	<-user.onlineSig

	user.destroy()

	node.removeUser(user.address)
}

func (c *customer) working() {
	c.onlineSig = make(chan struct{})
}

func (c *customer) destroy() {
}
