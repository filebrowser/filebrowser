package users

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"log"

	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"

	fbErrors "github.com/filebrowser/filebrowser/v2/errors"
)

// HashPwd hashes a password.
func HashPwd(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPwd checks if a password is correct.
func CheckPwd(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// returns cipher text and nonce in base64
func EncryptSymmetric(encryptionKey, secret []byte) (string, string, error) {
	if len(encryptionKey) != 32 {
		log.Printf("%s (key=\"%s\")", fbErrors.ErrInvalidEncryptionKey.Error(), string(encryptionKey))
		return "", "", fbErrors.ErrInvalidEncryptionKey
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", "", err
	}

	cipherText := gcm.Seal(nil, nonce, secret, nil)

	return base64.StdEncoding.EncodeToString(cipherText), base64.StdEncoding.EncodeToString(nonce), nil
}

func DecryptSymmetric(encryptionKey []byte, cipherTextB64, nonceB64 string) (string, error) {
	if len(encryptionKey) != 32 {
		log.Printf("%s (key=\"%s\")", fbErrors.ErrInvalidEncryptionKey.Error(), string(encryptionKey))
		return "", fbErrors.ErrInvalidEncryptionKey
	}

	cipherText, err := base64.StdEncoding.DecodeString(cipherTextB64)
	if err != nil {
		return "", err
	}

	nonce, err := base64.StdEncoding.DecodeString(nonceB64)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	secret, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", err
	}

	return string(secret), nil
}

// Decrypt the secret and validate the code
func CheckTOTP(totpEncryptionKey []byte, encryptedSecretB64, nonceB64, code string) (bool, error) {
	if len(totpEncryptionKey) != 32 {
		log.Printf("%s (key=\"%s\")", fbErrors.ErrInvalidEncryptionKey.Error(), string(totpEncryptionKey))
		return false, fbErrors.ErrInvalidEncryptionKey
	}

	secret, err := DecryptSymmetric(totpEncryptionKey, encryptedSecretB64, nonceB64)
	if err != nil {
		return false, err
	}

	return totp.Validate(code, secret), nil
}
