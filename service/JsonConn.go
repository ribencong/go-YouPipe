package service

import (
	"encoding/json"
	"fmt"
	"net"
)

type JsonConn struct {
	net.Conn
}

func (conn *JsonConn) Syn(v interface{}) error {
	if err := conn.WriteJsonMsg(v); err != nil {
		return err
	}

	ack := &YouPipeACK{}
	if err := conn.ReadJsonMsg(ack); err != nil {
		return err
	}

	if !ack.Success {
		return fmt.Errorf("create payment channel failed:%s", ack.Message)
	}

	return nil
}

//func (conn *JsonConn) Ack(v interface{}, ) error{
//
//
//	ack := &YouPipeACK{}
//
//	return nil
//}

func (conn *JsonConn) WriteJsonMsg(v interface{}) error {

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	if _, err := conn.Write(data); err != nil {
		return err
	}
	return nil
}

func (conn *JsonConn) ReadJsonMsg(v interface{}) error {
	buffer := make([]byte, BuffSize)
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

func (conn *JsonConn) writeAck(err error) {
	var data []byte
	if err == nil {
		data, _ = json.Marshal(&YouPipeACK{
			Success: true,
			Message: "Success",
		})
	} else {
		data, _ = json.Marshal(&YouPipeACK{
			Success: false,
			Message: err.Error(),
		})
	}

	_, errW := conn.Write(data)
	if errW != nil || err != nil {
		conn.Close()
	}
}
