package utils

import (
	"crypto/rand"
	"log"
	"os"
)

func CreateHmacKey() []byte {
	file, err := os.ReadFile("private.key")
	if err != nil {
		var key = make([]byte, 64)
		rand.Reader.Read(key)
		err = os.WriteFile("private.key", key, 0600)
		if err != nil {
			log.Fatal(err)
		}
		return key
	} else {
		return file
	}
}
