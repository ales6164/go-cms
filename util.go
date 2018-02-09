package api

import (
	"reflect"
	"golang.org/x/crypto/bcrypt"
)

func getType(class interface{}) reflect.Type {
	if t := reflect.TypeOf(class); t.Kind() == reflect.Ptr {
		return t.Elem()
	} else {
		return t
	}
}

func decrypt(hash []byte, password []byte) error {
	defer clear(password)
	return bcrypt.CompareHashAndPassword(hash, password)
}

func crypt(password []byte) ([]byte, error) {
	defer clear(password)
	return bcrypt.GenerateFromPassword(password, 13)
}

func clear(b []byte) {
	for i := 0; i < len(b); i++ {
		b[i] = 0
	}
}
