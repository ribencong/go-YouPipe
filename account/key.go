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

type curve25519KeyPair struct {
	priKey [KeyLen]byte
	pubKey [KeyLen]byte
}

type ed25519KeyPair struct {
	eDPriKey [ed25519.PrivateKeySize]byte
	eDPubKey [ed25519.PublicKeySize]byte
}

type Key struct {
	curve25519KeyPair
	ed25519KeyPair
	LockedKey []byte
}

func getAESKey(salt []byte, password string) ([]byte, error) {
	return scrypt.Key([]byte(password), salt, kp.N, kp.R, kp.P, kp.L)
}

func GenerateKey(password string) (*Key, error) {

	k := &Key{}
	_, priAsSeed, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	k.fillPrivateKey(priAsSeed)

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

func (k *Key) fillPrivateKey(rawKey []byte) {
	copy(k.eDPriKey[:], rawKey)
	copy(k.eDPubKey[:], rawKey[32:])
	copy(k.priKey[:], rawKey[:32])
	curve25519.ScalarBaseMult(&k.pubKey, &k.priKey)
}

func (k *Key) Lock() {
	k.priKey = [KeyLen]byte{0}
	k.pubKey = [KeyLen]byte{0}
	k.eDPriKey = [ed25519.PrivateKeySize]byte{0}
	k.eDPubKey = [ed25519.PublicKeySize]byte{0}
}

func (k *Key) Unlock(password string) bool {
	aesKey, err := getAESKey(k.pubKey[:kp.S], password) //scrypt.Key([]byte(password), k.PubKey[:kp.S], kp.N, kp.R, kp.P, kp.L)
	if err != nil {
		logger.Warning("error to generate aes key:->", err)
		return false
	}

	raw, err := Decrypt(aesKey, k.LockedKey)
	if err != nil {
		logger.Warning("error to unlock raw private key:->", err)
		return false
	}

	k.fillPrivateKey(raw)
	return true
}

func (k *Key) ToNodeId() string {
	return AccPrefix + base58.Encode(k.pubKey[:])
}
