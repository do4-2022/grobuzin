package user

import (
	"github.com/do4-2022/grobuzin/database"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	jwt.RegisteredClaims
	ID uint `json:"id"`
}

func (c *Controller) createJWT(user database.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		Claims{
			jwt.RegisteredClaims{
				Subject: user.Username,
			},
			user.ID,
		},
	)

	return token.SignedString([]byte(c.JWTSecret))

}
