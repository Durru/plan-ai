package change

import "testing"

func TestBuildDeepImpactReportDetectsDatabaseMigration(t *testing.T) {
	report := BuildDeepImpactReport("project", TechnologyChanged, "PostgreSQL to MariaDB")
	if report.Severity != SeverityHigh || len(report.MigrationConcerns) < 2 || len(report.ValidationCommands) == 0 {
		t.Fatalf("unexpected report: %#v", report)
	}
}
