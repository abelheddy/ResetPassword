package main

import (
	"crypto/rand"
	"crypto/tls"
	"database/sql"
	"fmt"
	"log"
	"math/big"

	"net"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// Configuración de la aplicación
type AppConfig struct {
	DBHost     string `json:"db_host"`
	DBPort     string `json:"db_port"`
	DBUser     string `json:"db_user"`
	DBPassword string `json:"db_password"`
	DBName     string `json:"db_name"`
	ServerPort string `json:"server_port"`
}

// Configuración SMTP
type SMTPConfig struct {
	ID        int       `json:"id,omitempty"`
	Host      string    `json:"host" binding:"required"`
	Port      int       `json:"port" binding:"required"`
	Username  string    `json:"username" binding:"required"`
	Password  string    `json:"password" binding:"required"`
	FromEmail string    `json:"from_email" binding:"required"`
	IsActive  bool      `json:"is_active,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// Estructura para solicitud de código
type RequestCode struct {
	Email string `json:"email"`
}

var db *sql.DB

func main() {
	// Cargar configuración
	cfg := loadConfig()

	// Conectar a la base de datos
	var err error
	db, err = connectDB(cfg)
	if err != nil {
		log.Fatalf("Error al conectar a la base de datos: %v", err)
	}
	defer db.Close()

	// Crear tablas si no existen
	if err := createTables(); err != nil {
		log.Fatalf("Error al crear tablas: %v", err)
	}

	// Configurar router
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // URL de tu frontend
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Rutas para la configuración SMTP
	admin := router.Group("/admin")
	{
		admin.GET("/smtp-config", getSMTPConfigHandler)
		admin.POST("/smtp-config", createSMTPConfigHandler)
		admin.PUT("/smtp-config", updateSMTPConfigHandler)
		admin.DELETE("/smtp-config", deleteSMTPConfigHandler)
		admin.POST("/test-smtp", testSMTPConnectionHandler)
	}

	// Rutas para recuperación de contraseña
	router.POST("/send-code", sendCode)
	router.POST("/verify-code", verifyCode)
	router.POST("/reset-password", resetPassword)

	// Iniciar servidor
	log.Printf("Servidor iniciado en el puerto %s", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}

func loadConfig() AppConfig {
	return AppConfig{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "root"),
		DBName:     getEnv("DB_NAME", "db_reset"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func connectDB(cfg AppConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error al conectar a la base de datos: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error al verificar conexión a la BD: %v", err)
	}

	return db, nil
}

func createTables() error {
	// Crear tabla de configuración SMTP
	smtpTable := `
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
	`

	// Crear tabla de códigos de recuperación
	resetCodesTable := `
	CREATE TABLE IF NOT EXISTS reset_codes (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL,
		code VARCHAR(10) NOT NULL,
		expiration_time TIMESTAMP WITH TIME ZONE NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		FOREIGN KEY (user_id) REFERENCES users(id)
	);
	CREATE INDEX IF NOT EXISTS idx_reset_codes_expiration ON reset_codes(expiration_time);
	`

	// Crear tabla de usuarios si no existe
	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		email VARCHAR(100) UNIQUE NOT NULL,
		password VARCHAR(100) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);
	`

	// Ejecutar todas las consultas
	if _, err := db.Exec(smtpTable); err != nil {
		return fmt.Errorf("error al crear tabla smtp_config: %v", err)
	}
	if _, err := db.Exec(resetCodesTable); err != nil {
		return fmt.Errorf("error al crear tabla reset_codes: %v", err)
	}
	if _, err := db.Exec(usersTable); err != nil {
		return fmt.Errorf("error al crear tabla users: %v", err)
	}

	return nil
}

func sendEmail(to, subject, body string) error {
	// Obtener la configuración SMTP activa
	var config SMTPConfig
	err := db.QueryRow(`
        SELECT host, port, username, password, from_email 
        FROM smtp_config 
        WHERE is_active = TRUE LIMIT 1`).Scan(
		&config.Host,
		&config.Port,
		&config.Username,
		&config.Password,
		&config.FromEmail,
	)

	if err != nil {
		return fmt.Errorf("no se pudo obtener la configuración SMTP: %v", err)
	}

	// Construir el mensaje
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		config.FromEmail, to, subject, body)

	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)

	// Opción 1: Puerto 587 con STARTTLS (recomendado)
	if config.Port == 587 {
		client, err := smtp.Dial(fmt.Sprintf("%s:%d", config.Host, config.Port))
		if err != nil {
			return fmt.Errorf("error al conectar al servidor SMTP: %v", err)
		}
		defer client.Close()

		if ok, _ := client.Extension("STARTTLS"); ok {
			tlsConfig := &tls.Config{
				ServerName:         config.Host,
				InsecureSkipVerify: true, // Solo para desarrollo/testing
			}
			if err = client.StartTLS(tlsConfig); err != nil {
				return fmt.Errorf("error al iniciar TLS: %v", err)
			}
		}

		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("error de autenticación: %v", err)
		}

		if err = client.Mail(config.FromEmail); err != nil {
			return fmt.Errorf("error al establecer remitente: %v", err)
		}

		if err = client.Rcpt(to); err != nil {
			return fmt.Errorf("error al establecer destinatario: %v", err)
		}

		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("error al preparar cuerpo del mensaje: %v", err)
		}
		defer w.Close()

		_, err = w.Write([]byte(msg))
		if err != nil {
			return fmt.Errorf("error al escribir mensaje: %v", err)
		}

		return nil
	}

	// Opción 2: Puerto 465 con SSL/TLS
	if config.Port == 465 {
		tlsConfig := &tls.Config{
			ServerName:         config.Host,
			InsecureSkipVerify: true, // Solo para desarrollo/testing
		}

		conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port), tlsConfig)
		if err != nil {
			return fmt.Errorf("error al conectar al servidor SMTP: %v", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, config.Host)
		if err != nil {
			return fmt.Errorf("error al crear cliente SMTP: %v", err)
		}
		defer client.Close()

		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("error de autenticación: %v", err)
		}

		if err = client.Mail(config.FromEmail); err != nil {
			return fmt.Errorf("error al establecer remitente: %v", err)
		}

		if err = client.Rcpt(to); err != nil {
			return fmt.Errorf("error al establecer destinatario: %v", err)
		}

		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("error al preparar cuerpo del mensaje: %v", err)
		}
		defer w.Close()

		_, err = w.Write([]byte(msg))
		if err != nil {
			return fmt.Errorf("error al escribir mensaje: %v", err)
		}

		return nil
	}

	return fmt.Errorf("puerto SMTP no soportado: %d", config.Port)
}

func sendCode(c *gin.Context) {
	var request RequestCode

	// Parsear el JSON de entrada
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		log.Println("Error al parsear el JSON:", err)
		return
	}

	// Normalizar el email (minúsculas y sin espacios)
	request.Email = strings.TrimSpace(strings.ToLower(request.Email))
	log.Println("Buscando email:", request.Email)

	var userId int
	// Buscar el usuario en la base de datos
	err := db.QueryRow("SELECT id FROM users WHERE email = $1", request.Email).Scan(&userId)
	if err != nil {
		if err == sql.ErrNoRows {
			// El correo no existe en la BD
			log.Println("Correo no encontrado en la BD:", request.Email)
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Correo no encontrado",
				"details": "El correo proporcionado no está registrado",
			})
		} else {
			// Error de base de datos
			log.Println("Error al verificar correo:", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Error al verificar el correo",
				"details": err.Error(),
			})
		}
		return
	}

	// Generar código aleatorio de 8 dígitos usando crypto/rand (más seguro)
	n, err := rand.Int(rand.Reader, big.NewInt(100000000))
	if err != nil {
		log.Println("Error al generar código aleatorio:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al generar código"})
		return
	}
	code := fmt.Sprintf("%08d", n)

	expirationTime := time.Now().Add(5 * time.Minute) // Código válido por 5 minutos

	// Guardar el código en la base de datos
	_, err = db.Exec("INSERT INTO reset_codes (user_id, code, expiration_time) VALUES ($1, $2, $3)",
		userId, code, expirationTime)
	if err != nil {
		log.Println("Error al insertar el código en la base de datos:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al guardar el código"})
		return
	}

	// Preparar y enviar el correo
	emailBody := "Tu código de restablecimiento de contraseña es: " + code
	if err := sendEmail(request.Email, "Restablecimiento de contraseña", emailBody); err != nil {
		log.Println("Error al enviar el correo:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al enviar el correo"})
		return
	}

	// Respuesta exitosa
	c.JSON(http.StatusOK, gin.H{
		"message": "Código enviado correctamente",
		"email":   request.Email,
	})
}

func verifyCode(c *gin.Context) {
	var request struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}

	// Parsear el JSON de entrada
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var expirationTime time.Time
	var validCode string
	var userId int
	// Buscar el código en la base de datos (solo códigos no expirados)
	err := db.QueryRow(`
		SELECT user_id, code, expiration_time 
		FROM reset_codes 
		WHERE code = $1 AND expiration_time > NOW()`,
		request.Code).Scan(&userId, &validCode, &expirationTime)

	if err != nil {
		if err == sql.ErrNoRows {
			// Código no encontrado o expirado
			c.JSON(http.StatusNotFound, gin.H{"error": "Código no válido o expirado"})
		} else {
			// Error de base de datos
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar el código"})
		}
		return
	}

	// Verificar que el código corresponde al email
	var dbEmail string
	err = db.QueryRow("SELECT email FROM users WHERE id = $1", userId).Scan(&dbEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar el correo del usuario"})
		return
	}

	// Comparación case-insensitive de emails
	if strings.ToLower(dbEmail) != strings.ToLower(request.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Correo y código no coinciden"})
		return
	}

	// Código verificado correctamente
	c.JSON(http.StatusOK, gin.H{"message": "Código verificado correctamente"})
}

func resetPassword(c *gin.Context) {
	var request struct {
		Email       string `json:"email"`
		NewPassword string `json:"newPassword"`
		Code        string `json:"code"`
	}

	// Parsear el JSON de entrada
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var userId int
	// Verificar que el código es válido y no ha expirado
	err := db.QueryRow(`
		SELECT user_id 
		FROM reset_codes 
		WHERE code = $1 AND expiration_time > NOW()`,
		request.Code).Scan(&userId)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Código no válido o expirado"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar el código"})
		}
		return
	}

	// Verificar que el código corresponde al email
	var dbEmail string
	err = db.QueryRow("SELECT email FROM users WHERE id = $1", userId).Scan(&dbEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar el correo del usuario"})
		return
	}

	// Comparación case-insensitive de emails
	if strings.ToLower(dbEmail) != strings.ToLower(request.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Correo y código no coinciden"})
		return
	}

	// Actualizar la contraseña en la base de datos
	_, err = db.Exec("UPDATE users SET password = $1 WHERE id = $2", request.NewPassword, userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al actualizar la contraseña"})
		return
	}

	// Eliminar el código usado (opcional)
	_, _ = db.Exec("DELETE FROM reset_codes WHERE code = $1", request.Code)

	// Contraseña actualizada correctamente
	c.JSON(http.StatusOK, gin.H{"message": "Contraseña actualizada correctamente"})
}

func getSMTPConfigHandler(c *gin.Context) {
	var config SMTPConfig
	query := `SELECT id, host, port, username, password, from_email, is_active, created_at, updated_at 
	          FROM smtp_config WHERE is_active = TRUE LIMIT 1`

	err := db.QueryRow(query).Scan(
		&config.ID,
		&config.Host,
		&config.Port,
		&config.Username,
		&config.Password,
		&config.FromEmail,
		&config.IsActive,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "No hay configuración SMTP activa"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener configuración SMTP"})
		return
	}

	c.JSON(http.StatusOK, config)
}

func createSMTPConfigHandler(c *gin.Context) {
	var config SMTPConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos: " + err.Error()})
		return
	}

	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al iniciar transacción"})
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE smtp_config SET is_active = FALSE")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al desactivar configuraciones existentes"})
		return
	}

	query := `INSERT INTO smtp_config 
	          (host, port, username, password, from_email, is_active) 
	          VALUES ($1, $2, $3, $4, $5, TRUE)
	          RETURNING id, created_at, updated_at`

	err = tx.QueryRow(query,
		config.Host,
		config.Port,
		config.Username,
		config.Password,
		config.FromEmail,
	).Scan(&config.ID, &config.CreatedAt, &config.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al crear configuración SMTP: " + err.Error()})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al confirmar transacción"})
		return
	}

	config.IsActive = true
	c.JSON(http.StatusCreated, config)
}

func updateSMTPConfigHandler(c *gin.Context) {
	var config SMTPConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}

	var currentID int
	err := db.QueryRow("SELECT id FROM smtp_config WHERE is_active = TRUE LIMIT 1").Scan(&currentID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "No hay configuración SMTP activa para actualizar"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar configuración existente"})
		return
	}

	query := `UPDATE smtp_config SET 
	          host = $1, port = $2, username = $3, password = $4, from_email = $5, updated_at = NOW()
	          WHERE id = $6
	          RETURNING created_at, updated_at`

	err = db.QueryRow(query,
		config.Host,
		config.Port,
		config.Username,
		config.Password,
		config.FromEmail,
		currentID,
	).Scan(&config.CreatedAt, &config.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al actualizar configuración SMTP"})
		return
	}

	config.ID = currentID
	config.IsActive = true
	c.JSON(http.StatusOK, config)
}

func deleteSMTPConfigHandler(c *gin.Context) {
	var currentID int
	err := db.QueryRow("SELECT id FROM smtp_config WHERE is_active = TRUE LIMIT 1").Scan(&currentID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "No hay configuración SMTP activa para eliminar"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar configuración existente"})
		return
	}

	_, err = db.Exec("DELETE FROM smtp_config WHERE id = $1", currentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al eliminar configuración SMTP"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Configuración SMTP eliminada correctamente"})
}

func testSMTPConnectionHandler(c *gin.Context) {
	var config struct {
		Host      string `json:"host" binding:"required"`
		Port      int    `json:"port" binding:"required"`
		Username  string `json:"username" binding:"required"`
		Password  string `json:"password" binding:"required"`
		FromEmail string `json:"from_email" binding:"required"`
	}

	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Validación adicional
	if config.Port <= 0 || config.Port > 65535 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid port number",
			"field": "port",
		})
		return
	}

	// Configurar timeout
	timeout := 15 * time.Second
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	// Paso 1: Conexión básica TCP
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Could not connect to SMTP server",
			"details":    err.Error(),
			"suggestion": "Verify server address and port availability",
		})
		return
	}
	defer conn.Close()

	// Configurar timeout para operaciones posteriores
	conn.SetDeadline(time.Now().Add(timeout))

	// Paso 2: Crear cliente SMTP
	client, err := smtp.NewClient(conn, config.Host)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "SMTP client creation failed",
			"details": err.Error(),
		})
		return
	}
	defer client.Close()

	// Paso 3: Manejo de STARTTLS
	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{
			ServerName:         config.Host,
			InsecureSkipVerify: true, // Solo para desarrollo/testing
		}

		if err := client.StartTLS(tlsConfig); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":      "TLS negotiation failed",
				"details":    err.Error(),
				"suggestion": "The server may require different TLS settings",
			})
			return
		}
	}

	// Paso 4: Autenticación
	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)
	if err := client.Auth(auth); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Authentication failed",
			"details":    err.Error(),
			"suggestion": "Verify username and password",
			"field":      "auth", // Indica que el problema está en las credenciales
		})
		return
	}

	// Paso 5: Verificar dirección del remitente
	if err := client.Mail(config.FromEmail); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Sender address not accepted",
			"details": err.Error(),
			"field":   "from_email",
		})
		return
	}

	// Éxito - conexión SMTP verificada
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "SMTP connection successfully verified",
		"details": fmt.Sprintf("Connected to %s:%d with TLS support", config.Host, config.Port),
	})
}
