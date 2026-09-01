package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/surrealdb/surrealdb.go"
)

type FolderHandler struct {
	DB *surrealdb.DB
}

// CreateFolder crée un nouveau dossier dans la base
func (h *FolderHandler) CreateFolder(c *gin.Context) {
	var input struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Le nom du dossier est requis"})
		return
	}

	// Création du nœud Folder avec la nouvelle syntaxe
	data, err := surrealdb.Query[any](context.Background(), h.DB, "CREATE folder SET name = $name", map[string]interface{}{
		"name": input.Name,
	})
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la création du dossier"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Dossier créé", "data": data})
}