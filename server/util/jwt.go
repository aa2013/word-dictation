package util

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UID      int
	Account  string
	UserName string
	JTI      string
}

const jwtIssuer = "word-dictation"

func SignJWT(secret string, cl Claims, ttl time.Duration) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"iss":  jwtIssuer,
		"sub":  cl.UID,
		"uid":  cl.UID,
		"acc":  cl.Account,
		"name": cl.UserName,
		"jti":  cl.JTI,
		"iat":  now.Unix(),
		"exp":  now.Add(ttl).Unix(),
	})
	return token.SignedString([]byte(secret))
}

func ParseJWT(secret, tokenStr string) (*Claims, error) {
	parsed, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("非法签名算法")
		}
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS512.Alg()}))
	if err != nil || !parsed.Valid {
		return nil, errors.New("token 无效")
	}
	mc, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("token 载荷无效")
	}
	uidF, ok := mc["uid"].(float64)
	if !ok {
		return nil, errors.New("token 载荷缺少 uid")
	}
	jti, _ := mc["jti"].(string)
	acc, _ := mc["acc"].(string)
	name, _ := mc["name"].(string)
	return &Claims{
		UID:      int(uidF),
		Account:  acc,
		UserName: name,
		JTI:      jti,
	}, nil
}
