package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Me(c *gin.Context) {
	claims, _ := c.Get("user") // set จาก middleware
	c.JSON(http.StatusOK, gin.H{"user": claims})
}
