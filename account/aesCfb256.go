package account

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

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
		return nil, fmt.Errorf("cipher text too short")
	}

	iv := cipherTxt[:aes.BlockSize]
	plainTxt := make([]byte, len(cipherTxt)-aes.BlockSize)

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(plainTxt, cipherTxt[aes.BlockSize:])

	return plainTxt, nil
}
