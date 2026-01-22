package handler

import (
	"context"

	jwtutil "github.com/JuniorCrafter/fooddelivery/internal/platform/jwt"
)

func contextWithClaims(ctx context.Context, c *jwtutil.Claims) context.Context {
	return context.WithValue(ctx, ctxClaimsKey, c)
}

func claimsFromContext(ctx context.Context) *jwtutil.Claims {
	v := ctx.Value(ctxClaimsKey)
	if v == nil {
		return &jwtutil.Claims{}
	}
	c, _ := v.(*jwtutil.Claims)
	return c
}
