package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func CreateTour(c *gin.Context) {
	var input Tour
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input.Status = Draft
	input.Price = 0
	input.Distance = 0

	if err := DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, input)
}

func AddKeyPoint(c *gin.Context) {
	var input KeyPoint
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
    
    // TODO: azurirati distanca ture

	c.JSON(http.StatusCreated, input)
}

func CreateBlog(c *gin.Context) {
	var input Blog
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input.DateCreated = time.Now()

	if err := DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, input)
}

func GetTours(c *gin.Context) {
    var tours []Tour
    
    DB.Preload("KeyPoints").Find(&tours)
    c.JSON(http.StatusOK, tours)
}