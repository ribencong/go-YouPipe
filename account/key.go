package account

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/scrypt"
	"io"
)

const (
	PriKeyLen = 32
	PubKeyLen = 32
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

type PublicKey [PubKeyLen]byte
type PrivateKey [PriKeyLen]byte

type Key struct {
	PubKey    PublicKey
	rawPriKey PrivateKey
	PriKey    []byte
}

func GenerateKey(password string) (*Key, error) {

	k := &Key{}

	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	copy(k.rawPriKey[:], priv)

	curve25519.ScalarBaseMult((*[32]byte)(&(k.PubKey)), (*[32]byte)(&(k.rawPriKey)))

	aesKey, err := scrypt.Key([]byte(password), k.PubKey[:kp.S], kp.N, kp.R, kp.P, kp.L)
	if err != nil {
		logger.Warning("error to generate aes key:->", err)
		return nil, err
	}

	k.PriKey, err = Encrypt(aesKey, k.rawPriKey[:])
	if err != nil {
		logger.Warning("error to encrypt the raw private key:->", err)
		return nil, err
	}

	return k, nil
}

func Encrypt(key []byte, plainTxt []byte) ([]byte, error) {

	block, err := aes.NewCipher(key)
	if err != nil {
		logger.Warning("error to create cipher:->", err)
		return nil, err
	}

	cipherText := make([]byte, aes.BlockSize+len(plainTxt))

	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		logger.Warning("error to generate IV data:->", err)
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText, plainTxt)

	return cipherText, nil
}

func Decrypt(key []byte, cipherTxt []byte) ([]byte, error) {

	block, err := aes.NewCipher(key)
	if err != nil {
		logger.Warning("error to create cipher:->", err)
		return nil, err
	}

	if len(cipherTxt) < aes.BlockSize {
		return nil, errors.New("cipher text too short")
	}

	iv := cipherTxt[:aes.BlockSize]
	plainTxt := make([]byte, len(cipherTxt)-aes.BlockSize)

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(plainTxt, cipherTxt[aes.BlockSize:])

	return plainTxt, nil
}

func (k *Key) Unlock(password string) bool {

	aesKey, err := scrypt.Key([]byte(password), k.PubKey[:kp.S], kp.N, kp.R, kp.P, kp.L)
	if err != nil {
		logger.Warning("error to generate aes key:->", err)
		return false
	}

	raw, err := Decrypt(aesKey, k.PriKey)
	if err != nil {
		logger.Warning("error to unlock raw private key:->", err)
		return false
	}
	copy(k.rawPriKey[:], raw)

	return true
}

func (k *Key) ToNodeId() string {
	return AccPrefix + base58.Encode(k.PubKey[:])
}
