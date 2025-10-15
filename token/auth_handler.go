package token

import (
	"context"
	"github.com/gofiber/fiber/v2"
)

func New(principal Principal) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenStr := GetToken(c)
		if tokenStr != "" {
			authentication := ReadToken(tokenStr, principal)
			c.SetUserContext(context.WithValue(c.UserContext(), "userInfo", authentication))
		}
		return c.Next()
	}
}
