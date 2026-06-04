package store

import (
	"database/sql"

	ctx "github.com/plan-ai/plan-ai/internal/context"
	"github.com/plan-ai/plan-ai/internal/planning"
)

type PlanEvolutionRepository struct{ db *sql.DB }
type ImplementationPackageRepository struct{ db *sql.DB }

func NewPlanEvolutionRepository(db *sql.DB) PlanEvolutionRepository {
	return PlanEvolutionRepository{db: db}
}
func NewImplementationPackageRepository(db *sql.DB) ImplementationPackageRepository {
	return ImplementationPackageRepository{db: db}
}

var _ planning.PlanEvolutionRepository = PlanEvolutionRepository{}
var _ ctx.ImplementationPackageRepository = ImplementationPackageRepository{}

func (r PlanEvolutionRepository) SaveBlueprint(bp planning.PlanEvolutionBlueprint) (planning.PlanEvolutionBlueprint, error) {
	c, u := timestamps(bp.CreatedAt, bp.UpdatedAt)
	_, err := r.db.Exec(`INSERT INTO plan_evolution_blueprints_v3 (id, project_id, objective, scope, exclusions, dependencies, stack, versions, libraries, folders, files, validations, tests, risks, rollback, approved_inputs, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET objective=excluded.objective, scope=excluded.scope, exclusions=excluded.exclusions, dependencies=excluded.dependencies, stack=excluded.stack, versions=excluded.versions, libraries=excluded.libraries, folders=excluded.folders, files=excluded.files, validations=excluded.validations, tests=excluded.tests, risks=excluded.risks, rollback=excluded.rollback, approved_inputs=excluded.approved_inputs, status=excluded.status, updated_at=excluded.updated_at`, bp.ID, bp.ProjectID, bp.Objective, mustJSON(bp.Scope), mustJSON(bp.Exclusions), mustJSON(bp.Dependencies), mustJSON(bp.Stack), mustJSON(bp.Versions), mustJSON(bp.Libraries), mustJSON(bp.Folders), mustJSON(bp.Files), mustJSON(bp.Validations), mustJSON(bp.Tests), mustJSON(bp.Risks), mustJSON(bp.Rollback), mustJSON(bp.ApprovedInputs), bp.Status, c, u)
	if err != nil {
		return planning.PlanEvolutionBlueprint{}, err
	}
	bp.CreatedAt, bp.UpdatedAt = parseRFC3339(c), parseRFC3339(u)
	return bp, nil
}

func (r PlanEvolutionRepository) ListBlueprints(projectID string) ([]planning.PlanEvolutionBlueprint, error) {
	rows, err := r.db.Query(`SELECT id, project_id, objective, scope, exclusions, dependencies, stack, versions, libraries, folders, files, validations, tests, risks, rollback, approved_inputs, status, created_at, updated_at FROM plan_evolution_blueprints_v3 WHERE project_id = ? ORDER BY created_at, id`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []planning.PlanEvolutionBlueprint
	for rows.Next() {
		var bp planning.PlanEvolutionBlueprint
		var scope, exclusions, deps, stack, versions, libraries, folders, files, validations, tests, risks, rollback, inputs, c, u string
		if err := rows.Scan(&bp.ID, &bp.ProjectID, &bp.Objective, &scope, &exclusions, &deps, &stack, &versions, &libraries, &folders, &files, &validations, &tests, &risks, &rollback, &inputs, &bp.Status, &c, &u); err != nil {
			return nil, err
		}
		decodeJSON(scope, &bp.Scope)
		decodeJSON(exclusions, &bp.Exclusions)
		decodeJSON(deps, &bp.Dependencies)
		decodeJSON(stack, &bp.Stack)
		decodeJSON(versions, &bp.Versions)
		decodeJSON(libraries, &bp.Libraries)
		decodeJSON(folders, &bp.Folders)
		decodeJSON(files, &bp.Files)
		decodeJSON(validations, &bp.Validations)
		decodeJSON(tests, &bp.Tests)
		decodeJSON(risks, &bp.Risks)
		decodeJSON(rollback, &bp.Rollback)
		decodeJSON(inputs, &bp.ApprovedInputs)
		bp.CreatedAt, bp.UpdatedAt = parseRFC3339(c), parseRFC3339(u)
		out = append(out, bp)
	}
	return out, rows.Err()
}

func (r ImplementationPackageRepository) SaveImplementationPackage(pkg ctx.ImplementationPackage) (ctx.ImplementationPackage, error) {
	c, u := timestamps(pkg.CreatedAt, pkg.UpdatedAt)
	_, err := r.db.Exec(`INSERT INTO implementation_packages_v2 (id, project_id, plan_id, model_target, what_to_do, how_to_do_it, files_to_touch, files_not_to_touch, examples, commands, validations, rollback_notes, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET model_target=excluded.model_target, what_to_do=excluded.what_to_do, how_to_do_it=excluded.how_to_do_it, files_to_touch=excluded.files_to_touch, files_not_to_touch=excluded.files_not_to_touch, examples=excluded.examples, commands=excluded.commands, validations=excluded.validations, rollback_notes=excluded.rollback_notes, status=excluded.status, updated_at=excluded.updated_at`, pkg.ID, pkg.ProjectID, pkg.PlanID, pkg.ModelTarget, pkg.WhatToDo, pkg.HowToDoIt, mustJSON(pkg.FilesToTouch), mustJSON(pkg.FilesNotToTouch), mustJSON(pkg.Examples), mustJSON(pkg.Commands), mustJSON(pkg.Validations), mustJSON(pkg.RollbackNotes), pkg.Status, c, u)
	if err != nil {
		return ctx.ImplementationPackage{}, err
	}
	pkg.CreatedAt, pkg.UpdatedAt = parseRFC3339(c), parseRFC3339(u)
	return pkg, nil
}

func (r ImplementationPackageRepository) ListImplementationPackages(projectID string) ([]ctx.ImplementationPackage, error) {
	rows, err := r.db.Query(`SELECT id, project_id, plan_id, model_target, what_to_do, how_to_do_it, files_to_touch, files_not_to_touch, examples, commands, validations, rollback_notes, status, created_at, updated_at FROM implementation_packages_v2 WHERE project_id = ? ORDER BY created_at, id`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ctx.ImplementationPackage
	for rows.Next() {
		var pkg ctx.ImplementationPackage
		var touch, notTouch, examples, commands, validations, rollback, c, u string
		if err := rows.Scan(&pkg.ID, &pkg.ProjectID, &pkg.PlanID, &pkg.ModelTarget, &pkg.WhatToDo, &pkg.HowToDoIt, &touch, &notTouch, &examples, &commands, &validations, &rollback, &pkg.Status, &c, &u); err != nil {
			return nil, err
		}
		decodeJSON(touch, &pkg.FilesToTouch)
		decodeJSON(notTouch, &pkg.FilesNotToTouch)
		decodeJSON(examples, &pkg.Examples)
		decodeJSON(commands, &pkg.Commands)
		decodeJSON(validations, &pkg.Validations)
		decodeJSON(rollback, &pkg.RollbackNotes)
		pkg.CreatedAt, pkg.UpdatedAt = parseRFC3339(c), parseRFC3339(u)
		out = append(out, pkg)
	}
	return out, rows.Err()
}
