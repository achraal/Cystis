package middlewares

import (
	"net/http"
	"strings"

	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	"github.com/gin-gonic/gin"
)

func CasdoorAuthMiddleware() gin.HandlerFunc {
	// Initialisation de la config Casdoor (idéalement à mettre dans infra, mais placé ici pour la simplicité du test)
	casdoorsdk.InitConfig("http://localhost:8080", "ton_client_id", "ton_client_secret", "", "built-in", "app-built-in")

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token d'authentification manquant"})
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		
		// Casdoor vérifie la validité du Token JWT
		claims, err := casdoorsdk.ParseJwtToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token invalide ou expiré"})
			return
		}

		// On extrait l'identité de l'utilisateur (ex: "admin") et on la stocke pour les requêtes (Déduplication, etc.)
		c.Set("userID", claims.Name)
		c.Next()
	}
}