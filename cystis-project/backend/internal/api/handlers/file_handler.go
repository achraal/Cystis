package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/surrealdb/surrealdb.go"
)

type FileHandler struct {
	MinioClient *minio.Client
	DB          *surrealdb.DB
	BucketName  string
}

func (h *FileHandler) Upload(c *gin.Context) {
	userID := c.GetString("userID") // Récupéré depuis le middleware Casdoor
	folderID := c.PostForm("folder_id") // Le dossier de destination (ex: folder:123)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Aucun fichier détecté"})
		return
	}
	defer file.Close()

	hasher := sha256.New()
	teeReader := io.TeeReader(file, hasher)
	tempObjectName := "temp_" + header.Filename

	// 1. Streaming Upload vers MinIO (Multipart auto > 16Mo)
	_, err = h.MinioClient.PutObject(context.Background(), h.BucketName, tempObjectName, teeReader, header.Size, minio.PutObjectOptions{
		ContentType: header.Header.Get("Content-Type"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Échec de l'upload physique"})
		return
	}

	fileHash := hex.EncodeToString(hasher.Sum(nil))

	// 2. Déduplication : Le fichier existe-t-il déjà dans tout Cystis ?
	res, _ := surrealdb.Query[any](context.Background(), h.DB, "SELECT * FROM file WHERE hash = $hash", map[string]interface{}{"hash": fileHash})
	
	if res != nil && fmt.Sprintf("%v", res) != "&[{[] ok}]" && fmt.Sprintf("%v", res) != "&[]" {
		// DÉDUPLICATION : Le fichier existe. On supprime le temp de MinIO
		_ = h.MinioClient.RemoveObject(context.Background(), h.BucketName, tempObjectName, minio.RemoveObjectOptions{})
		
		// ARBORESCENCE (Graphe) : On crée juste une relation vers le fichier existant
		if folderID != "" {
			_, _ = surrealdb.Query[any](context.Background(), h.DB, "RELATE type::thing($folder)->contains->(SELECT id FROM file WHERE hash = $hash)", map[string]interface{}{
				"folder": folderID,
				"hash": fileHash,
			})
		}
		c.JSON(http.StatusOK, gin.H{"message": "Fichier dédupliqué. Lien graphe créé.", "hash": fileHash})
		return
	}

	// 3. Nouveau Fichier : On le renomme proprement sur MinIO
	src := minio.CopySrcOptions{Bucket: h.BucketName, Object: tempObjectName}
	dst := minio.CopyDestOptions{Bucket: h.BucketName, Object: fileHash}
	_, _ = h.MinioClient.CopyObject(context.Background(), dst, src)
	_ = h.MinioClient.RemoveObject(context.Background(), h.BucketName, tempObjectName, minio.RemoveObjectOptions{})

	// 4. Enregistrement & Arborescence (Graphe)
	// On crée le fichier ET on le relie au dossier en une seule transaction !
	query := `
		LET $new_file = (CREATE file SET name = $name, hash = $hash, size = $size, owner = $user);
		IF $folder != "" THEN
			RELATE type::thing($folder)->contains->$new_file
		END;
	`
	_, _ = surrealdb.Query[any](context.Background(), h.DB, query, map[string]interface{}{
		"name": header.Filename,
		"hash": fileHash,
		"size": header.Size,
		"user": userID,
		"folder": folderID,
	})

	c.JSON(http.StatusCreated, gin.H{"message": "Nouveau fichier uploadé et lié avec succès", "hash": fileHash})
}

// Corbeille Intelligente : Suppression (Soft Delete)
func (h *FileHandler) Delete(c *gin.Context) {
	fileID := c.Param("id") 
	_, _ = surrealdb.Query[any](context.Background(), h.DB, "UPDATE type::thing($id) SET trashed = true", map[string]interface{}{"id": fileID})
	c.JSON(http.StatusOK, gin.H{"message": "Fichier déplacé vers la corbeille"})
}

// Corbeille Intelligente : Restauration
func (h *FileHandler) Recover(c *gin.Context) {
	fileID := c.Param("id") 
	_, _ = surrealdb.Query[any](context.Background(), h.DB, "UPDATE type::thing($id) SET trashed = false", map[string]interface{}{"id": fileID})
	c.JSON(http.StatusOK, gin.H{"message": "Fichier restauré de la corbeille"})
}