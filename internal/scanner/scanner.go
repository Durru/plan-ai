package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Scanner is a deterministic, rule-based project analyzer.
type Scanner struct {
	// MaxFileSize bounds the size of any single file the scanner indexes.
	MaxFileSize int64
	// SnippetSize bounds the size of file content the scanner reads into
	// memory to feed framework detectors that need a content preview
	// (e.g. pubspec.yaml flutter: marker).
	SnippetSize int64
	// Walker is the function used to enumerate files. Override for tests.
	Walker func(root string) ([]ScannedFile, error)
	// ReadFile, when set, is used to read snippets. Override for tests.
	ReadFile func(path string) (string, error)
	// Now is the time source used to populate Result.CreatedAt.
	Now func() time.Time
}

// Default returns a Scanner configured with the standard production values.
func Default() *Scanner {
	return &Scanner{
		MaxFileSize: DefaultMaxFileSize,
		SnippetSize: 64 * 1024,
		Walker:      WalkProject,
		ReadFile:    readFileSnippet,
		Now:         func() time.Time { return time.Now().UTC() },
	}
}

func readFileSnippet(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if info.Size() > 64*1024 {
		f, err := os.Open(path)
		if err != nil {
			return "", err
		}
		defer f.Close()
		buf := make([]byte, 64*1024)
		n, _ := f.Read(buf)
		return string(buf[:n]), nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Scan runs the full deterministic project analysis and returns the result.
// It is safe to call against any directory; missing or unrecognized files
// are treated as "not present" rather than errors.
func (s *Scanner) Scan(root string) (*Result, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	gitDetected, gitBranch := DetectGit(absRoot)
	files, err := s.Walker(absRoot)
	if err != nil {
		return nil, fmt.Errorf("walk project: %w", err)
	}

	fileMap := buildFileMap(files)
	otherFiles := map[string]string{}
	if s.ReadFile != nil {
		for _, candidate := range []string{"pubspec.yaml"} {
			if !fileMap[candidate] {
				continue
			}
			content, err := s.ReadFile(filepath.Join(absRoot, candidate))
			if err != nil {
				continue
			}
			otherFiles[candidate] = content
		}
	}

	packageJSONDeps, err := loadPackageJSONDeps(absRoot, fileMap, s.ReadFile)
	if err != nil {
		return nil, err
	}
	goModDeps, err := loadGoModDeps(absRoot, fileMap, s.ReadFile)
	if err != nil {
		return nil, err
	}

	languages := DetectLanguages(files)
	packageManagers := DetectPackageManagers(fileMap)
	frameworks := DetectFrameworks(fileMap, packageJSONDeps, goModDeps, otherFiles)
	dependencies := AllDependencies(absRoot)

	fingerprint, err := Fingerprint(absRoot, files, packageManagers)
	if err != nil {
		return nil, fmt.Errorf("fingerprint: %w", err)
	}

	now := s.Now()
	if now.IsZero() {
		now = time.Now().UTC()
	}

	result := &Result{
		ProjectRoot:     absRoot,
		GitDetected:     gitDetected,
		GitBranch:       gitBranch,
		Fingerprint:     fingerprint,
		Languages:       languages,
		Frameworks:      frameworks,
		PackageManagers: packageManagers,
		Dependencies:    dependencies,
		Files:           files,
		CreatedAt:       now,
	}
	result.Summary = buildSummary(result)
	return result, nil
}

func buildFileMap(files []ScannedFile) map[string]bool {
	out := make(map[string]bool, len(files))
	for _, f := range files {
		out[f.Path] = true
		out[filepath.Base(f.Path)] = true
	}
	return out
}

func loadPackageJSONDeps(root string, fileMap map[string]bool, read func(string) (string, error)) (map[string]string, error) {
	if !fileMap["package.json"] || read == nil {
		return map[string]string{}, nil
	}
	content, err := read(filepath.Join(root, "package.json"))
	if err != nil {
		return nil, nil
	}
	return parsePackageJSONDeps(content), nil
}

func loadGoModDeps(root string, fileMap map[string]bool, read func(string) (string, error)) (map[string]string, error) {
	if !fileMap["go.mod"] || read == nil {
		return map[string]string{}, nil
	}
	content, err := read(filepath.Join(root, "go.mod"))
	if err != nil {
		return nil, nil
	}
	return parseGoModDeps(content), nil
}

// buildSummary composes the human-readable single-line summary of a scan.
// The order of the components is fixed so downstream tests and the CLI
// output are stable.
func buildSummary(r *Result) string {
	frameworks := make([]string, 0, len(r.Frameworks))
	for _, f := range r.Frameworks {
		frameworks = append(frameworks, f.Name)
	}
	sort.Strings(frameworks)
	pms := make([]string, 0, len(r.PackageManagers))
	for _, pm := range r.PackageManagers {
		pms = append(pms, pm.Name)
	}
	sort.Strings(pms)

	parts := []string{
		fmt.Sprintf("%d files", len(r.Files)),
		fmt.Sprintf("%d languages", len(r.Languages)),
		fmt.Sprintf("%d frameworks", len(frameworks)),
		fmt.Sprintf("%d dependencies", len(r.Dependencies)),
	}
	if len(frameworks) > 0 {
		parts = append(parts, "frameworks="+strings.Join(frameworks, ","))
	}
	if len(pms) > 0 {
		parts = append(parts, "package_managers="+strings.Join(pms, ","))
	}
	return strings.Join(parts, ", ")
}
