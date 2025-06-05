package handlers

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func VerifyCode(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	var request struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var expirationTime time.Time
	var validCode string
	var userId int
	
	err := db.QueryRow(`
		SELECT user_id, code, expiration_time 
		FROM reset_codes 
		WHERE code = $1 AND expiration_time > NOW()`,
		request.Code).Scan(&userId, &validCode, &expirationTime)

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

	c.JSON(http.StatusOK, gin.H{"message": "Código verificado correctamente"})
}