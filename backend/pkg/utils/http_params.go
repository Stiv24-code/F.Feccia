package utils

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// GetUintParam gets a uint parameter from the URL
func GetUintParam(c *fiber.Ctx, name string) (uint, error) {
	idStr := c.Params(name)
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
