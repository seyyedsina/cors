package main

import (
	"github.com/gin-gonic/gin"
	cors "github.com/itsjamie/gin-cors"
	"io/ioutil"
	"log"
	"net/http"
	urlCheck "net/url"
	"strings"
	"time"
)

func main() {
	r := gin.Default()

	r.Use(cors.Middleware(cors.Config{
		Origins:         "*",
		Methods:         "GET, PUT, POST, PATCH, DELETE",
		RequestHeaders:  "Origin, Authorization, Content-Type",
		ExposedHeaders:  "",
		MaxAge:          50 * time.Second,
		Credentials:     true,
		ValidateHeaders: false,
	}))
	r.Use(callApi())

	if err := r.Run(":9091"); err != nil {
		log.Fatalln(err)
	}
}
func callApi() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		url := c.Request.RequestURI
		if strings.HasPrefix(url, "/") {
			url = url[1:]
		}
		if _, err := urlCheck.ParseRequestURI(url); err != nil {
			c.AbortWithStatusJSON(500, gin.H{"message": "bad url!"})
			return
		}
		method := c.Request.Method
		payload := c.Request.Body
		res := new(http.Response)
		var err error
		for {
			if _, err := urlCheck.Parse(url); err != nil {
				c.AbortWithStatusJSON(500, gin.H{"message": "bad url!"})
				return
			}

			req, err := http.NewRequestWithContext(ctx, method, url, payload)
			if err != nil {
				c.AbortWithStatusJSON(500, gin.H{"message": "creating req"})
				return
			}
			for key, value := range c.Request.Header {
				for _, headValue := range value {
					req.Header.Set(key, headValue)
				}
			}
			client := http.DefaultClient
			res, err = client.Do(req)
			if err != nil {
				c.AbortWithStatusJSON(500, gin.H{"message": "response err"})
				return
			}
			if res.StatusCode >= 300 && res.StatusCode < 400 {
				if value, ok := res.Header["Location"]; ok {
					url = value[0]
				}
			} else {
				break
			}
		}
		rr, err := ioutil.ReadAll(res.Body)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"message": "reading body"})
			return
		}
		c.String(res.StatusCode, "%v", string(rr))
		return
	}
}
