package controllers

import (

)
/*
// Login para modo setup (único uso)
func LoginSetupHandler(w http.ResponseWriter, r *http.Request) {
	if config.ConfigExists() {
		http.Error(w, "Setup ya realizado", http.StatusForbidden)
		return
	}

	var creds struct {
		User     string `json:"user"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Error en credenciales", http.StatusBadRequest)
		return
	}

	if err := config.ValidateSetupLogin(creds.User, creds.Password); err != nil {
		http.Error(w, "Credenciales inválidas", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "login exitoso"})
}

// Setup DB (guardar configuración)
func SetupDBHandler(w http.ResponseWriter, r *http.Request) {
	if config.ConfigExists() {
		http.Error(w, "Setup ya realizado", http.StatusForbidden)
		return
	}

	var cfg config.DBConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		http.Error(w, "Error en configuración: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validaciones básicas de configuración
	if cfg.Host == "" || cfg.User == "" || cfg.DBName == "" {
		http.Error(w, "Configuración incompleta", http.StatusBadRequest)
		return
	}

	if err := config.SaveConfig(cfg); err != nil {
		http.Error(w, "Error guardando configuración: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "configuración guardada correctamente"})
}

// Endpoint para saber si está configurado el sistema
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	resp := struct {
		Setup bool `json:"setup"`
	}{
		Setup: config.ConfigExists(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}*/