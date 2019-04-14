package service

import (
	"github.com/youpipe/go-youPipe/account"
	"github.com/youpipe/go-youPipe/thread"
	"net"
)

type customer struct {
	address    string
	onlineSig  chan struct{}
	license    *License
	pipeMng    *PipeAdmin
	payChannel *thread.Thread
}

func (node *SNode) newCustomer(conn net.Conn) {

	defer conn.Close()

	license, err := getValidLicense(conn, node)
	if err != nil {
		logger.Warning(err.Error())
		return
	}

	peerAddr := license.UserAddr
	user := &customer{
		address:    peerAddr,
		license:    license,
		pipeMng:    newAdmin(),
		payChannel: newMicroPayment(conn),
	}

	if err := account.GetAccount().CreateAesKey(&user.pipeMng.aesKey, peerAddr); err != nil {
		logger.Error("Aes key error when create customer", err)
		return
	}

	node.addCustomer(peerAddr, user)

	user.working()

	<-user.onlineSig

	user.destroy()

	node.removeUser(peerAddr)
}

func (c *customer) working() {
	c.onlineSig = make(chan struct{})
}

func (c *customer) destroy() {
}
