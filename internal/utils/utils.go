package utils

import (
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(plainTxt string) (h string, err error) {
	var hashed []byte
	hashed, err = bcrypt.GenerateFromPassword([]byte(plainTxt), bcrypt.DefaultCost)
	if err != nil {
		return
	}
	h = string(hashed)
	return
}

func ComparePassword(hashedPwd, plainText string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(plainText)) == nil
}
