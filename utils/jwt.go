package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"
)

//var stSigningKey = []byte(viper.GetString("jwt.signingKey"))

// getSigningKey 获取签名密钥
// 建议：不要用全局变量直接初始化，防止 Viper 还没加载完就读取
func getSigningKey() []byte {
	return []byte(viper.GetString("jwt.signingKey"))
}

type JwtCustClaims struct {
	ID   uint64
	Name string
	Role int
	jwt.RegisteredClaims
}

func GenerateToken(id uint64, name string, role int) (string, error) {
	iJwtCustClaims := JwtCustClaims{
		ID:   id,
		Name: name,
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(viper.GetDuration("jwt.tokenExpire") * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "Token",
			Issuer:    "TomatoBlog",
		},
	}

	fmt.Println("ExpiresAt:", jwt.NewNumericDate(time.Now().Add(viper.GetDuration("jwt.tokenExpire")*time.Minute)))
	fmt.Println("IssuedAt:", jwt.NewNumericDate(time.Now()))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, iJwtCustClaims)
	return token.SignedString(getSigningKey())
}

func ParseToken(tokenStr string) (*JwtCustClaims, error) {
	iJwtCustClaims := JwtCustClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, &iJwtCustClaims, func(token *jwt.Token) (interface{}, error) {
		return getSigningKey(), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JwtCustClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func IsTokenValid(tokenStr string) bool {
	_, err := ParseToken(tokenStr)
	return err == nil

}
