package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	
)

type Credentials struct {
	User string `json:"user"`
	Pass string `json:"pass"`
}

func encrypt(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)
	return append(nonce, ciphertext...), nil
}

func main() {
	// Usando tu key de prueba
	key := []byte("12345678901234567890123456789012")

	creds := Credentials{
		User: "admin",
		Pass: "shinon56",
	}

	jsonData, err := json.Marshal(creds)
	if err != nil {
		panic(err)
	}

	encryptedData, err := encrypt(jsonData, key)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile("config.json.enc", encryptedData, 0600)
	if err != nil {
		panic(err)
	}

	fmt.Println("✅ Archivo config.json.enc generado correctamente")
	fmt.Println("🔑 Key usada:", string(key))
	fmt.Println("👤 Usuario:", creds.User)
	fmt.Println("🔒 Contraseña:", creds.Pass)
}