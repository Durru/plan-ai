package ingestion

import "strings"

func Normalize(input string) NormalizedInput {
	text := strings.ReplaceAll(input, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	lines := strings.Split(text, "\n")
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.Join(strings.Fields(strings.TrimSpace(line)), " ")
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			line = "- " + strings.TrimSpace(line[2:])
		}
		cleaned = append(cleaned, line)
	}
	text = collapseBlankLines(strings.TrimSpace(strings.Join(cleaned, "\n")))
	blocks := splitBlocks(text)
	return NormalizedInput{Content: text, Blocks: blocks, ListItems: listItems(cleaned)}
}

func collapseBlankLines(text string) string {
	for strings.Contains(text, "\n\n\n") {
		text = strings.ReplaceAll(text, "\n\n\n", "\n\n")
	}
	return text
}

func splitBlocks(text string) []string {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	parts := strings.Split(text, "\n\n")
	blocks := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			blocks = append(blocks, part)
		}
	}
	return blocks
}

func listItems(lines []string) []string {
	var out []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			out = append(out, strings.TrimSpace(trimmed[2:]))
		}
	}
	return out
}
