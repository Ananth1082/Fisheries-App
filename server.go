package main

import (
	"fmt"
	"log"
	"math"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Glocation struct {
	Name      string  `gorm:"column:name"`
	Longitude float64 `gorm:"column:longitude"`
	Latitude  float64 `gorm:"column:latitude"`
}

type Coordinates struct {
	Longitude float64
	Latitude  float64
}

func toRadians(degree float64) float64 {
	return degree * math.Pi / 180
}

func getDistance(point1, point2 Coordinates) float64 {
	return math.Acos(math.Sin(toRadians(point1.Latitude))*math.Sin(toRadians(point2.Latitude))+
		math.Cos(toRadians(point1.Latitude))*math.Cos(toRadians(point2.Latitude))*
			math.Cos(toRadians(point2.Longitude)-toRadians(point1.Longitude))) * 6371
}

func parseCoordinates(coordinate string) (float64, float64, error) {
	decoded, err := url.QueryUnescape(coordinate)
	if err != nil {
		return 0, 0, fmt.Errorf("error decoding coordinates: %w", err)
	}
	lat, long, found := strings.Cut(decoded, ",")
	if !found {
		return 0, 0, fmt.Errorf("invalid coordinate format")
	}
	flat, err := strconv.ParseFloat(strings.Trim(lat, "("), 64)
	if err != nil {
		return 0, 0, err
	}
	flong, err := strconv.ParseFloat(strings.Trim(long, ")"), 64)
	if err != nil {
		return 0, 0, err
	}
	return flat, flong, nil
}

func main() {
	dsn := "postgresql://ananth:b9BTvlD_4VeqvxpqzbneMw@poster-app-4864.7s5.aws-ap-south-1.cockroachlabs.cloud:26257/defaultdb?sslmode=verify-full"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}

	r := gin.Default()

	r.GET("/locations", func(ctx *gin.Context) {
		var locations []Glocation
		if err := db.Find(&locations).Error; err != nil {
			log.Println("Error fetching locations:", err)
			ctx.JSON(500, gin.H{"error": "Error fetching locations"})
			return
		}
		ctx.JSON(200, locations)
	})

	r.GET("/distancebetween/:coordinates1/:coordinates2", func(ctx *gin.Context) {
		coordinates1 := ctx.Param("coordinates1")
		coordinates2 := ctx.Param("coordinates2")

		lat1, long1, err := parseCoordinates(coordinates1)
		if err != nil {
			ctx.JSON(400, gin.H{"error": err.Error()})
			return
		}
		lat2, long2, err := parseCoordinates(coordinates2)
		if err != nil {
			ctx.JSON(400, gin.H{"error": err.Error()})
			return
		}

		point1 := Coordinates{Latitude: lat1, Longitude: long1}
		point2 := Coordinates{Latitude: lat2, Longitude: long2}
		distance := getDistance(point1, point2)

		ctx.JSON(200, gin.H{"distance": distance})
	})

	r.GET("/findbyrange/:coordinates/:range", func(ctx *gin.Context) {
		userCoordinates := ctx.Param("coordinates")
		rng := ctx.Param("range")

		lat, long, err := parseCoordinates(userCoordinates)
		if err != nil {
			ctx.JSON(400, gin.H{"error": err.Error()})
			return
		}

		frng, err := strconv.ParseFloat(rng, 64)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid range value"})
			return
		}

		var locations []Glocation
		if err := db.Find(&locations).Error; err != nil {
			log.Println("Error fetching locations:", err)
			ctx.JSON(500, gin.H{"error": "Error fetching locations"})
			return
		}

		var filteredList []Glocation
		for _, loc := range locations {
			distance := getDistance(Coordinates{Latitude: lat, Longitude: long}, Coordinates{Latitude: loc.Latitude, Longitude: loc.Longitude})
			if distance <= frng {
				filteredList = append(filteredList, loc)
			}
		}

		ctx.JSON(200, filteredList)
	})

	r.Run()
}
