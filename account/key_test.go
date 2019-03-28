package account

import (
	"encoding/base64"
	"golang.org/x/crypto/scrypt"
	"testing"
)

var passwrod = "12345678"

func TestScrypt(t *testing.T) {
	salt := []byte{0xc8, 0x28, 0xf2, 0x58, 0xa7, 0x6a, 0xad, 0x7b}

	dk, err := scrypt.Key([]byte("some password"), salt, 1<<15, 8, 1, 32)
	if err != nil {
		t.Fatal(err)
	}
	t.Error(base64.StdEncoding.EncodeToString(dk))
}

func TestGenerateKeyPair(t *testing.T) {
	key, err := GenerateKey(passwrod)
	if err != nil {
		t.Error(err)
	}

	printKey(t, key)
	key.Lock()
	printKey(t, key)
	key.Unlock(passwrod)
	printKey(t, key)
}
func printKey(t *testing.T, key *Key) {
	t.Error("priKey:", key.priKey, "len", len(key.priKey))
	t.Error("pubKey:", key.pubKey, "len", len(key.pubKey))
	t.Error("eDPriKey:", key.eDPriKey, "len", len(key.eDPriKey))
	t.Error("eDPubKey:", key.eDPubKey, "len", len(key.eDPubKey))
	t.Error("LockedKey:", key.LockedKey, "len", len(key.LockedKey))
}
