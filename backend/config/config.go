package config

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
    ConfigFile    = "dbconfig.json" // Cambiado a mayúscula para exportar
    encConfigFile = "config.json.enc"
)

// User struct definition for GORM
type User struct {
    ID        uint      `gorm:"primaryKey"`
    Email     string    `gorm:"unique;not null"`
    Password  string    `gorm:"not null"`
    CreatedAt time.Time // Cambiado a time.Time
    UpdatedAt time.Time // Cambiado a time.Time
}

type DBConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
}
type SetupStatus struct {
	DBConfigured    bool `json:"db_configured"`
	DBTablesCreated bool `json:"db_tables_created"`
	AdminCreated    bool `json:"admin_created"`
}

var (
	SetupUser          string
	SetupPassword      string
	currentConfig      DBConfig
	CurrentSetupStatus SetupStatus
	configMutex        sync.RWMutex
)

// GetDSN returns the connection string for PostgreSQL
func (c DBConfig) GetDSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		c.Host,
		c.User,
		c.Password,
		c.DBName,
		c.Port)
}

// Agrega esta función en config/config.go
func IsSetupComplete() bool {
	return CurrentSetupStatus.DBConfigured &&
		CurrentSetupStatus.DBTablesCreated &&
		CurrentSetupStatus.AdminCreated
}

// inizializacion de las tablas de la base de datos
func InitializeDB(db *gorm.DB) error {
	// Crear tablas según el SQL proporcionado
	err := db.Exec(`
		CREATE TABLE IF NOT EXISTS smtp_config (
			id SERIAL PRIMARY KEY,
			host VARCHAR(100) NOT NULL,
			port INTEGER NOT NULL,
			username VARCHAR(100) NOT NULL,
			password VARCHAR(100) NOT NULL,
			from_email VARCHAR(100) NOT NULL,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		
		CREATE INDEX IF NOT EXISTS idx_smtp_config_active ON smtp_config(is_active);
		
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(100) UNIQUE NOT NULL,
			password VARCHAR(100) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		
		CREATE TABLE IF NOT EXISTS reset_codes (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			code VARCHAR(10) NOT NULL,
			expiration_time TIMESTAMP WITH TIME ZONE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
		
		CREATE INDEX IF NOT EXISTS idx_reset_codes_expiration ON reset_codes(expiration_time);
	`).Error

	if err != nil {
		return fmt.Errorf("error creating tables: %v", err)
	}

	CurrentSetupStatus.DBTablesCreated = true
	return nil
}

// creacion de un usuario administrador
func CreateAdminUser(db *gorm.DB, email, password string) (bool, error) {
    // Verificar si el usuario ya existe
    var existingUser User
    result := db.Where("email = ?", email).First(&existingUser)
    
    if result.Error == nil {
        // Usuario ya existe - no es un error, solo marcamos como completado
        CurrentSetupStatus.AdminCreated = true
        return false, nil
    }

    // Si no existe, crear nuevo usuario
    admin := User{
        Email:    email,
        Password: password,
    }

    if err := db.Create(&admin).Error; err != nil {
        return false, fmt.Errorf("error creating admin user: %v", err)
    }

    CurrentSetupStatus.AdminCreated = true
    return true, nil
}

// Cargar configuración de DB desde dbconfig.json
// Modificar la función LoadDBConfig para no bloquear la reconfiguración
func LoadDBConfig() (DBConfig, error) {
	configMutex.Lock()
	defer configMutex.Unlock()

	// Leer archivo
	data, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		if os.IsNotExist(err) {
			return DBConfig{}, fmt.Errorf("archivo de configuración no existe")
		}
		return DBConfig{}, fmt.Errorf("error leyendo archivo: %v", err)
	}

	// Verificar si el archivo está vacío
	if len(data) == 0 {
		return DBConfig{}, fmt.Errorf("archivo de configuración vacío")
	}

	var cfg DBConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DBConfig{}, fmt.Errorf("error parseando configuración: %v", err)
	}

	// Verificar si la configuración está vacía
	if cfg == (DBConfig{}) {
		return DBConfig{}, fmt.Errorf("configuración inválida: todos los campos vacíos")
	}

	currentConfig = cfg
	CurrentSetupStatus.DBConfigured = true
	return cfg, nil
}

// GetCurrentConfig devuelve la configuración actual cargada
func GetCurrentConfig() DBConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return currentConfig
}

/*
// Probar conexión a la base de datos
func (c *DBConfig) TestConnection() (map[string]interface{}, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		c.Host, c.User, c.Password, c.DBName, c.Port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("error al conectar: %v", err)
	}

	// Obtener información de la base de datos
	var version string
	db.Raw("SELECT version()").Scan(&version)

	var tableCount int64
	db.Raw("SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&tableCount)

	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}

	return map[string]interface{}{
		"db_type":    "postgres",
		"status":     "success",
		"version":    version,
		"tableCount": tableCount,
	}, nil
}
*/
// Actualizar configuración de DB
// UpdateDBConfig actualiza la configuración
func UpdateDBConfig(newConfig DBConfig) error {
	configMutex.Lock()
	defer configMutex.Unlock()

	data, err := json.MarshalIndent(newConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("error serializando configuración: %v", err)
	}

	if err := ioutil.WriteFile(ConfigFile, data, 0644); err != nil {
		return fmt.Errorf("error guardando configuración: %v", err)
	}

	currentConfig = newConfig
	CurrentSetupStatus.DBConfigured = true
	return nil
}

// Obtener configuración actual (sin contraseña)
func GetDBConfig() map[string]interface{} {
	return map[string]interface{}{
		"host":   currentConfig.Host,
		"port":   currentConfig.Port,
		"user":   currentConfig.User,
		"dbname": currentConfig.DBName,
	}
}
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
	_, err := os.Stat(ConfigFile)
	return !os.IsNotExist(err)
}

func SaveConfig(cfg DBConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(ConfigFile, data, 0644)
}

func LoadConfig() (DBConfig, error) {
	var cfg DBConfig
	data, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		return cfg, err
	}
	err = json.Unmarshal(data, &cfg)
	return cfg, err
}
func GetConfigPath() string {
	// Obtiene la ruta absoluta del archivo
	if absPath, err := filepath.Abs(ConfigFile); err == nil {
		return absPath
	}
	return ConfigFile
}

// funcion para resetear la configuración de la base de datos
// ResetConfig elimina el archivo de configuración y resetea el estado
/*func ResetConfig() error {
configPath := GetConfigPath()
fmt.Printf("Intentando eliminar: %s\n", configPath)

// Verifica permisos primero
if _, err := os.Stat(configPath); err == nil {
	// Intenta eliminar el archivo
	if err := os.Remove(configPath); err != nil {
		// Intenta forzar eliminación
		fmt.Printf("Error al eliminar (intentando forzar): %v\n", err)
		if err := forceRemove(configPath); err != nil {
			return fmt.Errorf("no se pudo eliminar el archivo: %v", err)
		}
	}
	fmt.Println("Archivo eliminado exitosamente")
}


// Resetear estado
currentConfig = DBConfig{}
CurrentSetupStatus = SetupStatus{
	DBConfigured:    false,
	DBTablesCreated: false,
	AdminCreated:    false,
}*/
// ResetConfig elimina el archivo y resetea el estado
func ResetConfig() error {
	configMutex.Lock()
	defer configMutex.Unlock()

	// 1. Eliminar físicamente el archivo
	if _, err := os.Stat(ConfigFile); err == nil {
		if err := os.Remove(ConfigFile); err != nil {
			return fmt.Errorf("no se pudo eliminar el archivo: %v", err)
		}
		fmt.Println("🗑️ Archivo dbconfig.json eliminado físicamente")
	}

	// 2. Resetear configuración en memoria
	currentConfig = DBConfig{}

	// 3. Resetear estado
	CurrentSetupStatus = SetupStatus{
		DBConfigured:    false,
		DBTablesCreated: false,
		AdminCreated:    false,
	}

	return nil
}

func forceRemove(path string) error {
	// Primero intenta cambiar permisos
	if err := os.Chmod(path, 0666); err != nil {
		return err
	}
	// Luego intenta eliminar
	return os.Remove(path)
}

// RefreshConfigState actualiza el estado basado en el archivo físico
func RefreshConfigState() {
	configMutex.Lock()
	defer configMutex.Unlock()

	// Verificar si el archivo existe físicamente
	if _, err := os.Stat(ConfigFile); os.IsNotExist(err) {
		CurrentSetupStatus.DBConfigured = false
		currentConfig = DBConfig{}
		fmt.Println("🔄 Estado actualizado: archivo no existe")
		return
	}

	// Si el archivo existe, cargarlo y verificar conexión
	cfg, err := loadConfigFromFile()
	if err != nil {
		CurrentSetupStatus.DBConfigured = false
		currentConfig = DBConfig{}
		fmt.Printf("🔄 Estado actualizado: archivo inválido (%v)\n", err)
		return
	}

	// Verificar conexión real
	if _, err := cfg.TestConnection(); err != nil {
		CurrentSetupStatus.DBConfigured = false
		currentConfig = DBConfig{}
		fmt.Printf("🔄 Estado actualizado: conexión fallida (%v)\n", err)
		return
	}

	// Todo está correcto
	currentConfig = cfg
	CurrentSetupStatus.DBConfigured = true
	fmt.Println("🔄 Estado actualizado: configuración válida y conexión exitosa")
}

// GetCurrentConfig devuelve la configuración actual
/*func GetCurrentConfig() DBConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return currentConfig
}*/

// ===== FUNCIONES INTERNAS =====
func loadConfigFromFile() (DBConfig, error) {
	var cfg DBConfig

	// Verificar si el archivo existe
	if _, err := os.Stat(ConfigFile); os.IsNotExist(err) {
		return cfg, fmt.Errorf("archivo no existe")
	}

	data, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		return cfg, fmt.Errorf("error leyendo archivo: %v", err)
	}

	// Verificar si el archivo está vacío
	if len(data) == 0 {
		return cfg, fmt.Errorf("archivo vacío")
	}

	// Verificar si el archivo contiene solo {} o espacios
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "{}" || trimmed == "" {
		return cfg, fmt.Errorf("archivo vacío")
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("error parseando JSON: %v", err)
	}

	// Verificar si la configuración está vacía
	if cfg.Host == "" && cfg.User == "" && cfg.DBName == "" {
		return cfg, fmt.Errorf("configuración vacía")
	}

	return cfg, nil
}

func (c *DBConfig) TestConnection() (map[string]interface{}, error) {
	// Verificar si la configuración es válida
	if c.Host == "" || c.User == "" || c.DBName == "" {
		return nil, fmt.Errorf("configuración incompleta: campos requeridos vacíos")
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		c.Host, c.User, c.Password, c.DBName, c.Port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("error al conectar: %v", err)
	}

	// Obtener información de la base de datos
	var version string
	db.Raw("SELECT version()").Scan(&version)

	var tableCount int64
	db.Raw("SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&tableCount)

	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}

	return map[string]interface{}{
		"db_type":    "postgres",
		"status":     "success",
		"version":    version,
		"tableCount": tableCount,
	}, nil
}
