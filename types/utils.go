package types

import (
	"crypto/rand"

	"golang.org/x/crypto/bcrypt"
)

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	return b, err
}

func isBinary(content string) bool {
	for _, b := range content {
		// 65533 is the unknown char
		// 8 and below are control chars (e.g. backspace, null, eof, etc)
		if b <= 8 || b == 65533 {
			return true
		}
	}
	return false
}

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
