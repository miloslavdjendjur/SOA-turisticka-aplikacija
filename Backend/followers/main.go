package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var driver neo4j.DriverWithContext

func initNeo4j() {
	uri := os.Getenv("NEO4J_URI")
	username := os.Getenv("NEO4J_USER")
	password := os.Getenv("NEO4J_PASSWORD")
	var err error
	driver, err = neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		log.Fatal("Neo4j error: ", err)
	}
}

type FollowRequest struct {
	FollowerID int `json:"followerId"`
	FollowedID int `json:"followedId"`
}

func FollowUser(c *gin.Context) {
	var req FollowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := context.Background()
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	query := `
		MERGE (u1:User {id: $followerId})
		MERGE (u2:User {id: $followedId})
		MERGE (u1)-[:FOLLOWS]->(u2)
	`
	_, err := session.Run(ctx, query, map[string]any{"followerId": req.FollowerID, "followedId": req.FollowedID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Followed!"})
}

func GetFollowedUsers(c *gin.Context) {
	idStr := c.Query("userId")
	id, _ := strconv.Atoi(idStr)
	ctx := context.Background()
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	query := `MATCH (u1:User {id: $id})-[:FOLLOWS]->(u2:User) RETURN u2.id as fid`
	res, err := session.Run(ctx, query, map[string]any{"id": id})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	var ids []int
	for res.Next(ctx) {
		if val, ok := res.Record().Get("fid"); ok {
			ids = append(ids, int(val.(int64)))
		}
	}
	c.JSON(200, ids)
}

func GetRecommendations(c *gin.Context) {
	idStr := c.Query("userId")
	id, _ := strconv.Atoi(idStr)
	ctx := context.Background()
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	query := `
		MATCH (me:User {id: $id})-[:FOLLOWS]->(f:User)-[:FOLLOWS]->(fof:User)
		WHERE NOT (me)-[:FOLLOWS]->(fof) AND me.id <> fof.id
		RETURN distinct fof.id as recid LIMIT 5
	`
	res, err := session.Run(ctx, query, map[string]any{"id": id})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	var ids []int
	for res.Next(ctx) {
		if val, ok := res.Record().Get("recid"); ok {
			ids = append(ids, int(val.(int64)))
		}
	}
	if ids == nil { ids = []int{} }
	c.JSON(200, ids)
}

func CheckFollow(c *gin.Context) {
	followerStr := c.Query("followerId")
	followedStr := c.Query("followedId")
	f1, _ := strconv.Atoi(followerStr)
	f2, _ := strconv.Atoi(followedStr)

	ctx := context.Background()
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	query := `
		MATCH (u1:User {id: $f1})-[r:FOLLOWS]->(u2:User {id: $f2})
		RETURN count(r) > 0 as isFollowing
	`
	res, _ := session.Run(ctx, query, map[string]any{"f1": f1, "f2": f2})
	
	isFollowing := false
	if res.Next(ctx) {
		if val, ok := res.Record().Get("isFollowing"); ok {
			isFollowing = val.(bool)
		}
	}
	c.JSON(200, gin.H{"isFollowing": isFollowing})
}

func main() {
	initNeo4j()
	r := gin.Default()
	r.POST("/followers", FollowUser)
	r.GET("/followers", GetFollowedUsers)
	r.GET("/followers/recommendations", GetRecommendations)
	r.GET("/followers/check", CheckFollow) 
	r.Run(":8083")
}