package api

import (
	"net/http"
	"github.com/gbrlsnchs/jwt"
	"github.com/google/logger"
	"context"
	"fmt"
	"net/url"
)

const ContextJwt = "jwt"

type TokenManager struct {
	secret string
}

func NewTokenManager(secret string) TokenManager {
	return TokenManager{
		secret: secret,
	}
}

func (j TokenManager) CreateToken(secret, androidId string) (string, error) {
	options := &jwt.Options{
		Public: map[string]interface{}{
			"secret":    secret,
			"androidId": androidId,
		},
	}

	token, err := jwt.Sign(jwt.HS256(j.secret), options)
	if err != nil {
		return "", fmt.Errorf("error signing JWT: %s", err)
	}

	return token, nil
}

func (j TokenManager) TokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queryParams, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			return
		}

		token := queryParams.Get("token")
		if token == "" {
			next.ServeHTTP(w, r)
			return
		}

		jwtToken, err := jwt.FromString(token)
		if err != nil {
			logger.Warningf("[1]error when retrieving JWT token: %s", err)
			logger.Warningf("[2]token: %s", token)

			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		err = jwtToken.Verify(jwt.HS256(j.secret))
		if err != nil {
			logger.Warningf("error when retrieving JWT token: %s", err)

			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ContextJwt, jwtToken)

		// Serve next
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
