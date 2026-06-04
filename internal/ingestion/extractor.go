package ingestion

import "strings"

func ExtractReferences(text string) []string {
	fields := strings.Fields(text)
	refs := make([]string, 0)
	for _, field := range fields {
		field = strings.Trim(field, " ,;()[]{}<>\"'")
		lower := strings.ToLower(field)
		if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") || strings.HasSuffix(lower, ".png") || strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg") || strings.HasSuffix(lower, ".gif") || strings.HasPrefix(lower, "./") || strings.HasPrefix(lower, "../") {
			refs = append(refs, field)
		}
	}
	return refs
}
