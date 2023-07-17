package middleware

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"bitbucket.org/pensarmais/cycleforlisbon/src/util/httputil"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

const (
	// Key used to store the auth token claims in the gin context.
	TokenClaimsKey = "token_claims"

	authHeaderKey  = "authorization"
	authTypeBearer = "bearer"

	connectProviderTimeout = 2 * time.Minute
)

var (
	errNoAuthorizationHeader = errors.New("missing authorization header")
	errNoOidcProvider        = errors.New("oidc provider not initialized")
)

type Claims struct {
	Issuer   string `json:"iss"`
	ClientID string `json:"aud"`
	Sub      string `json:"sub"`
	Email    string `json:"email"`
	Verified bool   `json:"email_verified"`
	Name     string `json:"name"`
}

// connectProvider waits for the server to become available to initialize the
// provider.
func connectProvider(
	ctx context.Context,
	providerCtx context.Context,
	issuerUrl string,
) (*oidc.Provider, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for server")
		default:
			provider, err := oidc.NewProvider(providerCtx, issuerUrl)
			if err == nil {
				return provider, nil
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func initProvider(
	ctx context.Context,
	providerCtx context.Context,
	issuerUrl string,
) (*oidc.Provider, error) {
	ctx, cancel := context.WithTimeout(ctx, connectProviderTimeout)
	defer cancel()
	return connectProvider(ctx, providerCtx, issuerUrl)
}

// Creates the middleware to validate oidc tokens for the given issuer.
// Returns a channel to signal when the oidc provider is initialized (mostly
// for testing purposes).
func Auth(
	issuerUrl string,
	clientIds []string,
) (gin.HandlerFunc, <-chan struct{}) {
	ctx := context.Background()
	var provider *oidc.Provider
	ch := make(chan struct{}, 1)
	go func() {
		prov, err := initProvider(context.Background(), ctx, issuerUrl)
		if err != nil {
			msg := fmt.Errorf("failed to initialize oidc provider: %v", err)
			sentry.CaptureException(msg)
			sentry.Flush(time.Second * 5)
			panic(msg)
		}
		provider = prov
		close(ch)
		log.Println("oidc provider successfully initialized")
	}()

	return func(c *gin.Context) {
		if provider == nil {
			c.AbortWithStatusJSON(
				http.StatusServiceUnavailable,
				ApiError{httputil.NewError(
					httputil.InternalServerError,
					errNoOidcProvider,
				)},
			)
			return
		}

		idTokenVerifier := provider.Verifier(&oidc.Config{
			SkipClientIDCheck: true,
		})

		bearerToken, err := getBearerToken(c)
		if err != nil {
			abortWithAuthError(httputil.MissingAuthToken, err, c)
			return
		}

		idToken, err := idTokenVerifier.Verify(ctx, bearerToken)
		if err != nil {
			abortWithAuthError(httputil.InvalidAuthToken, err, c)
			return
		}

		claims := Claims{}
		if err := idToken.Claims(&claims); err != nil {
			abortWithAuthError(httputil.InvalidAuthToken, err, c)
			return
		}

		if !knownClientId(claims.ClientID, clientIds) {
			abortWithAuthError(
				httputil.InvalidAuthToken,
				errors.New("unkown client id"),
				c,
			)
			return
		}

		c.Set(TokenClaimsKey, claims)
		c.Next()
	}, ch
}

func getBearerToken(c *gin.Context) (string, error) {
	authHeader := c.GetHeader(authHeaderKey)
	fields := strings.Fields(authHeader)
	if len(authHeader) == 0 || len(fields) < 2 {
		return "", errNoAuthorizationHeader
	}

	if strings.ToLower(fields[0]) != authTypeBearer {
		return "", fmt.Errorf("auth type '%s' not supported", fields[0])
	}

	return fields[1], nil
}

func knownClientId(clientId string, clientIds []string) bool {
	for _, id := range clientIds {
		if id == clientId {
			return true
		}
	}
	return false
}

func abortWithAuthError(code httputil.ErrorCode, err error, c *gin.Context) {
	c.AbortWithStatusJSON(
		http.StatusUnauthorized,
		ApiError{httputil.NewError(code, err)},
	)
}
