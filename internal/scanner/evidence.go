package scanner

import (
	"bufio"
	"encoding/json"
	"strings"
)

// parsePackageJSONDeps extracts dependencies and devDependencies from a
// package.json string. The result is a single merged map; collisions are
// kept under their first encountered name.
func parsePackageJSONDeps(content string) map[string]string {
	out := map[string]string{}
	var manifest struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal([]byte(content), &manifest); err != nil {
		return out
	}
	for name, version := range manifest.Dependencies {
		out[name] = version
	}
	for name, version := range manifest.DevDependencies {
		out[name] = version
	}
	return out
}

// parseGoModDeps returns the union of require entries (including indirect
// annotations) found in a go.mod string.
func parseGoModDeps(content string) map[string]string {
	out := map[string]string{}
	inRequire := false
	scanner := bufio.NewScanner(strings.NewReader(content))
	scanner.Buffer(make([]byte, 0, 64*1024), 1<<20)
	for scanner.Scan() {
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" || strings.HasPrefix(raw, "//") {
			continue
		}
		if strings.HasPrefix(raw, "require ") {
			if strings.HasSuffix(raw, ")") {
				inRequire = false
				entry := strings.TrimSpace(strings.TrimPrefix(raw, "require "))
				if entry == ")" || entry == "(" {
					continue
				}
				dep, ok := parseRequireLine(entry)
				if ok {
					out[dep.Name] = dep.Version
				}
				continue
			}
			inRequire = true
			entry := strings.TrimSpace(strings.TrimPrefix(raw, "require "))
			if entry == "" {
				continue
			}
			dep, ok := parseRequireLine(entry)
			if ok {
				out[dep.Name] = dep.Version
			}
			continue
		}
		if inRequire && raw == ")" {
			inRequire = false
			continue
		}
		if inRequire {
			dep, ok := parseRequireLine(raw)
			if ok {
				out[dep.Name] = dep.Version
			}
		}
	}
	return out
}
