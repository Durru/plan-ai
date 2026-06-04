package store

import (
	"database/sql"
	"time"
)

type Scan struct {
	ID              string
	ProjectRoot     string
	GitDetected     bool
	GitBranch       string
	Fingerprint     string
	Summary         string
	CreatedAt       time.Time
	Languages       []ScanLanguage
	Frameworks      []ScanFramework
	PackageManagers []ScanPackageManager
	Dependencies    []ScanDependency
	Files           []ScanFile
}

type ScanLanguage struct {
	ID         string
	ScanID     string
	Language   string
	FilesCount int
	CreatedAt  time.Time
}

type ScanFramework struct {
	ID        string
	ScanID    string
	Framework string
	Evidence  string
	CreatedAt time.Time
}

type ScanPackageManager struct {
	ID        string
	ScanID    string
	Manager   string
	Evidence  string
	CreatedAt time.Time
}

type ScanDependency struct {
	ID        string
	ScanID    string
	Name      string
	Version   string
	Source    string
	CreatedAt time.Time
}

type ScanFile struct {
	ID        string
	ScanID    string
	Path      string
	Kind      string
	SizeBytes int64
	CreatedAt time.Time
}

type ScanSummary struct {
	ID                  string
	ProjectRoot         string
	GitDetected         bool
	GitBranch           string
	Fingerprint         string
	Summary             string
	CreatedAt           time.Time
	LanguageNames       []string
	FrameworkNames      []string
	PackageManagerNames []string
	FileCount           int
}

type ScanRepository struct{ db *sql.DB }

func NewScanRepository(db *sql.DB) ScanRepository { return ScanRepository{db: db} }

func (r ScanRepository) CreateScan(scan Scan) (string, error) {
	scan.ID = ensureID(scan.ID, "scan")
	createdAt, _ := ensureTimestamps(scan.CreatedAt, time.Time{})

	tx, err := r.db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	gitDetected := 0
	if scan.GitDetected {
		gitDetected = 1
	}
	var branch any
	if scan.GitBranch != "" {
		branch = scan.GitBranch
	}

	if _, err := tx.Exec(`INSERT INTO project_scans (id, project_root, git_detected, git_branch, fingerprint, summary, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?)`, scan.ID, scan.ProjectRoot, gitDetected, branch, scan.Fingerprint, scan.Summary, createdAt); err != nil {
		return "", err
	}

	for _, language := range scan.Languages {
		if _, err := tx.Exec(`INSERT INTO project_scan_languages (id, scan_id, language, files_count, created_at)
VALUES (?, ?, ?, ?, ?)`, ensureID(language.ID, "scan_lang"), scan.ID, language.Language, language.FilesCount, createdAt); err != nil {
			return "", err
		}
	}
	for _, framework := range scan.Frameworks {
		if _, err := tx.Exec(`INSERT INTO project_scan_frameworks (id, scan_id, framework, evidence, created_at)
VALUES (?, ?, ?, ?, ?)`, ensureID(framework.ID, "scan_framework"), scan.ID, framework.Framework, framework.Evidence, createdAt); err != nil {
			return "", err
		}
	}
	for _, manager := range scan.PackageManagers {
		if _, err := tx.Exec(`INSERT INTO project_scan_package_managers (id, scan_id, manager, evidence, created_at)
VALUES (?, ?, ?, ?, ?)`, ensureID(manager.ID, "scan_pm"), scan.ID, manager.Manager, manager.Evidence, createdAt); err != nil {
			return "", err
		}
	}
	for _, dependency := range scan.Dependencies {
		var version any
		if dependency.Version != "" {
			version = dependency.Version
		}
		if _, err := tx.Exec(`INSERT INTO project_scan_dependencies (id, scan_id, name, version, source, created_at)
VALUES (?, ?, ?, ?, ?, ?)`, ensureID(dependency.ID, "scan_dep"), scan.ID, dependency.Name, version, dependency.Source, createdAt); err != nil {
			return "", err
		}
	}
	for _, file := range scan.Files {
		if _, err := tx.Exec(`INSERT INTO project_scan_files (id, scan_id, path, kind, size_bytes, created_at)
VALUES (?, ?, ?, ?, ?, ?)`, ensureID(file.ID, "scan_file"), scan.ID, file.Path, file.Kind, file.SizeBytes, createdAt); err != nil {
			return "", err
		}
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}
	return scan.ID, nil
}

func (r ScanRepository) GetLatestScan() (*Scan, error) {
	var scan Scan
	var createdAt string
	var gitDetected int
	var branch sql.NullString
	err := r.db.QueryRow(`SELECT id, project_root, git_detected, git_branch, fingerprint, summary, created_at
FROM project_scans ORDER BY created_at DESC, id DESC LIMIT 1`).Scan(&scan.ID, &scan.ProjectRoot, &gitDetected, &branch, &scan.Fingerprint, &scan.Summary, &createdAt)
	if err != nil {
		return nil, err
	}
	scan.GitDetected = gitDetected == 1
	if branch.Valid {
		scan.GitBranch = branch.String
	}
	scan.CreatedAt = parseTime(createdAt)
	if scan.Languages, err = r.ListLanguages(scan.ID); err != nil {
		return nil, err
	}
	if scan.Frameworks, err = r.ListFrameworks(scan.ID); err != nil {
		return nil, err
	}
	if scan.PackageManagers, err = r.ListPackageManagers(scan.ID); err != nil {
		return nil, err
	}
	if scan.Dependencies, err = r.ListDependencies(scan.ID); err != nil {
		return nil, err
	}
	if scan.Files, err = r.ListFiles(scan.ID); err != nil {
		return nil, err
	}
	return &scan, nil
}

func (r ScanRepository) GetScanSummary() (ScanSummary, error) {
	var summary ScanSummary
	var createdAt string
	var gitDetected int
	var branch sql.NullString
	err := r.db.QueryRow(`SELECT id, project_root, git_detected, git_branch, fingerprint, summary, created_at
FROM project_scans ORDER BY created_at DESC, id DESC LIMIT 1`).Scan(&summary.ID, &summary.ProjectRoot, &gitDetected, &branch, &summary.Fingerprint, &summary.Summary, &createdAt)
	if err != nil {
		return ScanSummary{}, err
	}
	summary.GitDetected = gitDetected == 1
	if branch.Valid {
		summary.GitBranch = branch.String
	}
	summary.CreatedAt = parseTime(createdAt)

	if summary.LanguageNames, err = r.listNames(`SELECT language FROM project_scan_languages WHERE scan_id = ? ORDER BY language`, summary.ID); err != nil {
		return ScanSummary{}, err
	}
	if summary.FrameworkNames, err = r.listNames(`SELECT framework FROM project_scan_frameworks WHERE scan_id = ? ORDER BY framework`, summary.ID); err != nil {
		return ScanSummary{}, err
	}
	if summary.PackageManagerNames, err = r.listNames(`SELECT manager FROM project_scan_package_managers WHERE scan_id = ? ORDER BY manager`, summary.ID); err != nil {
		return ScanSummary{}, err
	}
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM project_scan_files WHERE scan_id = ?`, summary.ID).Scan(&summary.FileCount); err != nil {
		return ScanSummary{}, err
	}
	return summary, nil
}

func (r ScanRepository) listNames(query, scanID string) ([]string, error) {
	rows, err := r.db.Query(query, scanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		out = append(out, value)
	}
	return out, rows.Err()
}

func (r ScanRepository) ListLanguages(scanID string) ([]ScanLanguage, error) {
	rows, err := r.db.Query(`SELECT id, scan_id, language, files_count, created_at FROM project_scan_languages WHERE scan_id = ? ORDER BY language`, scanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ScanLanguage
	for rows.Next() {
		var item ScanLanguage
		var createdAt string
		if err := rows.Scan(&item.ID, &item.ScanID, &item.Language, &item.FilesCount, &createdAt); err != nil {
			return nil, err
		}
		item.CreatedAt = parseTime(createdAt)
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r ScanRepository) ListFrameworks(scanID string) ([]ScanFramework, error) {
	rows, err := r.db.Query(`SELECT id, scan_id, framework, evidence, created_at FROM project_scan_frameworks WHERE scan_id = ? ORDER BY framework`, scanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ScanFramework
	for rows.Next() {
		var item ScanFramework
		var createdAt string
		if err := rows.Scan(&item.ID, &item.ScanID, &item.Framework, &item.Evidence, &createdAt); err != nil {
			return nil, err
		}
		item.CreatedAt = parseTime(createdAt)
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r ScanRepository) ListPackageManagers(scanID string) ([]ScanPackageManager, error) {
	rows, err := r.db.Query(`SELECT id, scan_id, manager, evidence, created_at FROM project_scan_package_managers WHERE scan_id = ? ORDER BY manager`, scanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ScanPackageManager
	for rows.Next() {
		var item ScanPackageManager
		var createdAt string
		if err := rows.Scan(&item.ID, &item.ScanID, &item.Manager, &item.Evidence, &createdAt); err != nil {
			return nil, err
		}
		item.CreatedAt = parseTime(createdAt)
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r ScanRepository) ListDependencies(scanID string) ([]ScanDependency, error) {
	rows, err := r.db.Query(`SELECT id, scan_id, name, version, source, created_at FROM project_scan_dependencies WHERE scan_id = ? ORDER BY source, name`, scanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ScanDependency
	for rows.Next() {
		var item ScanDependency
		var version sql.NullString
		var createdAt string
		if err := rows.Scan(&item.ID, &item.ScanID, &item.Name, &version, &item.Source, &createdAt); err != nil {
			return nil, err
		}
		if version.Valid {
			item.Version = version.String
		}
		item.CreatedAt = parseTime(createdAt)
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r ScanRepository) ListFiles(scanID string) ([]ScanFile, error) {
	rows, err := r.db.Query(`SELECT id, scan_id, path, kind, size_bytes, created_at FROM project_scan_files WHERE scan_id = ? ORDER BY path`, scanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ScanFile
	for rows.Next() {
		var item ScanFile
		var createdAt string
		if err := rows.Scan(&item.ID, &item.ScanID, &item.Path, &item.Kind, &item.SizeBytes, &createdAt); err != nil {
			return nil, err
		}
		item.CreatedAt = parseTime(createdAt)
		out = append(out, item)
	}
	return out, rows.Err()
}
