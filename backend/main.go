package main

import (
	_ "fratelli-feccia/docs"
	"fratelli-feccia/internal/app"

	"github.com/joho/godotenv"
)

// @title fratelli-feccia API
// @version 1.0
// @description fratelli-feccia backend API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// @host localhost:8080
// @BasePath /
// @schemes  https http
func main() {
	_ = godotenv.Load()

	a := app.New("fratelli-feccia")
	a.Start()
}
