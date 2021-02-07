package JwtAuth

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var (
	TokenExpired     error = errors.New("Token is expired")
	TokenNotValidYet error = errors.New("Token not active yet")
	TokenMalformed   error = errors.New("That's not even a token")
	TokenInvalid     error = errors.New("Couldn't handle this token:")
)

type JWT struct {
	SigningKey []byte
}

type CustomClaims struct {
	Id       int
	Phone    string
	Avatar   string
	Nickname string
	jwt.StandardClaims
}

// func DeCode(tokenString string) (*CustomClaims, error) {
// 	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
// 		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
// 		}
// 		return []byte(SigningKey), nil
// 	})
// 	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
// 		return claims, nil
// 	} else {
// 		return nil, err
// 	}
// }

//创建token
func (j *JWT) CreateToken(claims CustomClaims) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	res, err := token.SignedString(j.SigningKey)
	return res, err
}

//解析token
func (j *JWT) ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Panicln("unexpected signing method")
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.SigningKey, nil
	})

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, TokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				// Token is expired
				return nil, TokenExpired
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, TokenNotValidYet
			} else {
				return nil, TokenInvalid
			}
		}
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, TokenInvalid
}

//更新token
func (j *JWT) RefreshToken(tokenString string) (string, error) {
	jwt.TimeFunc = func() time.Time {
		return time.Unix(0, 0)
	}

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		jwt.TimeFunc = time.Now
		claims.StandardClaims.ExpiresAt = time.Now().Add(1 * time.Hour).Unix()
		return j.CreateToken(*claims)
	}

	return "", TokenInvalid
}

// func Login(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
// 	generateToken(w)
// }

// func generateToken(w http.ResponseWriter) {
// 	j := &JWT{[]byte("man")}
// 	claims := CustomClaims{
// 		"1", "Jaya", "123456", jwt.StandardClaims{
// 			NotBefore: int64(time.Now().Unix() - 1000),
// 			ExpiresAt: int64(time.Now().Unix() + 3600),
// 			Issuer:    "man",
// 		},
// 	}

// 	token, err := j.CreateToken(claims)
// 	if err != nil {
// 		io.WriteString(w, "it is wrong")
// 	}

// 	io.WriteString(w, token)
// }
