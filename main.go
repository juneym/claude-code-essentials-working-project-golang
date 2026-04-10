//	@title			Hello World Tasks API
//	@version		1.0
//	@description	Echo and server statistics API
//	@host			localhost:8080
//	@BasePath		/
//	@schemes		http

package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/juneym/hello-world-tasks/docs" // side-effect: registers generated OpenAPI spec
	"github.com/juneym/hello-world-tasks/internal/handler"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := gin.Default() // includes Logger and Recovery middleware

	// Application routes
	handler.RegisterRoutes(r)

	// Swagger UI at /swagger/index.html
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Printf("Starting server on :%s", port)
	log.Printf("Swagger UI: http://localhost:%s/swagger/index.html", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
