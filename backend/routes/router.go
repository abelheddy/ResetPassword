package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
	"password-recovery/config"
)

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type, Authorization, X-Requested-With")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func jsonResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func SetupRouter() http.Handler {
	r := mux.NewRouter()
	r.Use(enableCORS)

	// Actualizar estado inicial
	config.RefreshConfigState()

	// Cargar configuración inicial
	// Intenta cargar la configuración
	// Cargar configuración con verificación de conexión
	if err := loadAndVerifyConfig(); err != nil {
		fmt.Printf("⚠️ Advertencia: %v\n", err)
	} else {
		fmt.Println("✅ Configuración de DB cargada y verificada exitosamente")
	}

	// Inicializar estado de configuración
	config.CurrentSetupStatus = config.SetupStatus{
		DBConfigured:    config.ConfigExists(),
		DBTablesCreated: false,
		AdminCreated:    false,
	}

	// Estado del sistema - ACTUALIZADO
	// Endpoint para estado
	r.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		// Actualizar estado verificando conexión REAL
		loadAndVerifyConfig()

		setupComplete := config.IsSetupComplete()
		jsonResponse(w, map[string]interface{}{
			"setup": setupComplete,
			"user":  config.SetupUser,
			"setup_stages": map[string]bool{
				"db_configured":     config.CurrentSetupStatus.DBConfigured,
				"tables_created":    config.CurrentSetupStatus.DBTablesCreated,
				"admin_created":     config.CurrentSetupStatus.AdminCreated,
				"allow_reconfigure": true,
			},
			// Información actual de conexión
			"db_info": map[string]string{
				"host":   config.GetCurrentConfig().Host,
				"port":   config.GetCurrentConfig().Port,
				"dbname": config.GetCurrentConfig().DBName,
				"user":   config.GetCurrentConfig().User,
			},
		}, http.StatusOK)
	}).Methods("GET", "OPTIONS")

	// Login setup
	r.HandleFunc("/api/login-setup", func(w http.ResponseWriter, r *http.Request) {
		if config.ConfigExists() && !config.CurrentSetupStatus.DBConfigured {
			jsonResponse(w, map[string]interface{}{
				"error": "El sistema ya tiene configuración de DB",
			}, http.StatusForbidden)
			return
		}

		var creds struct {
			User string `json:"user"`
			Pass string `json:"pass"`
		}

		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			jsonResponse(w, map[string]interface{}{
				"error": "Formato JSON inválido",
			}, http.StatusBadRequest)
			return
		}

		if creds.User == config.SetupUser && creds.Pass == config.SetupPassword {
			jsonResponse(w, map[string]interface{}{
				"status": "success",
			}, http.StatusOK)
			return
		}

		jsonResponse(w, map[string]interface{}{
			"error": "Credenciales inválidas",
		}, http.StatusUnauthorized)
	}).Methods("POST", "OPTIONS")

	// Configuración DB
	r.HandleFunc("/api/setup-db", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("📡 /api/setup-db llamado desde:", r.RemoteAddr)

		// Verificar si ya está configurado (a menos que permitamos reconfiguración)
		if config.ConfigExists() && !config.CurrentSetupStatus.DBConfigured {
			jsonResponse(w, map[string]interface{}{
				"error": "El sistema ya tiene configuración de DB",
			}, http.StatusForbidden)
			return
		}

		// Decodificar el JSON recibido
		var cfg config.DBConfig
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			jsonResponse(w, map[string]interface{}{
				"error": "Error en formato de configuración: " + err.Error(),
			}, http.StatusBadRequest)
			return
		}

		// Validar campos obligatorios
		if cfg.Host == "" || cfg.User == "" || cfg.DBName == "" || cfg.Port == "" {
			jsonResponse(w, map[string]interface{}{
				"error": "Faltan campos obligatorios: host, user, dbname o port",
			}, http.StatusBadRequest)
			return
		}

		// Probar conexión ANTES de guardar
		testResult, err := cfg.TestConnection()
		if err != nil {
			fmt.Printf("❌ Error probando conexión: %v\n", err)
			jsonResponse(w, map[string]interface{}{
				"error": "Error probando conexión: " + err.Error(),
				"details": map[string]interface{}{
					"host":   cfg.Host,
					"port":   cfg.Port,
					"dbname": cfg.DBName,
					"user":   cfg.User,
				},
			}, http.StatusBadRequest)
			return
		}

		fmt.Printf("✅ Conexión probada exitosamente: %+v\n", testResult)

		// Guardar configuración
		if err := config.SaveConfig(cfg); err != nil {
			fmt.Printf("❌ Error guardando configuración: %v\n", err)
			jsonResponse(w, map[string]interface{}{
				"error": "Error guardando configuración: " + err.Error(),
			}, http.StatusInternalServerError)
			return
		}

		// Actualizar estado
		config.CurrentSetupStatus.DBConfigured = true

		fmt.Println("🔧 Configuración guardada exitosamente")

		// Responder con éxito
		jsonResponse(w, map[string]interface{}{
			"status":          "success",
			"message":         "Configuración guardada y conexión verificada",
			"setup":           true,
			"connection_test": testResult,
			"config": map[string]interface{}{
				"host":   cfg.Host,
				"port":   cfg.Port,
				"dbname": cfg.DBName,
				"user":   cfg.User,
			},
		}, http.StatusOK)
	}).Methods("POST", "OPTIONS")

	// Endpoint para crear tablas
	r.HandleFunc("/api/setup/create-tables", func(w http.ResponseWriter, r *http.Request) {
		if !config.CurrentSetupStatus.DBConfigured {
			jsonResponse(w, map[string]interface{}{
				"success": false,
				"error":   "Database connection not configured",
			}, http.StatusBadRequest)
			return
		}

		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
			config.GetCurrentConfig().Host,
			config.GetCurrentConfig().User,
			config.GetCurrentConfig().Password,
			config.GetCurrentConfig().DBName,
			config.GetCurrentConfig().Port)

		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			jsonResponse(w, map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			}, http.StatusInternalServerError)
			return
		}

		if err := config.InitializeDB(db); err != nil {
			jsonResponse(w, map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			}, http.StatusInternalServerError)
			return
		}

		jsonResponse(w, map[string]interface{}{
			"success": true,
			"message": "Tables created successfully",
		}, http.StatusOK)
	}).Methods("POST", "OPTIONS")

	// Endpoint para crear admin
	r.HandleFunc("/api/setup/create-admin", func(w http.ResponseWriter, r *http.Request) {
		if !config.CurrentSetupStatus.DBTablesCreated {
			jsonResponse(w, map[string]interface{}{
				"success": false,
				"error":   "Database tables not created",
			}, http.StatusBadRequest)
			return
		}

		var request struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			jsonResponse(w, map[string]interface{}{
				"success": false,
				"error":   "Invalid request body",
			}, http.StatusBadRequest)
			return
		}

		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
			config.GetCurrentConfig().Host,
			config.GetCurrentConfig().User,
			config.GetCurrentConfig().Password,
			config.GetCurrentConfig().DBName,
			config.GetCurrentConfig().Port)

		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			jsonResponse(w, map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			}, http.StatusInternalServerError)
			return
		}

		created, err := config.CreateAdminUser(db, request.Email, request.Password)
		if err != nil {
			jsonResponse(w, map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			}, http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"success": true,
			"message": "Admin user created successfully",
		}

		if !created {
			response["message"] = "Admin user already exists, setup completed"
			response["already_exists"] = true
		}

		jsonResponse(w, response, http.StatusOK)
	}).Methods("POST", "OPTIONS")

	// Nuevo endpoint para resetear configuración
	// routes/router.go
	r.HandleFunc("/api/setup/reset", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("🔄 Solicitud de reset recibida")

		if err := config.ResetConfig(); err != nil {
			fmt.Printf("❌ Error en ResetConfig: %v\n", err)
			jsonResponse(w, map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			}, http.StatusInternalServerError)
			return
		}

		// Verificar eliminación usando el nombre del archivo directamente
		if _, err := os.Stat("dbconfig.json"); !os.IsNotExist(err) {
			fmt.Println("⚠️ Advertencia: dbconfig.json todavía existe")
			jsonResponse(w, map[string]interface{}{
				"success": false,
				"error":   "El archivo de configuración no pudo ser eliminado",
			}, http.StatusInternalServerError)
			return
		}

		config.RefreshConfigState()

		jsonResponse(w, map[string]interface{}{
			"success": true,
			"message": "Configuración completamente reseteada",
		}, http.StatusOK)
	}).Methods("POST")

	//endpoint temporar para verificar rutas y permisos
	r.HandleFunc("/api/debug/config-path", func(w http.ResponseWriter, r *http.Request) {
		path := config.GetConfigPath()
		info := make(map[string]interface{})

		fileInfo, err := os.Stat(path)
		if err == nil {
			info["exists"] = true
			info["path"] = path
			info["permissions"] = fileInfo.Mode().String()
			info["size"] = fileInfo.Size()
			info["mod_time"] = fileInfo.ModTime()
		} else {
			info["exists"] = false
			info["error"] = err.Error()
		}

		// Verificar permisos de escritura
		if file, err := os.OpenFile(path, os.O_WRONLY, 0666); err == nil {
			file.Close()
			info["writable"] = true
		} else {
			info["writable"] = false
			info["write_error"] = err.Error()
		}

		jsonResponse(w, info, http.StatusOK)
	}).Methods("GET")

	// Endpoint para probar configuración temporal
	r.HandleFunc("/api/db/test-config", func(w http.ResponseWriter, r *http.Request) {
		var testConfig config.DBConfig
		if err := json.NewDecoder(r.Body).Decode(&testConfig); err != nil {
			jsonResponse(w, map[string]interface{}{
				"success": false,
				"error":   "Formato de configuración inválido",
			}, http.StatusBadRequest)
			return
		}

		result, err := testConfig.TestConnection()
		if err != nil {
			jsonResponse(w, map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			}, http.StatusInternalServerError)
			return
		}

		jsonResponse(w, map[string]interface{}{
			"success": true,
			"result":  result,
		}, http.StatusOK)
	}).Methods("POST", "OPTIONS")

	// Endpoint para probar conexión DB
	r.HandleFunc("/api/db/test", func(w http.ResponseWriter, r *http.Request) {
		currentConfig := config.GetCurrentConfig()
		result, err := currentConfig.TestConnection()
		if err != nil {
			jsonResponse(w, map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			}, http.StatusInternalServerError)
			return
		}

		// Añadir información adicional
		result["dbname"] = currentConfig.DBName
		result["user"] = currentConfig.User

		jsonResponse(w, map[string]interface{}{
			"success": true,
			"result":  result,
		}, http.StatusOK)
	}).Methods("POST", "OPTIONS")

	// Endpoint para obtener/configurar DB
	r.HandleFunc("/api/db/config", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			jsonResponse(w, map[string]interface{}{
				"success": true,
				"config":  config.GetDBConfig(),
			}, http.StatusOK)

		case "PUT":
			var newConfig config.DBConfig
			if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
				jsonResponse(w, map[string]interface{}{
					"success": false,
					"error":   "Formato de configuración inválido",
				}, http.StatusBadRequest)
				return
			}

			if err := config.UpdateDBConfig(newConfig); err != nil {
				jsonResponse(w, map[string]interface{}{
					"success": false,
					"error":   err.Error(),
				}, http.StatusInternalServerError)
				return
			}

			// Actualizar estado
			config.CurrentSetupStatus.DBConfigured = true

			jsonResponse(w, map[string]interface{}{
				"success": true,
				"message": "Configuración actualizada",
				"config":  config.GetDBConfig(),
			}, http.StatusOK)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}).Methods("GET", "PUT", "OPTIONS")

	return r
}
func loadAndVerifyConfig() error {
	// Cargar configuración desde archivo
	cfg, err := config.LoadDBConfig()
	if err != nil {
		config.CurrentSetupStatus.DBConfigured = false
		return err
	}

	// Probar conexión real
	_, err = cfg.TestConnection()
	if err != nil {
		config.CurrentSetupStatus.DBConfigured = false
		return fmt.Errorf("error verificando conexión: %v", err)
	}

	config.CurrentSetupStatus.DBConfigured = true
	return nil
}
