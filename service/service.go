package service

import (
	"github.com/youpipe/go-youPipe/thread"
)

type service struct {
	address    string
	onlineSig  chan struct{}
	license    *License
	pipeMng    *PipeAdmin
	payChannel *thread.Thread
}

func (node *PipeMiner) newCustomer(conn *JsonConn) {
	user, err := initCustomer(conn, node)

	if err != nil {
		logger.Warning(err.Error())
		conn.writeAck(err)
		return
	}

	user.working()

	<-user.onlineSig

	user.destroy()

	node.removeUser(user.address)
}

func (c *service) working() {
	c.onlineSig = make(chan struct{})
}

func (c *service) destroy() {
}
