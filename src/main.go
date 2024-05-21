package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

const taxRate = 0.1
const taxFactor = 1.1

var token string = os.Getenv("SERVER_TOKEN")

var DATABASE_PASSWORD = os.Getenv("SERVER_DBPIN")

func main() {
	r := gin.Default()
	r.Use(CORSMiddleware())
	r.Use(gin.Recovery())
	file, fileErr := os.Create("log")
	if fileErr != nil {
		return
	}
	gin.DefaultWriter = file

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
	r.RunTLS(":8443", "fullchain.pem", "privkey.pem")
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
		fmt.Println("Bearer " + token)
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
