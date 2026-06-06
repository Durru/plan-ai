// Package change implements the Change Engine — event detection, impact
// analysis, entity invalidation, snapshots, and versioning for
// continuous planning. It enforces v4 Principle 5: "Solo se actualiza
// lo afectado por un cambio aprobado."
//
// Main types: ChangeEvent, Service, Analyzer, SnapshotManager, VersionManager.
// Main entry: Service.RegisterChange for event detection → impact → proposal flow.
package change
