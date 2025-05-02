package main

import (
	"go-simple-warehouse/internal/auth"
	"go-simple-warehouse/internal/config"
	"go-simple-warehouse/internal/dashboard"
	"go-simple-warehouse/internal/database"
	"go-simple-warehouse/internal/middleware"
	"go-simple-warehouse/internal/product"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: no .env file found, relying on real environment")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("config error:", err)
	}

	db, err := database.ConnectDB(cfg)
	if err != nil {
		log.Fatal("db connect:", err)
	}
	db.AutoMigrate(&auth.User{}, &product.Product{})

	if err := database.SeedSuperAdmin(db, cfg); err != nil {
		log.Fatal("db seed:", err)
	}

	// ensure barcodes dir exists
	if err := os.MkdirAll("barcodes", os.ModePerm); err != nil {
		log.Fatal("could not create barcodes folder:", err)
	}

	r := gin.Default()

	// add CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://127.0.0.1:5173"}, // your Vite dev origin
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.Static("/barcodes", "./barcodes")

	r.Use(middleware.Error())

	// routes
	r.POST("/auth/login", auth.LoginHandler(db, cfg))

	r.GET("/dashboard", middleware.JWT(cfg), dashboard.DashboardMetrics(db))

	p := r.Group("/products", middleware.JWT(cfg))
	{
		p.GET("", product.ListProducts(db))
		p.POST("", product.CreateProduct(db))
		p.GET("/:id", product.GetProduct(db))
		p.PUT("/:id", product.UpdateProduct(db))
		p.DELETE("/:id", product.DeleteProduct(db))
		p.GET("/export.csv", product.ExportCSV(db))
	}

	r.Run(":8080")
}
