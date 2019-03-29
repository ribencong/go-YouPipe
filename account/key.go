package account

import (
	"crypto/rand"
	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/scrypt"
)

const (
	KeyLen    = 32
	AccPrefix = "YP"
)

var kp = struct {
	S int
	N int
	R int
	P int
	L int
}{
	S: 8,
	N: 1 << 15,
	R: 8,
	P: 1,
	L: 32,
}

type YouPipeKey [KeyLen]byte
type curve25519KeyPair struct {
	priKey YouPipeKey
	pubKey YouPipeKey
}

type ed25519KeyPair struct {
	eDPriKey ed25519.PrivateKey
	eDPubKey ed25519.PublicKey
}

type Key struct {
	*curve25519KeyPair
	*ed25519KeyPair
	LockedKey []byte
}

func getAESKey(salt []byte, password string) ([]byte, error) {
	return scrypt.Key([]byte(password), salt, kp.N, kp.R, kp.P, kp.L)
}

func GenerateKey(password string) (*Key, error) {

	pub, pri, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	k := &Key{
		ed25519KeyPair: &ed25519KeyPair{
			eDPubKey: pub,
			eDPriKey: pri,
		},
		curve25519KeyPair: curveKey(pri[:]),
	}

	aesKey, err := getAESKey(k.pubKey[:kp.S], password)
	if err != nil {
		logger.Warning("error to generate aes key:->", err)
		return nil, err
	}

	k.LockedKey, err = Encrypt(aesKey, k.eDPriKey[:])
	if err != nil {
		logger.Warning("error to encrypt the raw private key:->", err)
		return nil, err
	}

	return k, nil
}

func (k Key) ToNodeId() string {
	return AccPrefix + base58.Encode(k.pubKey[:])
}

func curveKey(data []byte) *curve25519KeyPair {
	ck := &curve25519KeyPair{}
	copy(ck.priKey[:], data[:KeyLen])
	curve25519.ScalarBaseMult((*[KeyLen]byte)(&ck.pubKey), (*[KeyLen]byte)(&ck.priKey))
	return ck
}

func edKey(data []byte) *ed25519KeyPair {
	ek := &ed25519KeyPair{
		eDPriKey: make([]byte, ed25519.PrivateKeySize),
		eDPubKey: make([]byte, ed25519.PublicKeySize),
	}
	copy(ek.eDPriKey[:], data[:])
	copy(ek.eDPubKey[:], data[32:])
	return ek
}

//
//func fillPrivateKey(k *Key, rawKey []byte) {
//	copy(k.eDPriKey[:], rawKey)
//	copy(k.eDPubKey[:], rawKey[32:])
//	copy(k.priKey[:], rawKey[:32])
//	curve25519.ScalarBaseMult((*[32]byte)(&k.pubKey), (*[32]byte)(&k.priKey))
//}
