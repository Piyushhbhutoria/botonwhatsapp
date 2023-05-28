package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func helloworld(c *gin.Context) {
	c.String(http.StatusOK, "The bot is running")
}

func sendText(c *gin.Context) {
	to := strings.Replace(c.DefaultQuery("to", "1234567890"), " ", "", -1)
	mess := c.DefaultQuery("msg", "testing")
	c.String(http.StatusOK, texting(to, mess))
}

func sendBulk(c *gin.Context) {
	var data sendBulkText
	m := make(map[string]string)

	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, each := range data.List {
		each.Receiver = strings.Replace(each.Receiver, " ", "", -1)
		if each.Receiver != "" {
			m[each.Receiver] = texting(each.Receiver, each.Message)
		}
	}
	c.JSON(http.StatusOK, gin.H{"result": data})
}
