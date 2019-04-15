package account

import (
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ed25519"
	"hash/fnv"
)

const (
	AccPrefix       = "YP"
	AccIDLen        = 40
	SocketPortInit  = 50000
	SocketPortRange = 15000
)

var (
	EInvalidID = fmt.Errorf("invalid ID")
)

type ID string

func (id ID) ToServerPort() uint16 {
	h := fnv.New32a()
	h.Write([]byte(id))
	sum := h.Sum32()
	return uint16(SocketPortInit + sum%SocketPortRange)
}

func (id ID) ToString() string {
	return string(id)
}

func (id ID) ToPubKey() ed25519.PublicKey {
	if len(id) <= len(AccPrefix) {
		return nil
	}
	ss := string(id[len(AccPrefix):])
	return base58.Decode(ss)
}

func (id ID) IsValid() bool {
	if len(id) <= AccIDLen {
		return false
	}
	if id[:len(AccPrefix)] != AccPrefix {
		return false
	}
	if len(id.ToPubKey()) != ed25519.PublicKeySize {
		return false
	}
	return true
}

func ConvertToID(addr string) (ID, error) {
	id := ID(addr)
	if id.IsValid() {
		return id, nil
	}
	return "", EInvalidID
}
