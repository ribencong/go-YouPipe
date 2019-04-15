package service

import "golang.org/x/crypto/ed25519"

type BWCharger struct {
	*JsonConn
	priKey ed25519.PrivateKey
}

func monitoPipe() {

}
