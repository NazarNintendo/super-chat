package database

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	_ "fmt"
	log "gitlab.starlink.ua/high-school-prod/chat/logger"
	"io"
	_ "io/ioutil"
	mRand "math/rand"
	_ "os"
	"time"
)

// key - used in encrypting / decrypting page tokens
var key = randHex(32)

// randHex - generates a random Hex value
func randHex(length int) string {
	mRand.Seed(time.Now().UnixNano())
	var letters = []rune("0123456789ABCDEF")
	b := make([]rune, length)
	for i := range b {
		b[i] = letters[mRand.Intn(len(letters))]
	}
	return string(b)
}

func encrypt(stringToEncrypt string) (encryptedString string) {
	if stringToEncrypt == "" {
		return ""
	}
	key, _ := hex.DecodeString(key)
	plaintext := []byte(stringToEncrypt)
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Logger.Fatal(err)
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		log.Logger.Fatal(err)
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		log.Logger.Fatal(err)
	}
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return fmt.Sprintf("%x", ciphertext)
}

func decrypt(encryptedString string) (decryptedString string) {
	if encryptedString == "" {
		return ""
	}
	key, _ := hex.DecodeString(key)
	enc, _ := hex.DecodeString(encryptedString)
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Logger.Fatal(err)
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		log.Logger.Fatal(err)
	}
	nonceSize := aesGCM.NonceSize()
	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		log.Logger.Fatal(err)
	}
	return fmt.Sprintf("%s", plaintext)
}
