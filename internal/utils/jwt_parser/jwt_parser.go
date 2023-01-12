package jwtparser

import (
	"sync"

	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
)

type identifire struct {
	// http(S)://host:port/events?stream=<stream_id>
	AddrSSE string
	ClientID      string
}

var once sync.Once
var i identifire

func Parse(log *logrus.Logger, jwtToken string) identifire {

	once.Do(func() {
		token, _, err := new(jwt.Parser).ParseUnverified(jwtToken, jwt.MapClaims{})
		if err != nil {
			log.Fatal(err)
		}

		// TODO: еще нужен эндпонит для отправка снепшотов
		claims, ok := token.Claims.(jwt.MapClaims)
		addrSSE := claims["AddrSSE"]
		clientID := claims["ClientID"]
		// TODO: добавить еше проверку времени работы токена 
		if ok && clientID != nil && addrSSE != nil {
			i = identifire{
				AddrSSE: addrSSE.(string),
				ClientID:      clientID.(string),
			}
		} else {
			log.Fatal("invalid token")
		}
	})
	return i
}
