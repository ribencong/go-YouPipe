package account

import (
	"github.com/btcsuite/btcutil/base58"
	"hash/fnv"
)

const (
	AccPrefix       = "YP"
	SocketPortInit  = 50000
	SocketPortRange = 15000
)

type ID string

func (id ID) ToSocketPort() uint16 {
	h := fnv.New32a()
	h.Write([]byte(id))
	sum := h.Sum32()
	return uint16(SocketPortInit + sum%SocketPortRange)
}

func (id ID) ToPubKey() []byte {
	if len(id) <= len(AccPrefix) {
		return nil
	}
	ss := string(id[len(AccPrefix):])
	return base58.Decode(ss)
}
