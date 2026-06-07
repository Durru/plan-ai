package repositories

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/Durru/plan-ai/internal/domain"
)

type rowScanner interface{ Scan(dest ...any) error }

func now() string { return time.Now().UTC().Format(time.RFC3339) }

func ensureID(id, prefix string) string {
	if strings.TrimSpace(id) != "" {
		return id
	}
	return domain.NewID(prefix)
}

func times(created, updated time.Time) (string, string) {
	n := time.Now().UTC()
	if created.IsZero() {
		created = n
	}
	if updated.IsZero() {
		updated = created
	}
	return created.UTC().Format(time.RFC3339), updated.UTC().Format(time.RFC3339)
}

func parse(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	t, _ := time.Parse(time.RFC3339, value)
	return t
}

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func intBool(v int) bool { return v != 0 }

func updateStatus(db *sql.DB, table, id string, status any) error {
	_, err := db.Exec(`UPDATE `+table+` SET status = ?, updated_at = ? WHERE id = ?`, status, now(), id)
	return err
}

func deleteByID(db *sql.DB, table, id string) error {
	_, err := db.Exec(`DELETE FROM `+table+` WHERE id = ?`, id)
	return err
}

func jsonList(values []string) string {
	data, _ := json.Marshal(values)
	return string(data)
}

func scanJSONList(value string) []string {
	var values []string
	_ = json.Unmarshal([]byte(value), &values)
	return values
}
