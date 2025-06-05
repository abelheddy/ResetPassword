package handlers

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type RequestCode struct {
	Email string `json:"email"`
}

func SendCode(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	var request RequestCode

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	request.Email = strings.TrimSpace(strings.ToLower(request.Email))

	var userId int
	err := db.QueryRow("SELECT id FROM users WHERE email = $1", request.Email).Scan(&userId)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Correo no encontrado"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar el correo"})
		}
		return
	}

	n, err := rand.Int(rand.Reader, big.NewInt(100000000))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al generar código"})
		return
	}
	code := fmt.Sprintf("%08d", n)

	expirationTime := time.Now().Add(5 * time.Minute)

	_, err = db.Exec("INSERT INTO reset_codes (user_id, code, expiration_time) VALUES ($1, $2, $3)",
		userId, code, expirationTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al guardar el código"})
		return
	}

	// En producción, usarías el servicio de email aquí
	// emailService := email.NewEmailService(db)
	// err = emailService.SendEmail(request.Email, "Restablecimiento de contraseña", "Tu código es: "+code)
	
	c.JSON(http.StatusOK, gin.H{"message": "Código enviado correctamente"})
}

func ResetPassword(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	var request struct {
		Email       string `json:"email"`
		NewPassword string `json:"newPassword"`
		Code        string `json:"code"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var userId int
	err := db.QueryRow("SELECT user_id FROM reset_codes WHERE code = $1 AND expiration_time > NOW()", 
		request.Code).Scan(&userId)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Código no válido o expirado"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar el código"})
		}
		return
	}

	var dbEmail string
	err = db.QueryRow("SELECT email FROM users WHERE id = $1", userId).Scan(&dbEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar el correo del usuario"})
		return
	}

	if strings.ToLower(dbEmail) != strings.ToLower(request.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Correo y código no coinciden"})
		return
	}

	_, err = db.Exec("UPDATE users SET password = $1 WHERE id = $2", request.NewPassword, userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al actualizar la contraseña"})
		return
	}

	_, _ = db.Exec("DELETE FROM reset_codes WHERE code = $1", request.Code)

	c.JSON(http.StatusOK, gin.H{"message": "Contraseña actualizada correctamente"})
}