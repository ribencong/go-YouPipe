package account

import (
	"bytes"
	"testing"
)

var password = "12345678"
var salt = []byte{0xc8, 0x28, 0xf2, 0x58, 0xa7, 0x6a, 0xad, 0x7b}
var plainTxt = "hello word"

func TestAesEncrypt(t *testing.T) {
	aesKey, err := getAESKey(salt, password)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(aesKey)

	cp, err := Encrypt(aesKey, []byte(plainTxt))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(cp)

	deStr, err := Decrypt(aesKey, cp)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(deStr))

	if plainTxt != string(deStr) {
		t.Error("failed:->", plainTxt, string(deStr))
	}
}

func TestNodeId(t *testing.T) {
	key, err := GenerateKey(password)
	if err != nil {
		t.Error(err)
	}

	printKey(t, key)
	nid := key.ToNodeId()
	t.Log("nid:", nid)

	pub := ToPubKey(nid)
	t.Log("pub", pub)

	if !bytes.Equal(pub, key.PubKey[:]) {
		t.Errorf("node id(%s)and public key (%s)"+
			" is not equal with original (%s)", nid, pub, key.PubKey)
	}
}

func TestGenerateKeyPair(t *testing.T) {
	key, err := GenerateKey(password)
	if err != nil {
		t.Error(err)
	}

	printKey(t, key)

	pub := make([]byte, len(key.PubKey))
	copy(pub, key.PubKey)

	pri := make([]byte, len(key.priKey))
	copy(pri, key.priKey)

	acc := &Account{
		NodeId: key.ToNodeId(),
		Key: &Key{
			LockedKey: key.LockedKey,
		},
	}

	acc.UnlockAcc(password)

	if !bytes.Equal(pub, acc.Key.PubKey) {
		t.Error("public key is not equal", pub, key.PubKey)
	}
	//if !bytes.Equal(edpri, key.eDPriKey){
	if !bytes.Equal(pri, acc.Key.priKey) {
		t.Error("pri key is not equal", pri, key.priKey)
	}
}

func printKey(t *testing.T, key *Key) {
	t.Log("priKey:", key.priKey, "len", len(key.priKey))
	t.Log("pubKey:", key.PubKey, "len", len(key.PubKey))
	t.Log("LockedKey:", key.LockedKey, "len", len(key.LockedKey))
}
