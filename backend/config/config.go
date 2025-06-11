package config

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

const (
	configFile    = "dbconfig.json"
	encConfigFile = "config.json.enc"
)

type DBConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
}

var (
	SetupUser     string
	SetupPassword string
)

func LoadEncryptedCredentials(key []byte) (string, string, error) {
	if _, err := os.Stat(encConfigFile); os.IsNotExist(err) {
		return "", "", errors.New("archivo encriptado no encontrado")
	}

	encryptedData, err := ioutil.ReadFile(encConfigFile)
	if err != nil {
		return "", "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", "", err
	}

	if len(encryptedData) < 12 {
		return "", "", errors.New("datos cifrados incompletos")
	}

	nonce := encryptedData[:12]
	ciphertext := encryptedData[12:]

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", err
	}

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", "", err
	}

	var creds struct {
		User string `json:"user"`
		Pass string `json:"pass"`
	}
	if err := json.Unmarshal(plaintext, &creds); err != nil {
		return "", "", err
	}

	return creds.User, creds.Pass, nil
}

func ConfigExists() bool {
	_, err := os.Stat(configFile)
	return !os.IsNotExist(err)
}

func SaveConfig(cfg DBConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(configFile, data, 0644)
}

func LoadConfig() (DBConfig, error) {
	var cfg DBConfig
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return cfg, err
	}
	err = json.Unmarshal(data, &cfg)
	return cfg, err
}