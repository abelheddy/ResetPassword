package main

import (
	"fmt"
	"net/http"
	"os"
	"password-recovery/config"
	"password-recovery/routes"
)

func main() {
	// Configuración de credenciales
	key := []byte("12345678901234567890123456789012") // Tu key de prueba
	
	// Cargar credenciales encriptadas
	user, pass, err := config.LoadEncryptedCredentials(key)
	if err != nil {
		fmt.Println("⚠️ Usando credenciales por defecto (para desarrollo):")

	} else {
		fmt.Println("🔑 Credenciales cargadas desde archivo encriptado")
		config.SetupUser = user
		config.SetupPassword = pass
	}

	// Mostrar credenciales (solo para desarrollo)
	//fmt.Printf("\n🔐 Credenciales de acceso:\nUsuario: %s\nContraseña: %s\n\n", 
	//	config.SetupUser, config.SetupPassword)

	// Configurar servidor
	router := routes.SetupRouter()

	port := ":8080"
	if portEnv := os.Getenv("PORT"); portEnv != "" {
		port = ":" + portEnv
	}

	fmt.Printf("🚀 Servidor iniciado en http://localhost%s\n", port)
	fmt.Println("📡 Endpoints disponibles:")
	fmt.Println("- POST /api/login-setup")
	fmt.Println("- POST /api/setup-db")
	fmt.Println("- GET  /api/status")

	if err := http.ListenAndServe(port, router); err != nil {
		fmt.Printf("❌ Error iniciando servidor: %v\n", err)
		os.Exit(1)
	}
}