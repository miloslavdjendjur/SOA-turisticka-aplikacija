package main

import (
	"time"
	"github.com/lib/pq" 
)

type TourStatus string

const (
	Draft     TourStatus = "DRAFT"
	Published TourStatus = "PUBLISHED"
	Archived  TourStatus = "ARCHIVED"
)


type TourDifficulty string

const (
	Easy   TourDifficulty = "EASY"
	Medium TourDifficulty = "MEDIUM"
	Hard   TourDifficulty = "HARD"
)


type Tour struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	AuthorID    int            `json:"authorId"` 
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Difficulty  TourDifficulty `json:"difficulty"`
	Tags        pq.StringArray `gorm:"type:text[]" json:"tags"` 
	Status      TourStatus     `json:"status"`
	Price       float64        `json:"price"`
	Distance    float64        `json:"distance"` 
	PublishDate time.Time      `json:"publishDate"`
	ArchiveDate time.Time      `json:"archiveDate"`
	
	
	KeyPoints   []KeyPoint     `gorm:"foreignKey:TourID" json:"keyPoints"`
}

type KeyPoint struct {
	ID          uint    `gorm:"primaryKey" json:"id"`
	TourID      uint    `json:"tourId"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Image       string  `json:"image"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Order       int     `json:"order"` 
}

type Blog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	AuthorID    int       `json:"authorId"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DateCreated time.Time `json:"dateCreated"`
	Image       string    `json:"image"`
}