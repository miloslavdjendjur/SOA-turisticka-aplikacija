package main

import (
	"fmt"
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


func BuyTour(c *gin.Context) {
	var purchase TourPurchase
	if err := c.ShouldBindJSON(&purchase); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	purchase.Token = fmt.Sprintf("TOKEN-%d-%d", purchase.TouristID, purchase.TourID)
	purchase.PurchaseDate = time.Now()

	if err := DB.Create(&purchase).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, purchase)
}

func UpdatePosition(c *gin.Context) {
	var pos TouristPosition
	if err := c.ShouldBindJSON(&pos); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pos.UpdatedAt = time.Now()

	var existing TouristPosition
	if err := DB.Where("tourist_id = ?", pos.TouristID).First(&existing).Error; err == nil {
		// Update
		existing.Latitude = pos.Latitude
		existing.Longitude = pos.Longitude
		existing.UpdatedAt = time.Now()
		DB.Save(&existing)
		c.JSON(http.StatusOK, existing)
	} else {
		
		DB.Create(&pos)
		c.JSON(http.StatusCreated, pos)
	}
}

func StartTour(c *gin.Context) {
	var exec TourExecution
	if err := c.ShouldBindJSON(&exec); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var purchase TourPurchase
	if err := DB.Where("tourist_id = ? AND tour_id = ?", exec.TouristID, exec.TourID).First(&purchase).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Morate prvo kupiti turu!"})
		return
	}

	exec.Status = "STARTED"
	exec.StartTime = time.Now()
	exec.LastActivity = time.Now()

	if err := DB.Create(&exec).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, exec)
}