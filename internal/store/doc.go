// Package store implements Plan-AI's persistence layer — SQLite schema,
// migrations, and all repository implementations. It is the only package
// that contains raw SQL. All other packages use typed interfaces.
//
// Main types: Migrations, Repositories, GlobalLayout, ProjectLayout.
// Main entry: Open, RunProjectMigrations, NewRepositories.
package store
