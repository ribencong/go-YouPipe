package account

import (
	"crypto/rand"
	"crypto/sha512"
	"github.com/btcsuite/btcutil/base58"
	"github.com/youpipe/go-youPipe/account/edwards25519"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/scrypt"
)

const (
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

type Key struct {
	priKey    ed25519.PrivateKey
	PubKey    ed25519.PublicKey
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
		priKey: pri,
		PubKey: pub,
	}

	aesKey, err := getAESKey(k.PubKey[:kp.S], password)
	if err != nil {
		logger.Warning("error to generate aes key:->", err)
		return nil, err
	}

	k.LockedKey, err = Encrypt(aesKey, k.priKey[:])
	if err != nil {
		logger.Warning("error to encrypt the raw private key:->", err)
		return nil, err
	}

	return k, nil
}

func (k Key) ToNodeId() string {
	return AccPrefix + base58.Encode(k.PubKey[:])
}

func populateKey(data []byte) (ed25519.PublicKey, ed25519.PrivateKey) {
	pri := ed25519.PrivateKey(data)
	pub := pri.Public().(ed25519.PublicKey)
	return pub, pri
}

func PrivateKeyToCurve25519(curve25519Private *[32]byte, privateKey *[64]byte) {
	h := sha512.New()
	h.Write(privateKey[:32])
	digest := h.Sum(nil)

	digest[0] &= 248
	digest[31] &= 127
	digest[31] |= 64

	copy(curve25519Private[:], digest)
}

func edwardsToMontgomeryX(outX, y *edwards25519.FieldElement) {
	// We only need the x-coordinate of the curve25519 point, which I'll
	// call u. The isomorphism is u=(y+1)/(1-y), since y=Y/Z, this gives
	// u=(Y+Z)/(Z-Y). We know that Z=1, thus u=(Y+1)/(1-Y).
	var oneMinusY edwards25519.FieldElement
	edwards25519.FeOne(&oneMinusY)
	edwards25519.FeSub(&oneMinusY, &oneMinusY, y)
	edwards25519.FeInvert(&oneMinusY, &oneMinusY)

	edwards25519.FeOne(outX)
	edwards25519.FeAdd(outX, outX, y)

	edwards25519.FeMul(outX, outX, &oneMinusY)
}

func PublicKeyToCurve25519(curve25519Public *[32]byte, publicKey *[32]byte) bool {
	var A edwards25519.ExtendedGroupElement
	if !A.FromBytes(publicKey) {
		return false
	}

	// A.Z = 1 as a postcondition of FromBytes.
	var x edwards25519.FieldElement
	edwardsToMontgomeryX(&x, &A.Y)
	edwards25519.FeToBytes(curve25519Public, &x)
	return true
}
