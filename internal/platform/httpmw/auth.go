package httpmw

import (
	"context"
	"net/http"
	"strings"

	jwtutil "github.com/JuniorCrafter/fooddelivery/internal/platform/jwt"
)

type claimsKey struct{}

// Auth проверяет Bearer JWT, парсит и кладёт Claims в context.
// Возвращает 401, если токена нет/он невалидный.
func Auth(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hdr := r.Header.Get("Authorization")
			if hdr == "" || !strings.HasPrefix(hdr, "Bearer ") {
				http.Error(w, "missing token", http.StatusUnauthorized)
				return
			}

			tok := strings.TrimPrefix(hdr, "Bearer ")
			claims, err := jwtutil.ParseHS256(secret, tok)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Claims возвращает Claims из контекста (если Auth уже отработал).
func Claims(ctx context.Context) (*jwtutil.Claims, bool) {
	v := ctx.Value(claimsKey{})
	if v == nil {
		return nil, false
	}
	c, ok := v.(*jwtutil.Claims)
	return c, ok
}

// RequireRole требует одну из ролей. Если роль не подходит — 403.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := Claims(r.Context())
			if !ok {
				http.Error(w, "missing claims", http.StatusUnauthorized)
				return
			}
			if _, ok := allowed[claims.Role]; !ok {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
