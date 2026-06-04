package research

import "strings"

// BuildSearchQuery builds a LIKE query for searching research entries.
// It wraps the query in percent signs and lowercases it.
func BuildSearchQuery(query string) string {
	return "%" + strings.ToLower(strings.TrimSpace(query)) + "%"
}
