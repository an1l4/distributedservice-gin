package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/an1l4/distributedservice-gin/models"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RecipesHandler struct {
	collection  *mongo.Collection
	ctx         context.Context
	redisClient *redis.Client
}

func NewRecipesHandler(ctx context.Context, collection *mongo.Collection, redisClient *redis.Client) *RecipesHandler {
	return &RecipesHandler{
		collection:  collection,
		ctx:         ctx,
		redisClient: redisClient,
	}
}

// swagger:operation GET /recipes recipes listRecipes
// Returns list of recipes
// ---
// produces:
// - application/json
// responses:
//
//	'200':
//	    description: Successful operation
func (handler *RecipesHandler) ListRecipesHandler(c *gin.Context) {

	val, err := handler.redisClient.Get("recipes").Result()
	if err == redis.Nil {
		log.Println("Request to MongoDB")

		cur, err := handler.collection.Find(handler.ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		defer cur.Close(handler.ctx)

		recipes := make([]models.Recipe, 0)
		for cur.Next(handler.ctx) {
			var recipe models.Recipe
			cur.Decode(&recipe)
			recipes = append(recipes, recipe)
		}

		data, _ := json.Marshal(recipes)
		handler.redisClient.Set("recipes", data, 0)
		c.JSON(http.StatusOK, recipes)
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return

	} else {
		log.Println("Request to Redis")
		recipes := make([]models.Recipe, 0)
		json.Unmarshal([]byte(val), &recipes)
		c.JSON(http.StatusOK, recipes)
	}
}

// swagger:operation POST /recipes recipes newRecipe
// Create a new recipe
// ---
// produces:
// - application/json
// responses:
//
//	'200':
//	    description: Successful operation
//	'400':
//	    description: Invalid input
func (handler *RecipesHandler) NewRecipeHandler(c *gin.Context) {
	/*if c.GetHeader("X-API-KEY") != os.Getenv("X_API_KEY") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "API key not provided or invalid"})
		return
	}*/
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//recipe.ID = xid.New().String()
	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()
	_, err := handler.collection.InsertOne(handler.ctx, recipe)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while inserting a new recipe"})
		return
	}
	log.Println("Remove data from redis")
	handler.redisClient.Del("recipes")
	c.JSON(http.StatusOK, recipe)

}

// swagger:operation PUT /recipes/{id} recipes updateRecipe
// Update an existing recipe
// ---
// parameters:
//   - name: id
//     in: path
//     description: ID of the recipe
//     required: true
//     type: string
//
// produces:
// - application/json
// responses:
//
//	'200':
//	    description: Successful operation
//	'400':
//	    description: Invalid input
//	'404':
//	    description: Invalid recipe ID
func (handler *RecipesHandler) UpdateRecipehandler(c *gin.Context) {
	id := c.Param("id")
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// index := -1
	// for i := 0; i < len(recipes); i++ {
	// 	if recipes[i].ID == id {
	// 		index = i
	// 	}
	// }

	// if index == -1 {
	// 	c.JSON(http.StatusNotFound, gin.H{"error": "Recipe Not Found"})
	// 	return
	// }

	// recipes[index] = recipe

	objectId, _ := primitive.ObjectIDFromHex(id)

	_, err := handler.collection.UpdateOne(handler.ctx, bson.M{
		"_id": objectId,
	}, bson.D{{"$set", bson.D{
		{"name", recipe.Name},
		{"instructions", recipe.Instructions},
		{"ingredients", recipe.Ingredients},
		{"tags", recipe.Tags},
	}}})

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Println("Remove data from redis")
	handler.redisClient.Del("recipes")
	c.JSON(http.StatusOK, gin.H{"message": "Recipe has been updated"})
}

// swagger:operation DELETE /recipes/{id} recipes deleteRecipe
// Delete an existing recipe
// ---
// produces:
// - application/json
// parameters:
//   - name: id
//     in: path
//     description: ID of the recipe
//     required: true
//     type: string
//
// responses:
//
//	'200':
//	    description: Successful operation
//	'404':
//	    description: Invalid recipe ID
func (handler *RecipesHandler) DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)

	// index := -1
	// for i := 0; i < len(recipes); i++ {
	// 	if recipes[i].ID == id {
	// 		index = i
	// 	}
	// }

	// if index == -1 {
	// 	c.JSON(http.StatusNotFound, gin.H{"error": "Recipe Not Found"})
	// 	return
	// }
	// recipes = append(recipes[:index], recipes[index+1:]...)

	_, err := handler.collection.DeleteOne(handler.ctx, bson.M{"_id": objectId})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Println("Remove data from Redis")
	handler.redisClient.Del("recipes")

	c.JSON(http.StatusOK, gin.H{"message": "Recipe has been Deleted"})
}

// swagger:operation GET /recipes/search recipes findRecipe
// Search recipes based on tags
// ---
// produces:
// - application/json
// parameters:
//   - name: tag
//     in: query
//     description: recipe tag
//     required: true
//     type: string
//
// responses:
//
//	'200':
//	    description: Successful operation
// func (handler *RecipesHandler) SearchRecipesHandler(c *gin.Context) {
// 	tag := c.Query("tag")
// 	listOfRecipes := make([]Recipe, 0)

// 	for i := 0; i < len(recipes); i++ {
// 		found := false
// 		for _, t := range recipes[i].Tags {
// 			if strings.EqualFold(t, tag) {
// 				found = true
// 			}
// 		}
// 		if found {
// 			listOfRecipes = append(listOfRecipes, recipes[i])
// 		}
// 	}
// 	c.JSON(http.StatusOK, listOfRecipes)

// }

// swagger:operation GET /recipes/{id} recipes oneRecipe
// Get one recipe
// ---
// produces:
// - application/json
// parameters:
//   - name: id
//     in: path
//     description: ID of the recipe
//     required: true
//     type: string
//
// responses:
//
//	'200':
//	    description: Successful operation
//	'404':
//	    description: Invalid recipe ID
func (handler *RecipesHandler) GetOneRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	// for i := 0; i < len(recipes); i++ {
	// 	if recipes[i].ID == id {
	// 		c.JSON(http.StatusOK, recipes[i])
	// 		return
	// 	}
	// }
	objectId, _ := primitive.ObjectIDFromHex(id)
	cur := handler.collection.FindOne(handler.ctx, bson.M{"_id": objectId})

	var recipe models.Recipe
	err := cur.Decode(&recipe)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, recipe)
}
