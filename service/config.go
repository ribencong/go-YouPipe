package service

import (
	"encoding/json"
	"fmt"
	"net"
)

const (
	buffSize         = 1 << 15
	DefaultKingKey   = "YP5rttHPzRsAe2RmF52sLzbBk4jpoPwJLtABaMv6qn7kVm"
	SocksServerPoint = "0.0.0.0"
)

var Config = SrvConf{
	KingKey:   DefaultKingKey,
	ServiceIP: SocksServerPoint,
}

type SrvConf struct {
	KingKey   string `json:"kingKey"`
	ServiceIP string `json:"accessPoint"`
}

func (conf SrvConf) String() string {
	return fmt.Sprintf("+%-15s:%40s+\n",
		"ServiceIP", conf.ServiceIP)
}

type ACK struct {
	Success bool
	Message string
}

type CtrlConn struct {
	net.Conn
}

func (conn *CtrlConn) ReadMsg(v interface{}) error {
	buffer := make([]byte, buffSize)
	n, err := conn.Read(buffer)
	if err != nil {
		err = fmt.Errorf("failed to read request:->%v", err)
		return err
	}

	if err = json.Unmarshal(buffer[:n], v); err != nil {
		err = fmt.Errorf("unmarshal address:->%v", err)
		return err
	}
	return nil
}

func (conn *CtrlConn) writeAck(err error) {
	var data []byte
	if err == nil {
		data, _ = json.Marshal(&ACK{
			Success: true,
			Message: "Success",
		})
	} else {
		data, _ = json.Marshal(&ACK{
			Success: false,
			Message: err.Error(),
		})
	}

	_, errW := conn.Write(data)
	if errW != nil || err != nil {
		conn.Close()
	}
}
