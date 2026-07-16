package swagger

import (
	"fratelli-feccia/config"
	"fratelli-feccia/docs"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

func SetupSwagger(app *fiber.App, cfg *config.Config) {
	app.Get("/api/v1/swagger/*", swagger.HandlerDefault)

	docs.SwaggerInfo.Title = "fratelli-feccia API"
	docs.SwaggerInfo.Description = `
## Authentication
Use Bearer token in Authorization header. Click **Authorize** and enter:
` + "`Bearer <your_access_token>`" + `

## Roles
- **admin** — full access including user management
- **amministrazione** — invoicing, pricelists, accounting master data
- **planner** — dispatch, GPS, driver availability
- **operatore** — master data (customers, vehicles, drivers, ...) and orders`

	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = cfg.Swagger.Host
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Schemes = cfg.Swagger.Schemes
}
