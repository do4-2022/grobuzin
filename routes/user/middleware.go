package user

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Middleware that requires a valid JWT token to be present in the Authorization header, meaning the user has an account to modify the functions they own.
// You want to add this middleware to a group (router.Group()) of routes that require authentication.
// JWTSecret is the secret key used to sign the JWT tokens.
// This middleware will set the "userID" and "username" keys in the gin.Context.
//
// # Example:
//
//	func ConfigureRoutes(router *gin.Engine, db *gorm.DB, jwtSecret string) {
//		controller := Controller{DB: db, JWTSecret: jwtSecret}
//		group := router.Group("/my-prefix")
//		group.Use(RequireAuth(jwtSecret))
//		group.POST("/", controller.postFunction)
//		group.GET("/test", controller.testFunction)
//	}
func RequireAuth(JWTSecret string) gin.HandlerFunc {

	parsefunc := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(JWTSecret), nil
	}

	return func(c *gin.Context) {

		const BearerSchema = "Bearer "
		authHeader := c.GetHeader("Authorization")

		if len(authHeader) <= len(BearerSchema) {
			c.AbortWithStatus(401)
			return
		}

		token := authHeader[len(BearerSchema):]

		parsedToken, err := jwt.ParseWithClaims(token, &Claims{}, parsefunc)

		if err != nil {
			log.Println(err)
			c.AbortWithStatus(401)
			return
		}

		if !parsedToken.Valid {
			c.AbortWithStatus(401)
			return
		}

		claims, ok := parsedToken.Claims.(*Claims)
		if !ok {
			c.AbortWithStatus(401)
			return
		}

		c.Set("userID", claims.ID)
		c.Set("username", claims.Subject)

		c.Next()
	}
}
