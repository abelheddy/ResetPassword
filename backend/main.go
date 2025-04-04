/*
Servidor de recuperación de contraseña

Este servidor proporciona tres endpoints principales:
1. /send-code - Envía un código de verificación al correo del usuario
2. /verify-code - Verifica si el código ingresado es válido
3. /reset-password - Actualiza la contraseña del usuario

Configuración requerida:
- Base de datos PostgreSQL con tablas 'users' y 'reset_codes'
- Servidor SMTP para envío de correos (configurado para Gmail en este ejemplo)
*/

package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq" // Driver PostgreSQL
	"gopkg.in/gomail.v2"  // Biblioteca para envío de correos
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// db es la conexión global a la base de datos
var db *sql.DB

// RequestCode estructura para el JSON de solicitud de código
type RequestCode struct {
	Email string `json:"email"` // Email del usuario que solicita el código
}

/*
sendEmail envía un correo electrónico al destinatario especificado
Parámetros:
- to: Dirección de correo del destinatario
- subject: Asunto del correo
- body: Cuerpo del mensaje
Retorna:
- error si ocurre algún problema durante el envío
*/
func sendEmail(to string, subject string, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "abelhecks@gmail.com") // Cuenta de correo remitente
	m.SetHeader("To", to)                     // Destinatario
	m.SetHeader("Subject", subject)           // Asunto del correo
	m.SetBody("text/plain", body)             // Cuerpo como texto plano

	// Configuración del servidor SMTP (Gmail)
	dialer := gomail.NewDialer(
		"smtp.gmail.com",                // Servidor SMTP
		587,                             // Puerto
		"abelhecks@gmail.com",           // Usuario
		"exzmnbttgpsoxhdl",              // Contraseña/Token
	)

	// Intenta enviar el correo
	if err := dialer.DialAndSend(m); err != nil {
		log.Println("Error al enviar el correo:", err)
		return err
	}

	log.Println("Correo enviado con éxito a", to)
	return nil
}

/*
sendCode maneja la solicitud de envío de código de verificación
Método: POST
Ruta: /send-code
Body esperado: {"email": "usuario@example.com"}
*/
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

	// Generar código aleatorio de 8 dígitos
	code := fmt.Sprintf("%08d", rand.Intn(100000000))
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

/*
verifyCode verifica si el código de recuperación es válido
Método: POST
Ruta: /verify-code
Body esperado: {"email": "usuario@example.com", "code": "12345678"}
*/
func verifyCode(c *gin.Context) {
	var request struct {
		Email string `json:"email"` // Email del usuario
		Code  string `json:"code"`  // Código de verificación
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

/*
resetPassword actualiza la contraseña del usuario
Método: POST
Ruta: /reset-password
Body esperado: {"email": "usuario@example.com", "newPassword": "nuevaContraseña", "code": "12345678"}
*/
func resetPassword(c *gin.Context) {
	var request struct {
		Email       string `json:"email"`       // Email del usuario
		NewPassword string `json:"newPassword"` // Nueva contraseña
		Code        string `json:"code"`        // Código de verificación
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

	// Contraseña actualizada correctamente
	c.JSON(http.StatusOK, gin.H{"message": "Contraseña actualizada correctamente"})
}

/*
Función principal que inicia el servidor
*/
func main() {
	var err error
	// Conectar a la base de datos PostgreSQL
	db, err = sql.Open("postgres", "user=postgres password=root dbname=db_reset sslmode=disable")
	if err != nil {
		log.Fatal("Error al conectar a la base de datos:", err)
	}
	defer db.Close()

	// Crear router Gin
	r := gin.Default()

	// Configuración CORS (Permitir acceso desde cualquier origen en desarrollo)
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // En producción, reemplazar con dominios específicos
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Registrar rutas
	r.POST("/send-code", sendCode)           // Enviar código de verificación
	r.POST("/verify-code", verifyCode)       // Verificar código
	r.POST("/reset-password", resetPassword) // Restablecer contraseña

	log.Println("Servidor iniciado en el puerto :8080")
	err = r.Run(":8080")
	if err != nil {
		log.Fatal("Error al iniciar el servidor:", err)
	}
}