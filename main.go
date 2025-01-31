package main

import (
	"aino_document/routes"
	"aino_document/utils"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := routes.Route()

	e.Use(middleware.Logger())
	e.Use(middleware.CORS())
	// e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
	// 	AllowOrigins: []string{"http://localhost:4200", "http://localhost:8080"}, // Izinkan akses dari dua origin
	// 	AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},       // Izinkan metode permintaan tertentu
	// }))
	customValidator := &utils.CustomValidator{Validator: validator.New()}
	e.Validator = customValidator
	// e.Logger.Fatal(e.Start("192.168.1.12:8080"))
	e.Logger.Fatal(e.Start(":8080"))
	// e.Logger.Fatal(e.Start(":8080"))

}
