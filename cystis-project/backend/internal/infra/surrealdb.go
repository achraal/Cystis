package infra

import (
	"context"
	"log"
	"os"

	"github.com/surrealdb/surrealdb.go"
)

// InitSchema exécute les requêtes SurrealQL pour préparer les tables et relations (Graphe)
func InitSchema(db *surrealdb.DB) {
	query := `
		-- Configuration de la table Fichier
		DEFINE TABLE file SCHEMAFULL;
		DEFINE FIELD name ON file TYPE string;
		DEFINE FIELD hash ON file TYPE string;
		DEFINE FIELD size ON file TYPE number;
		DEFINE FIELD trashed ON file TYPE bool DEFAULT false;
		DEFINE INDEX idx_hash ON file COLUMNS hash UNIQUE;

		-- Configuration de la table Dossier
		DEFINE TABLE folder SCHEMAFULL;
		DEFINE FIELD name ON folder TYPE string;
		DEFINE FIELD created_at ON folder TYPE datetime DEFAULT time::now();

		-- Configuration de la table Tag
		DEFINE TABLE tag SCHEMAFULL;
		DEFINE FIELD label ON tag TYPE string;

		-- Définition des relations (Graphe)
		-- On rend les relations flexibles (sans contraintes restrictives IN/OUT) 
		-- pour supporter l'arborescence infinie
		DEFINE TABLE contains TYPE RELATION;
		DEFINE TABLE tagged_with TYPE RELATION;
	`

	// Exécution via la fonction globale générique du driver v1.6.0
	_, err := surrealdb.Query[any](context.Background(), db, query, nil)
	if err != nil {
		log.Fatalln("Erreur lors de l'initialisation du schéma SurrealDB:", err)
	}

	log.Println("✅ Schéma SurrealDB (Graphe & Tables) initialisé avec succès")
}

// InitSurrealDB initialise et retourne la connexion à la base de données
func InitSurrealDB() *surrealdb.DB {
	url := os.Getenv("SURREAL_URL")
	user := os.Getenv("SURREAL_USER")
	pass := os.Getenv("SURREAL_PASS")
	ns := os.Getenv("SURREAL_NS")
	dbName := os.Getenv("SURREAL_DB")

	// 1. Connexion au serveur
	db, err := surrealdb.New(url)
	if err != nil {
		log.Fatalln("Erreur de connexion initiale à SurrealDB:", err)
	}

	// 2. Authentification avec le struct surrealdb.Auth
	if _, err = db.SignIn(context.Background(), surrealdb.Auth{
		Username: user,
		Password: pass,
	}); err != nil {
		log.Fatalln("Erreur d'authentification SurrealDB:", err)
	}

	// 3. Sélection du Namespace et de la Base de données
	if err = db.Use(context.Background(), ns, dbName); err != nil {
		log.Fatalln("Erreur lors de la sélection du Namespace et de la DB:", err)
	}

	log.Println("✅ Connecté à SurrealDB avec succès")
	return db
}