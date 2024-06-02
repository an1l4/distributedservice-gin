// Recipes API
//
// This is a sample recipes API. You can find out more about the API at https://github.com/ani1l4/distributedservice-gin.
//
//		Schemes: http
//	 Host: localhost:8080
//		BasePath: /
//		Version: 1.0.0
//		Contact: Anila Soman <anila@gmail.com>
//
//		Consumes:
//		- application/json
//
//		Produces:
//		- application/json
//
// swagger:meta
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/an1l4/distributedservice-gin/handlers"
	"github.com/gin-contrib/sessions"
	redisStore "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var recipesHandler *handlers.RecipesHandler
var authHandler *handlers.AuthHandler

func init() {
	/*recipes = make([]Recipe, 0)
	data, err := os.ReadFile("recipes.json")
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(data, &recipes)

	ctx = context.Background()
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))

	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB!")

	var listOfRecipes []interface{}

	for _, recipe := range recipes {
		listOfRecipes = append(listOfRecipes, recipe)
	}

	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")

	insertManyResult, err := collection.InsertMany(ctx, listOfRecipes)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Inserted recipes: ", len(insertManyResult.InsertedIDs))*/

	users := map[string]string{
		"admin": "fCRmh4Q2J7Rseqkz",
		"packt": "RE4zfHB35VPtTkbT",
		"anila": "L3nSFRcZzNQ67bcc",
	}
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))

	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB!")

	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	status := redisClient.Ping()
	fmt.Println(status)

	recipesHandler = handlers.NewRecipesHandler(ctx, collection, redisClient)

	collectionUsers := client.Database(os.Getenv("MONGO_DATABASE")).Collection("users")

	authHandler = handlers.NewAuthHandler(ctx, collectionUsers)

	h := sha256.New()

	for username, password := range users {
		h.Reset()
		h.Write([]byte(password))
		hashedPassword := hex.EncodeToString(h.Sum(nil))
		_, err := collectionUsers.InsertOne(ctx, bson.M{
			"username": username,
			"password": hashedPassword,
		})
		if err != nil {
			log.Printf("Error inserting user %s: %v\n", username, err)
		}
	}

}

// func AuthMiddleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		if c.GetHeader("X-API-KEY") != os.Getenv("X_API_KEY") {
// 			c.AbortWithStatus(401)
// 		}

// 		c.Next()
// 	}
// }

func main() {
	router := gin.Default()

	store, _ := redisStore.NewStore(10, "tcp", "localhost:6379", "", []byte("secret"))
	router.Use(sessions.Sessions("recipes_api", store))

	router.GET("/recipes", recipesHandler.ListRecipesHandler)
	router.POST("/signin", authHandler.SignInHandler)
	router.POST("/refresh", authHandler.RefreshHandler)
	router.POST("/signout", authHandler.SignOutHandler)

	authorized := router.Group("/")
	authorized.Use(authHandler.AuthMiddleware())
	{
		authorized.POST("/recipes", recipesHandler.NewRecipeHandler)
		authorized.PUT("/recipes/:id", recipesHandler.UpdateRecipehandler)
		authorized.DELETE("/recipes/:id", recipesHandler.DeleteRecipeHandler)
		//router.GET("/recipes/search", recipesHandler.SearchRecipesHandler)
		authorized.GET("/recipes/:id", recipesHandler.GetOneRecipeHandler)

	}
	//router.RunTLS(":443","certs/localhost.crt","certs/localhost.key")
	router.Run()

}
