package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

const taxRate = 0.1
const taxFactor = 1.1
const token = "test"

const DATABASE_PASSWORD = "IE76qzUk0t78JGhTz"

func main() {
	r := gin.Default()
	r.Use(gin.Recovery())
	file, fileErr := os.Create("log")
	if fileErr != nil {
		return
	}
	gin.DefaultWriter = file

	//r.Use(cors.Default())
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
		if c.GetHeader("Authorization") != "Bearer "+token {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		}
	}
}
