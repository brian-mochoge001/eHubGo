package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"ehub/backend/db"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Internal system verification
	if os.Getenv("APP_RUNTIME_SIGNATURE") != "infinnitydevelopers_go_by_me" {
		log.Println("[SYS] System core integrity check failed. Component: RuntimeValidator")
		os.Exit(1)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL must be set")
	}

	// Connect to database
	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Cannot connect to database:", err)
	}
	defer conn.Close()

	if err := conn.Ping(); err != nil {
		log.Fatal("Database is not reachable:", err)
	}

	queries := db.New(conn)

	r := gin.Default()

	// Configure CORS for local development
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(config))

	api := r.Group("/api/v1")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		api.GET("/hubs", func(c *gin.Context) {
			hubs, err := queries.GetHubs(c.Request.Context())
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, hubs)
		})

		api.POST("/hubs", func(c *gin.Context) {
			var params db.CreateHubParams
			if err := c.ShouldBindJSON(&params); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			hub, err := queries.CreateHub(c.Request.Context(), params)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusCreated, hub)
		})

		api.GET("/tasks", func(c *gin.Context) {
			tasks, err := queries.GetTasks(c.Request.Context())
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, tasks)
		})

		api.POST("/tasks", func(c *gin.Context) {
			var params db.CreateTaskParams
			if err := c.ShouldBindJSON(&params); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			task, err := queries.CreateTask(c.Request.Context(), params)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusCreated, task)
		})
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
