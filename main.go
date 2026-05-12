package main

import "github.com/saleh-ghazimoradi/GopherMarket/cmd"

// @title          GopherMarket API
// @version        1.0
// @description    A modern e-commerce API built with pure Go, GORM, and Redis
// @termsOfService http://swagger.io/terms/

// @contact.name   Saleh Ghazimoradi
// @contact.url    http://github.com/saleh-ghazimoradi
// @contact.email  no-email@no-email

// @license.name   Apache 2.0
// @license.url    http://www.apache.org/licenses/LICENSE-2.0.html

// @host           localhost:4000
// @BasePath       /v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Bearer token (e.g., "Bearer {token}")
func main() {
	cmd.Execute()
}
