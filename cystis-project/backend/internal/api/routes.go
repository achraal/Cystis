package api

import (
	"cystis-backend/internal/api/handlers"
	"cystis-backend/internal/api/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/surrealdb/surrealdb.go"
)

func SetupRouter(minioClient *minio.Client, db *surrealdb.DB) *gin.Engine {
	r := gin.Default()

	fileHandler := &handlers.FileHandler{MinioClient: minioClient, DB: db, BucketName: "cystis-storage"}
	folderHandler := &handlers.FolderHandler{DB: db}

	v1 := r.Group("/api/v1")
	// Toutes les requêtes en dessous de cette ligne nécessiteront un token Casdoor valide.
	v1.Use(middlewares.CasdoorAuthMiddleware())
	{
		v1.POST("/folders", folderHandler.CreateFolder)
		v1.POST("/files/upload", fileHandler.Upload)
		v1.DELETE("/files/:id", fileHandler.Delete)
		v1.PUT("/files/:id/recover", fileHandler.Recover)
	}

	return r
}