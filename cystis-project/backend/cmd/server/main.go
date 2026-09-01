package main

import (
	"log"
	"os"

	"cystis-backend/internal/api"
	"cystis-backend/internal/infra"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	minioClient := infra.InitMinio()
	surrealClient := infra.InitSurrealDB()
	infra.InitSchema(surrealClient)

	// On injecte les connexions ici !
	router := api.SetupRouter(minioClient, surrealClient)
	
	port := os.Getenv("PORT")
	if port == "" { port = "8080" }

	log.Printf("🚀 Serveur Cystis démarré sur le port %s\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalln(err)
	}
}