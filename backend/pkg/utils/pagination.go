package utils

import "github.com/gofiber/fiber/v2"

// PageParams holds parsed `page`/`limit` query params for offset-based list
// pagination, clamped to sane bounds so a malicious/careless caller can't
// force an unbounded or negative-offset query.
type PageParams struct {
	Page  int
	Limit int
}

// ParsePageParams reads page/limit from the query string (default 1/20, max
// limit 100) — same defaults as AdminService.ListUsers, kept consistent
// across every paginated list endpoint.
func ParsePageParams(c *fiber.Ctx) PageParams {
	page := c.QueryInt("page", 1)
	if page < 1 {
		page = 1
	}
	limit := c.QueryInt("limit", 20)
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return PageParams{Page: page, Limit: limit}
}

// Offset returns the GORM offset for these page params.
func (p PageParams) Offset() int {
	return (p.Page - 1) * p.Limit
}
