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

type glocation struct {
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

func getDistance(point1 Coordinates, point2 Coordinates) float64 {
	return math.Acos(math.Sin(toRadians(point1.Latitude))*math.Sin(toRadians(point2.Latitude))+
		math.Cos(toRadians(point1.Latitude))*math.Cos(toRadians(point2.Latitude))*
			math.Cos(toRadians(point2.Longitude)-toRadians(point1.Longitude))) * 6371
}

func get_coordinates(coordinate string) (float64, float64, error) {
	coordinate, err := url.QueryUnescape(coordinate)
	if err != nil {
		return 0, 0, fmt.Errorf("error decoding coordinates: %w", err)
	}
	lat, long, found := strings.Cut(coordinate, ",")
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
		log.Fatal("failed to connect database", err)
	}

	r := gin.Default()
	r.GET("/locations", func(ctx *gin.Context) {
		var locations []glocation
		if err := db.Find(&locations).Error; err != nil {
			log.Fatal(err)
		}
		ctx.JSON(200, locations)
	})
	r.GET("/distancebetween/:coordinates1/:coordinates2", func(ctx *gin.Context) {
		coordinates1 := ctx.Param("coordinates1")
		coordinates2 := ctx.Param("coordinates2")
		lat1, long1, err := get_coordinates(coordinates1)
		if err != nil {
			ctx.JSON(400, gin.H{"error": err.Error()})
			return
		}
		lat2, long2, err := get_coordinates(coordinates2)
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
		user_coordinates := ctx.Param("coordinates")
		rng := ctx.Param("range")
		frng,_ := strconv.ParseFloat(rng,64) 
		lat,long,_ := get_coordinates(user_coordinates)
		var locations []glocation

		if err := db.Find(&locations).Error; err != nil {
			log.Fatal(err)
		}
		filtered_list := []glocation{} 
		for _, loc := range locations {
			distance := getDistance(Coordinates{Latitude: lat,Longitude:long},Coordinates{Latitude: loc.Latitude,Longitude: loc.Longitude})
			
			if distance <= frng {
				filtered_list = append(filtered_list, loc)
			}
		}
		ctx.JSON(200,filtered_list)
	})
	r.Run()
}
