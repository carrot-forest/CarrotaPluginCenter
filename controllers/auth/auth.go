package auth

import (
	"carrota-plugin-center/utils"
	"carrota-plugin-center/utils/logs"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const (
	tokenHeaderName               = "Authorization"
	accessTokenExpirationDuration = 36500 * 24 * time.Hour
	UserIdLength                  = 32
)

var jwtAccessSecretKey string

type Authorization struct {
	AccessSecretKey string `config:"secret-key"`
}

type Claims struct {
	ID          string `json:"id"` // 随机字符串作为用户唯一标识符，避免 JWT token 签名字符串重复
	Progress    uint32 `json:"progress"`
	SubProgress uint32 `json:"sub_progress"`
	jwt.StandardClaims
}

func InitAuthorization(a Authorization) error {
	if a.AccessSecretKey == "" {
		return errors.New("access-secret-key is empty")
	}
	jwtAccessSecretKey = a.AccessSecretKey
	return nil
}

func GetJwtAccessSecretKey() string {
	return jwtAccessSecretKey
}

func GenerateAccessToken(id string, isGenerateNewID bool, progress uint32, subProgress uint32) (token string, expireAt time.Time, err error) {
	expireAt = time.Now().Add(accessTokenExpirationDuration)

	if isGenerateNewID {
		id = utils.RandSeq(UserIdLength)
	}

	token, err = generateToken(id, progress, subProgress, expireAt, GetJwtAccessSecretKey())
	return token, expireAt, err
}

func generateToken(id string, progress uint32, subProgress uint32, expireAt time.Time, secretKey string) (tokenString string, err error) {
	claims := &Claims{
		ID:          id,
		Progress:    progress,
		SubProgress: subProgress,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireAt.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err = token.SignedString([]byte(secretKey))
	if err != nil {
		logs.Warn("Generate token failed.", zap.Error(err))
		return "", err
	}
	return tokenString, err
}

func GetClaimsFromHeader(c echo.Context) (claims Claims, err error) {
	bearerToken := strings.Split(c.Request().Header.Get(tokenHeaderName), " ")
	if len(bearerToken) < 2 {
		return Claims{}, errors.New("invalid header")
	}
	if bearerToken[0] != "Bearer" {
		return Claims{}, errors.New("invalid header")
	}

	tokenString := bearerToken[1]
	claims = Claims{}
	_, err = jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(GetJwtAccessSecretKey()), nil
	})
	if err != nil {
		return Claims{}, err
	}

	return claims, nil
}
