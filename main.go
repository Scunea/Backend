package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"

	"github.com/Scunea/Backend/routes"
	"github.com/Scunea/Backend/utils"

	"github.com/joho/godotenv"

	"log"

	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	gin.SetMode(os.Getenv("GIN_MODE"))

	r := gin.New()

	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"school", "Content-Type", "Authorization"},
	}))

	r.Use(gin.Logger())

	r.Use(gin.Recovery())

	utils.CreateHmacKey()

	connStr := "user=" + os.Getenv("DB_USER") + " dbname=" + os.Getenv("DB_DATABASE") + " sslmode=disable"
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	utils.InitDb(db)

	go routes.Start(r, db)

	r.Run(":" + os.Getenv("SVR_PORT"))
}
