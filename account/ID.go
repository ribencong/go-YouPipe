package account

import (
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"hash/fnv"
)

const (
	AccPrefix       = "YP"
	SocketPortInit  = 50000
	SocketPortRange = 15000
)

var (
	EInvalidPre = fmt.Errorf("invalid ID prefix")
	EInvalidLen = fmt.Errorf("invalid ID length")
)

type ID string

func (id ID) ToSocketPort() uint16 {
	h := fnv.New32a()
	h.Write([]byte(id))
	sum := h.Sum32()
	return uint16(SocketPortInit + sum%SocketPortRange)
}

func (id ID) ToString() string {
	return string(id)
}

func (id ID) ToPubKey() []byte {
	if len(id) <= len(AccPrefix) {
		return nil
	}
	ss := string(id[len(AccPrefix):])
	return base58.Decode(ss)
}

func ConvertToID(addr string) (ID, error) {
	if addr[:len(AccPrefix)] != AccPrefix {
		return "", EInvalidPre
	}
	//if len(addr) != AccIDLen {
	//	return "", EInvalidLen
	//}

	return ID(addr), nil
}
