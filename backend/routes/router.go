package routes

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"password-recovery/config"
)

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Asegura que todas las respuestas sean JSON
func jsonResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func SetupRouter() http.Handler {
	r := mux.NewRouter()
	r.Use(enableCORS)

	// Estado del sistema
	r.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("📡 /api/status fue llamado desde:", r.RemoteAddr)
		setupComplete := config.ConfigExists()
		jsonResponse(w, map[string]interface{}{
			"setup": setupComplete,
			"user":  config.SetupUser,
		}, http.StatusOK)
	}).Methods("GET", "OPTIONS")

	// Login setup
	r.HandleFunc("/api/login-setup", func(w http.ResponseWriter, r *http.Request) {
		if config.ConfigExists() {
			http.Error(w, "El sistema ya fue configurado", http.StatusForbidden)
			return
		}

		var creds struct {
			User string `json:"user"`
			Pass string `json:"pass"`
		}

		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			http.Error(w, "Formato JSON inválido", http.StatusBadRequest)
			return
		}

		// Debug en consola
		configData := map[string]string{
			"Recibido - Usuario":    creds.User,
			"Recibido - Contraseña": creds.Pass,
			"Esperado - Usuario":    config.SetupUser,
			"Esperado - Contraseña": config.SetupPassword,
		}
		debugInfo, _ := json.MarshalIndent(configData, "", "  ")
		fmt.Println("🔍 Intento de login:\n" + string(debugInfo))

		if creds.User == config.SetupUser && creds.Pass == config.SetupPassword {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "success"})
			return
		}

		http.Error(w, "Credenciales inválidas", http.StatusUnauthorized)
	}).Methods("POST", "OPTIONS")

	// Configuración DB
	r.HandleFunc("/api/setup-db", func(w http.ResponseWriter, r *http.Request) {
		if config.ConfigExists() {
			jsonResponse(w, map[string]string{"error": "El sistema ya fue configurado"}, http.StatusForbidden)
			return
		}

		var cfg config.DBConfig
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			jsonResponse(w, map[string]string{"error": "Error en configuración: " + err.Error()}, http.StatusBadRequest)
			return
		}

		if err := config.SaveConfig(cfg); err != nil {
			jsonResponse(w, map[string]string{"error": "Error guardando configuración: " + err.Error()}, http.StatusInternalServerError)
			return
		}

		jsonResponse(w, map[string]interface{}{
			"status": "configuración guardada",
			"setup":  true,
		}, http.StatusOK)
	}).Methods("POST", "OPTIONS")

	return r
}
