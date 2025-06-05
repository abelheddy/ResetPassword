package main

import (
	"log"
	"password-recovery/internal/config"
	"password-recovery/internal/database"
	"password-recovery/internal/server"
)

func main() {
	// Cargar configuración
	cfg := config.LoadConfig()

	// Inicializar base de datos
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Error al conectar a la base de datos: %v", err)
	}
	defer db.Close()

	// Crear tablas si no existen
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Error al ejecutar migraciones: %v", err)
	}

	// Iniciar servidor
	if err := server.Start(cfg, db); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}