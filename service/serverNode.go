package service

import (
	"encoding/json"
	"fmt"
	"github.com/op/go-logging"
	"github.com/youpipe/go-youPipe/account"
	"github.com/youpipe/go-youPipe/network"
	"github.com/youpipe/go-youPipe/utils"
	"net"
	"sync"
	"time"
)

var (
	instance  *SNode = nil
	once      sync.Once
	logger, _ = logging.GetLogger(utils.LMService)
)

type SNode struct {
	sync.RWMutex
	serviceConn  net.Listener
	microPayConn net.Listener
	users        map[string]*customer
}

func GetSNode() *SNode {
	once.Do(func() {
		instance = newServiceNode()
	})

	return instance
}

func newServiceNode() *SNode {
	id := account.GetAccount().Address
	servicePort := id.ToSocketPort()
	addr := network.JoinHostPort(Config.ServiceIP, servicePort)

	l, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	logger.Infof("Socks5 server listening TCP on %s", addr)

	payPort := servicePort + 1
	addr = network.JoinHostPort(Config.ServiceIP, payPort)
	p, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	logger.Infof("MicroPayment chanel starting at: %s", addr)

	node := &SNode{
		serviceConn:  l,
		microPayConn: p,
		users:        make(map[string]*customer),
	}

	return node
}
func (node *SNode) OpenPaymentChannel() {
	defer node.microPayConn.Close()
	for {
		conn, err := node.serviceConn.Accept()
		if err != nil {
			panic(err)
		}

		go node.newCustomer(conn)
	}
}
func (node *SNode) Mining() {
	defer node.serviceConn.Close()

	for {
		conn, err := node.serviceConn.Accept()
		if err != nil {
			panic(err)
		}

		node.newWaiter(conn).Start()
	}
}

func (node *SNode) newCustomer(conn net.Conn) {
	defer conn.Close()

	content, err := node.getValidLicense(conn)
	if err != nil {
		logger.Warning(err)
		data, _ := json.Marshal(&LicenseCheckResult{
			Success: false,
			ErrMsg:  err.Error(),
		})
		conn.Write(data)
		return
	}

	user := node.CreateCustomer(content.UserAddr)
	user.license = content
	user.Conn = conn
	user.StartProvePay()
}

func (node *SNode) getValidLicense(conn net.Conn) (*LicenseContent, error) {

	buff := make([]byte, buffSize)
	n, err := conn.Read(buff)
	if err != nil {
		return nil, fmt.Errorf("payment channel open err:%v", err)
	}

	l := &License{}
	if err = json.Unmarshal(buff[:n], l); err != nil {
		return nil, fmt.Errorf("payment channel read data err:%v", err)

	}
	if !CheckLicense(l) {
		return nil, fmt.Errorf("signature failed err:%v", err)
	}
	content := &LicenseContent{}
	if err = json.Unmarshal(l.Content, content); err != nil {
		return nil, fmt.Errorf("parse license content err:%v", err)
	}

	peerAddr := content.UserAddr
	if nil != node.getCustomer(peerAddr) {
		return nil, fmt.Errorf("duplicate payment channel (%s) err:%v", peerAddr, err)
	}

	now := time.Now()
	if now.Before(content.StartDate) || now.After(content.EndDate) {
		return nil, fmt.Errorf("license time invalid(%s)", peerAddr)
	}

	return content, nil
}

func (node *SNode) CreateCustomer(peerId string) *customer {
	node.Lock()
	defer node.Unlock()
	if u, ok := node.users[peerId]; ok {
		return u
	}

	user := &customer{
		address: peerId,
		pipes:   make(map[string]*Pipe),
	}

	if err := account.GetAccount().CreateAesKey(&user.aesKey, peerId); err != nil {
		return nil
	}

	node.users[peerId] = user
	return user
}

func (node *SNode) getCustomer(peerId string) *customer {
	node.RLock()
	defer node.RUnlock()
	return node.users[peerId]
}

func (node *SNode) removeUser(peerId string) {
	node.Lock()
	defer node.Unlock()
	delete(node.users, peerId)
	logger.Debugf("remove user(%s)", peerId)
}
