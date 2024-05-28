package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/oschwald/geoip2-golang"

	"github.com/gin-gonic/gin"
)

const geoblock = false

const taxRate = 0.1
const taxFactor = 1.1
const token = "W_97xyk8G]]w"

const DATABASE_PASSWORD = "IE76qzUk0t78JGhTz"

func main() {
	r := gin.Default()
	r.Use(CORSMiddleware())
	r.Use(gin.Recovery())
	file, fileErr := os.Create("log")
	if fileErr != nil {
		return
	}
	gin.DefaultWriter = file
	if geoblock {
		r.Use(Geoblock())
	}
	r.Use(Authorize())
	r.POST("/pay", pay)
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "0")
	})
	r.POST("/addAccount", addAccount)
	r.POST("/balanceCheck", checkBalance)
	r.POST("/getLogs", getLogs)
	r.POST("/verify", verfiy_account)

	gin.SetMode(gin.ReleaseMode)
	err := r.RunTLS(":8443", "fullchain.pem", "privkey.pem")
	fmt.Println(err)
}

func currTime() string {
	locat, error := time.LoadLocation("Europe/Berlin")
	var dt time.Time
	if error != nil {
		dt = time.Now()
	} else {
		dt = time.Now().In(locat)
	}
	return dt.Format("01-02-2006 15:04:05")
}

func Authorize() gin.HandlerFunc {
	return func(c *gin.Context) {
		//fmt.Println("Bearer " + token)
		if c.GetHeader("Authorization") != "Bearer "+token {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		}
	}
}
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
func Geoblock() gin.HandlerFunc {
	return func(c *gin.Context) {
		db, err := geoip2.Open("GeoLite2-Country.mmdb")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		// If you are using strings that may be invalid, check that ip is not nil
		ip := net.ParseIP(c.ClientIP())
		record, err := db.Country(ip)
		if err != nil {
			log.Fatal(err)
		}
		if record.Country.IsoCode != "DE" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized/geoip"})
		}
	}
}
