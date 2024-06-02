package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"time"

	"github.com/an1l4/distributedservice-gin/models"
	"github.com/auth0-community/go-auth0"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	jose "gopkg.in/square/go-jose.v2"
)

type AuthHandler struct {
	collection *mongo.Collection
	ctx        context.Context
}

func NewAuthHandler(ctx context.Context, collection *mongo.Collection) *AuthHandler {
	return &AuthHandler{
		collection: collection,
		ctx:        ctx,
	}
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type JWTOutput struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

// swagger:operation POST /signin auth signIn
// Login with username and password
// ---
// produces:
// - application/json
// responses:
//     '200':
//         description: Successful operation
//     '401':
//         description: Invalid credentials
func (handler *AuthHandler) SignInHandler(c *gin.Context) {
	var user models.User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h := sha256.New()
	h.Write([]byte(user.Password))
	hashedPassword := hex.EncodeToString(h.Sum(nil))

	cur := handler.collection.FindOne(handler.ctx, bson.M{
		"username": user.Username,
		"password": hashedPassword,
	})

	if cur.Err() != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// if user.Username != "admin" || user.Password != "password" {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
	// 	return
	// }

	// expirationTime := time.Now().Add(10 * time.Minute)
	// claims := Claims{
	// 	Username: user.Username,
	// 	StandardClaims: jwt.StandardClaims{
	// 		ExpiresAt: expirationTime.Unix(),
	// 	},
	// }

	// token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }

	// jwtOutput := JWTOutput{
	// 	Token:   tokenString,
	// 	Expires: expirationTime,
	// }

	sessionToken := xid.New().String()
	session := sessions.Default(c)
	session.Set("username", user.Username)
	session.Set("token", sessionToken)
	session.Save()

	c.JSON(http.StatusOK, gin.H{"message": "User signed in"})
	//c.JSON(http.StatusOK, jwtOutput)
}

// func (handler *AuthHandler) RefreshHandler(c *gin.Context) {
// 	tokenValue := c.GetHeader("Authorization")

// 	if tokenValue == "" {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
// 		return
// 	}

// 	claims := &Claims{}
// 	tkn, err := jwt.ParseWithClaims(tokenValue, claims, func(token *jwt.Token) (interface{}, error) {
// 		return []byte(os.Getenv("JWT_SECRET")), nil
// 	})
// 	if err != nil {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
// 		return
// 	}
// 	if !tkn.Valid {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
// 		return
// 	}

// 	// if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) > 30*time.Second {
// 	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Token is not expired yet"})
// 	// 	return
// 	// }

// 	if time.Until(time.Unix(claims.ExpiresAt, 0)) > 30*time.Second {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Token is not expired yet"})
// 		return
// 	}

// 	expirationTime := time.Now().Add(5 * time.Minute)
// 	claims.ExpiresAt = expirationTime.Unix()
// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
// 	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	jwtOutput := JWTOutput{
// 		Token:   tokenString,
// 		Expires: expirationTime,
// 	}
// 	c.JSON(http.StatusOK, jwtOutput)
// }

// swagger:operation POST /refresh auth refresh
// Refresh token
// ---
// produces:
// - application/json
// responses:
//     '200':
//         description: Successful operation
//     '401':
//         description: Invalid credentials
func (handler *AuthHandler) RefreshHandler(c *gin.Context) {
	session := sessions.Default(c)
	sessionToken := session.Get("token")
	sessionUser := session.Get("username")
	if sessionToken == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session cookie"})
		return
	}

	sessionToken = xid.New().String()
	session.Set("username", sessionUser.(string))
	session.Set("token", sessionToken)
	session.Save()

	c.JSON(http.StatusOK, gin.H{"message": "New session issued"})
}

// func (handler *AuthHandler) AuthMiddleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		tokenValue := c.GetHeader("Authorization")
// 		claims := &Claims{}

// 		tkn, err := jwt.ParseWithClaims(tokenValue, claims, func(token *jwt.Token) (interface{}, error) {
// 			return []byte(os.Getenv("JWT_SECRET")), nil
// 		})
// 		if err != nil {
// 			c.AbortWithStatus(http.StatusUnauthorized)
// 		}

// 		if tkn == nil || !tkn.Valid {
// 			c.AbortWithStatus(http.StatusUnauthorized)
// 		}

// 		c.Next()

// 	}
// }

// swagger:operation POST /signout auth signOut
// Signing out
// ---
// responses:
//     '200':
//         description: Successful operation
func (handler *AuthHandler) SignOutHandler(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.JSON(http.StatusOK, gin.H{"message": "Signed out..."})
}

// func (handler *AuthHandler) AuthMiddleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		session := sessions.Default(c)
// 		sessionToken := session.Get("token")
// 		if sessionToken == nil {
// 			c.JSON(http.StatusForbidden, gin.H{
// 				"message": "Not logged",
// 			})
// 			c.Abort()
// 		}
// 		c.Next()
// 	}
// }

func (handler *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var auth0Domain = "https://" + os.Getenv("AUTH0_DOMAIN") + "/"

		client := auth0.NewJWKClient(auth0.JWKClientOptions{URI: auth0Domain + ".well-known/jwks.json"}, nil)

		configuration := auth0.NewConfiguration(client, []string{os.Getenv("AUTH0_API_IDENTIFIER")}, auth0Domain, jose.RS256)

		validator := auth0.NewValidator(configuration, nil)

		_, err := validator.ValidateRequest(c.Request)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})

			c.Abort()

			return

		}
		c.Next()
	}

}
