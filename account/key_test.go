package account

import (
	"bytes"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/ed25519"
	"testing"
)

var password = "12345678"
var salt = []byte{0xc8, 0x28, 0xf2, 0x58, 0xa7, 0x6a, 0xad, 0x7b}
var plainTxt = "Ed25519 is a public-key signature system with several attractive features:"

func TestAesEncrypt(t *testing.T) {
	aesKey, err := AESKey(salt, password)
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
	t.Log("nid:->", nid)

	pub := ToPubKey(nid)
	t.Log("pub:->", pub)

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

	pri := make([]byte, len(key.PriKey))
	copy(pri, key.PriKey)

	acc := &Account{
		Address: key.ToNodeId(),
		Key: &Key{
			LockedKey: key.LockedKey,
		},
	}

	acc.UnlockAcc(password)

	if !bytes.Equal(pub, acc.Key.PubKey) {
		t.Error("public key is not equal", pub, key.PubKey)
	}
	//if !bytes.Equal(edpri, key.eDPriKey){
	if !bytes.Equal(pri, acc.Key.PriKey) {
		t.Error("pri key is not equal", pri, key.PriKey)
	}
}

func printKey(t *testing.T, key *Key) {
	t.Log("PriKey::->", key.PriKey, "len", len(key.PriKey))
	t.Log("pubKey:->", key.PubKey, "len", len(key.PubKey))
	t.Log("LockedKey:->", key.LockedKey, "len", len(key.LockedKey))
}

func TestSign(t *testing.T) {
	key, err := GenerateKey(password)
	if err != nil {
		t.Error(err)
	}

	sign := ed25519.Sign(key.PriKey, []byte(plainTxt))

	t.Log("sing:->", sign)

	tt := ed25519.Verify(key.PubKey, []byte(plainTxt), sign)
	if !tt {
		t.Error("verify failed:->")
	}
	tt2 := ed25519.Verify(key.PubKey, []byte("some failed message"), sign)
	if tt2 {
		t.Error("verify failed:->")
	}
}
func TestConvert(t *testing.T) {
	key, _ := GenerateKey(password)

	var pubCC1 [32]byte
	var publicKeyBytes [32]byte
	copy(publicKeyBytes[:], key.PubKey)
	tt := PublicKeyToCurve25519(&pubCC1, &publicKeyBytes)
	t.Log("cc pub key :->", pubCC1)

	if !tt {
		t.Error("convert pub key to curve pub failed")
	}

	var priCC1 [32]byte
	var priKeyBytes [64]byte
	copy(priKeyBytes[:], key.PriKey)
	PrivateKeyToCurve25519(&priCC1, &priKeyBytes)
	t.Log("cc private key :->", priCC1)

	var pubCC2 [32]byte
	curve25519.ScalarBaseMult(&pubCC2, &priCC1)
	t.Log("cc2 pub key :->", pubCC1)

	if pubCC2 != pubCC1 {
		t.Error("convert ed curve to curve encrypt failed")
	}

}
func TestCrypt(t *testing.T) {
	key1, _ := GenerateKey(password)
	printKey(t, key1)
	key2, _ := GenerateKey(password)
	printKey(t, key2)

	var pubCC1 [32]byte
	var publicKeyBytes [32]byte
	copy(publicKeyBytes[:], key1.PubKey)
	PublicKeyToCurve25519(&pubCC1, &publicKeyBytes)
	var priCC1 [32]byte
	var privateKeyBytes [64]byte
	copy(privateKeyBytes[:], key1.PriKey)
	PrivateKeyToCurve25519(&priCC1, &privateKeyBytes)

	var pubCC2 [32]byte
	copy(publicKeyBytes[:], key2.PubKey)
	PublicKeyToCurve25519(&pubCC2, &publicKeyBytes)
	var priCC2 [32]byte
	copy(privateKeyBytes[:], key2.PriKey)
	PrivateKeyToCurve25519(&priCC2, &privateKeyBytes)

	var aesKey1 [32]byte
	curve25519.ScalarMult(&aesKey1, &priCC1, &pubCC2)
	t.Log("aeskey1:->", aesKey1)
	t.Log("priCC1:->", priCC1)
	t.Log("pubCC2:->", pubCC2)

	var aesKey2 [32]byte
	curve25519.ScalarMult(&aesKey2, &priCC2, &pubCC1)
	t.Log("aesKey2:->", aesKey2)
	t.Log("priCC2:->", priCC2)
	t.Log("pubCC1:->", pubCC1)

	if aesKey1 != aesKey2 {
		t.Error("send aes key failed")
	}
}
