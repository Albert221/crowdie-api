package api

import (
	"net/http"
	"strings"
	"github.com/gbrlsnchs/jwt"
	"github.com/google/logger"
	"context"
	"fmt"
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
		authHeader := r.Header.Get("Authentication")

		if authHeader == "" {
			// If Authentication header isn't provided, proceed to next handler
			next.ServeHTTP(w, r)
			return
		}

		split := strings.Split(authHeader, " ")
		if len(split) != 2 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		jwtToken, err := jwt.FromString(split[1])
		if err != nil {
			logger.Warningf("error when retrieving JWT token: %s", err)

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
