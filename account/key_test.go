package account

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/scrypt"
	"io"
	"testing"
)

var password = "12345678"
var salt = []byte{0xc8, 0x28, 0xf2, 0x58, 0xa7, 0x6a, 0xad, 0x7b}
var plainTxt = "hello word"

func newCFBEncrypter(t *testing.T, key, plaintext []byte) []byte {
	// Load your secret key from a safe place and reuse it across multiple
	// NewCipher calls. (Obviously don't use this example key for anything
	// real.) If you want to convert a passphrase to a key, use a suitable
	// package like bcrypt or scrypt.
	//key, _ := hex.DecodeString("6368616e676520746869732070617373")
	//plaintext := []byte("some plaintext")
	//t.Log(key)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	// It's important to remember that ciphertexts must be authenticated
	// (i.e. by using crypto/hmac) as well as being encrypted in order to
	// be secure.
	t.Logf("%x\n", ciphertext)
	return ciphertext
}

func newCFBDecrypter(t *testing.T, key, ciphertext []byte) string {
	// Load your secret key from a safe place and reuse it across multiple
	// NewCipher calls. (Obviously don't use this example key for anything
	// real.) If you want to convert a passphrase to a key, use a suitable
	// package like bcrypt or scrypt.
	//key, _ := hex.DecodeString("6368616e676520746869732070617373")
	//ciphertext, _ := hex.DecodeString("7dd015f06bec7f1b8f6559dad89f4131da62261786845100056b353194ad")

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		panic("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(ciphertext, ciphertext)
	s := fmt.Sprintf("%s", ciphertext)
	t.Log(s)
	return s
	// Output: some plaintext
}

func TestAesEncrypt1(t *testing.T) {
	aesKey, err := getAESKey(salt, password)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(aesKey)

	ciptxt := newCFBEncrypter(t, aesKey, []byte(plainTxt))
	h_2 := newCFBDecrypter(t, aesKey, ciptxt)
	if plainTxt != h_2 {
		t.Error("basic aes failed", plainTxt, h_2)
	}
}

func TestAesEncrypt(t *testing.T) {
	//h:="hello word"
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
	t.Log(deStr)

	if plainTxt != string(deStr) {
		t.Error("failed:->", plainTxt, string(deStr))
	}
}

func TestScrypt(t *testing.T) {
	sw := "some password"
	dk, err := scrypt.Key([]byte(sw), salt, 1<<15, 8, 1, 32)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(base64.StdEncoding.EncodeToString(dk))
}

func TestNodeId(t *testing.T) {
	key, err := GenerateKey(password)
	if err != nil {
		t.Error(err)
	}
	t.Log("pubKey", key.pubKey)
	t.Log("priKey", key.priKey)
	t.Log("eDPriKey", key.eDPriKey)
	t.Log("eDPubKey", key.eDPubKey)
	t.Log("LockedKey", key.LockedKey)

	nid := key.ToNodeId()
	t.Log("nid", nid)
	pub := ToPubKey(nid)
	t.Log("pub", pub)
	if !bytes.Equal(pub, key.pubKey[:]) {
		t.Errorf("node id(%s)and public key (%s)"+
			" is not equal with original (%s)", nid, pub, key.pubKey)
	}
}

func TestGenerateKeyPair(t *testing.T) {
	key, err := GenerateKey(password)
	if err != nil {
		t.Error(err)
	}
	edpub, edpri, pub, pri := key.eDPubKey, key.eDPriKey, key.pubKey, key.priKey
	//key.priKey = [KeyLen]byte{0}
	//aa := [64]byte{0}
	////
	//copy(key.eDPriKey, aa[:64])
	//key.eDPriKey = make([]byte, 64)

	printKey(t, key)

	aesKey, err := getAESKey(key.pubKey[:kp.S], password) //scrypt.Key([]byte(password), k.PubKey[:kp.S], kp.N, kp.R, kp.P, kp.L)
	if err != nil {
		t.Error("error to generate aes key:->", err)
	}
	t.Log("aesKey:", aesKey)
	raw, err := Decrypt(aesKey, key.LockedKey)
	if err != nil {
		t.Error("error to unlock raw private key:->", err)
	}

	t.Log("raw:", raw)

	key.curve25519KeyPair = curveKey(raw)
	key.ed25519KeyPair = edKey(raw)

	printKey(t, key)

	//if !bytes.Equal(edpub, key.eDPubKey) {
	if !bytes.Equal(edpub, key.eDPubKey) {
		t.Error("ed public key is not equal", edpub, key.eDPubKey)
	}
	//if !bytes.Equal(edpri, key.eDPriKey){
	if !bytes.Equal(edpri, key.eDPriKey) {
		t.Error("ed pri key is not equal", edpri, key.eDPriKey)
	}

	if pub != key.pubKey {
		t.Error("public key is not equal", pub, key.pubKey)
	}
	if pri != key.priKey {
		t.Error("pri key is not equal", pri, key.priKey)
	}
}

func printKey(t *testing.T, key *Key) {
	t.Log("priKey:", key.priKey, "len", len(key.priKey))
	t.Log("pubKey:", key.pubKey, "len", len(key.pubKey))
	t.Log("eDPriKey:", key.eDPriKey, "len", len(key.eDPriKey))
	t.Log("eDPubKey:", key.eDPubKey, "len", len(key.eDPubKey))
	t.Log("LockedKey:", key.LockedKey, "len", len(key.LockedKey))
}
