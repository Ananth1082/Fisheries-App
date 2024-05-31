package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type glocation struct {
	Name      string  `gorm:"column:name"`
	Longitude float64 `gorm:"column:longitude"`
	Latitude  float64 `gorm:"column:latitude"`
}

func main() {
	dsn := "postgresql://ananth:b9BTvlD_4VeqvxpqzbneMw@poster-app-4864.7s5.aws-ap-south-1.cockroachlabs.cloud:26257/defaultdb?sslmode=verify-full"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database", err)
	}
	var locations []glocation

	if err := db.Find(&locations).Error; err != nil {
		log.Fatal(err)
	}
	for _, loc := range locations {
		fmt.Printf("Name: %s, \n", loc.Name)
	}
	r := gin.Default()
	r.GET("/locations", func(ctx *gin.Context) {
		ctx.JSON(200, locations)
	})
	r.Run()
}
