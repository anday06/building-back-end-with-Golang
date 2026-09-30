package password

import "golang.org/x/crypto/bcrypt"

func Hash(value string) (string, error) {
	result, err := bcrypt.GenerateFromPassword([]byte(value), bcrypt.DefaultCost)
	return string(result), err
}

func Compare(hash, value string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(value)) == nil
}
