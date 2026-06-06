package mcp

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/plan-ai/plan-ai/internal/config"
	"github.com/plan-ai/plan-ai/internal/conversation"
	"github.com/plan-ai/plan-ai/internal/domain"
	"github.com/plan-ai/plan-ai/internal/guard"
	"github.com/plan-ai/plan-ai/internal/intentv3"
	"github.com/plan-ai/plan-ai/internal/store"
)

// ── Helpers ──

func getStringArg(args map[string]any, key string) string {
	v, ok := args[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

func getIntArg(args map[string]any, key string) int {
	v, ok := args[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	}
	return 0
}

func getProjectRoot(args map[string]any) (string, error) {
	if root := getStringArg(args, "project_root"); root != "" {
		return root, nil
	}
	return store.ResolveProjectRoot()
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

var (
	sharedProjectStore   *store.ProjectStore
	sharedProjectRoot    string
	sharedProjectClosed  bool
)

// SetSharedProjectStore opens and caches a project store that will be reused
// by openStore for the same project root. Callers should call
// CloseSharedProjectStore when the process is done.
func SetSharedProjectStore(projectRoot string) error {
	ps, err := store.OpenProjectStore(projectRoot)
	if err != nil {
		return fmt.Errorf("open shared store: %w", err)
	}
	if sharedProjectStore != nil && !sharedProjectClosed {
		sharedProjectStore.Close()
	}
	sharedProjectStore = ps
	sharedProjectRoot = projectRoot
	sharedProjectClosed = false
	return nil
}

// CloseSharedProjectStore closes the shared project store if one is open.
func CloseSharedProjectStore() error {
	if sharedProjectStore != nil && !sharedProjectClosed {
		sharedProjectClosed = true
		return sharedProjectStore.Close()
	}
	return nil
}

// openStore opens the project store for the given root path.
// Returns the store, a cleanup function (call via defer cleanup()), and any error.
// If a shared store is active for this project root, the cleanup is a no-op.
func openStore(projectRoot string) (*store.ProjectStore, func(), error) {
	if sharedProjectStore != nil && sharedProjectRoot == projectRoot && !sharedProjectClosed {
		return sharedProjectStore, func() {}, nil
	}
	ps, err := store.OpenProjectStore(projectRoot)
	if err != nil {
		return nil, nil, err
	}
	return ps, func() { ps.Close() }, nil
}

// projectID returns the canonical project ID for the given root path.
func projectID(projectRoot string) string {
	return store.ProjectID(projectRoot)
}

// ── Core Project Handlers ──

func HandleInitProject(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, fmt.Errorf("resolve project root: %w", err)
	}
	name := getStringArg(args, "name")
	if name == "" {
		name = projectRoot
	}

	// Default mode is "external" unless the caller explicitly requests "local".
	requestedMode := strings.ToLower(strings.TrimSpace(getStringArg(args, "mode")))
	if requestedMode == "" {
		requestedMode = config.ProjectModeExternal
	}
	if requestedMode != config.ProjectModeExternal && requestedMode != config.ProjectModeLocal {
		return nil, fmt.Errorf("invalid mode %q: must be %q or %q", requestedMode, config.ProjectModeExternal, config.ProjectModeLocal)
	}

	home, err := store.ResolveHomeRoot()
	if err != nil {
		return nil, fmt.Errorf("resolve home root: %w", err)
	}
	if _, err := os.Stat(config.GlobalDBPath(home)); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("global store not initialized; run `plan-ai install` first")
		}
		return nil, fmt.Errorf("stat global db: %w", err)
	}
	globalDB, err := store.Open(config.GlobalDBPath(home))
	if err != nil {
		return nil, fmt.Errorf("open global db: %w", err)
	}
	defer globalDB.Close()
	if err := store.RunGlobalMigrations(globalDB); err != nil {
		return nil, fmt.Errorf("run global migrations: %w", err)
	}

	resolver := store.NewProjectResolver(home, globalDB)

	// Check if the project is already registered so we can detect real
	// "mode mismatch" cases versus fresh registrations.
	registry := store.NewProjectRegistryRepository(globalDB)
	existing, existingErr := registry.GetByPath(projectRoot)
	alreadyRegistered := existingErr == nil

	loc, err := resolver.Resolve(projectRoot)
	if err != nil {
		if errors.Is(err, store.ErrLegacyLocalStoreFound) {
			// A legacy <root>/.plan-ai/project.db exists. Do NOT silently
			// migrate or overwrite it — surface the choice to the caller.
			return map[string]any{
				"status":       "legacy_local_detected",
				"project_root": projectRoot,
				"name":         name,
				"db_path":      config.ProjectDBPath(projectRoot),
				"hint":         "legacy project-local store detected; pass mode='local' to use it OR run 'plan-ai migrate local-to-global' first",
			}, nil
		}
		return nil, fmt.Errorf("resolve project: %w", err)
	}

	// Decide the effective mode. The resolver returns a DRAFT external
	// ProjectLocation when the project is not registered and no legacy
	// store exists. In that case the caller's requested mode wins.
	effectiveMode := loc.Mode
	if !alreadyRegistered {
		effectiveMode = requestedMode
		if effectiveMode == config.ProjectModeLocal {
			// Rebuild a local ProjectLocation manually.
			layout, err := store.EnsureProjectLayout(projectRoot)
			if err != nil {
				return nil, fmt.Errorf("ensure local layout: %w", err)
			}
			loc = store.ProjectLocation{
				ProjectID: store.ProjectID(projectRoot),
				Name:      name,
				RootPath:  projectRoot,
				Slug:      config.ProjectSlug(projectRoot),
				Mode:      config.ProjectModeLocal,
				Layout: store.ProjectLocationLayout{
					Dir:          layout.Dir,
					ConfigPath:   layout.ConfigPath,
					CacheDir:     layout.CacheDir,
					SnapshotsDir: layout.SnapshotsDir,
					ExportsDir:   layout.ExportsDir,
					DocsDir:      layout.DocsDir,
					LocksDir:     layout.LocksDir,
					BackupsDir:   layout.BackupsDir,
				},
				DBPath: layout.DBPath,
			}
		}
	} else {
		// Project is already registered: surface mode mismatches instead
		// of silently switching storage layouts.
		if existing.Mode != requestedMode {
			return map[string]any{
				"status":       "mode_mismatch",
				"project_root": projectRoot,
				"name":         name,
				"db_path":      loc.DBPath,
				"registered_as": existing.Mode,
				"requested":    requestedMode,
				"hint":         "project is already registered in a different mode; run 'plan-ai migrate local-to-global' first or pass mode='" + existing.Mode + "'",
			}, nil
		}
	}

	db, err := store.Open(loc.DBPath)
	if err != nil {
		return nil, fmt.Errorf("open project db: %w", err)
	}
	defer db.Close()
	if err := store.RunProjectMigrations(db); err != nil {
		return nil, fmt.Errorf("run project migrations: %w", err)
	}

	// Persist project metadata in the project DB (project_state table).
	if err := store.UpsertProjectState(db, loc.ProjectID, name, projectRoot, "active"); err != nil {
		return nil, fmt.Errorf("save project state: %w", err)
	}

	// Register the project in the global known_projects registry, recording
	// the chosen mode and slug.
	if _, err := registry.Register(store.ProjectRegistryEntry{
		ID:       loc.ProjectID,
		Name:     name,
		RootPath: projectRoot,
		Slug:     loc.Slug,
		Mode:     effectiveMode,
	}); err != nil {
		return nil, fmt.Errorf("register project: %w", err)
	}

	// Persist the per-project config.json (now includes Mode).
	projectCfg := config.ProjectConfig{
		Version:      "2.0.0",
		ProjectName:  name,
		ProjectRoot:  projectRoot,
		ProjectDB:    loc.DBPath,
		Mode:         effectiveMode,
		CreatedAt:    nowUTCString(),
		Integrations: map[string]any{},
	}
	if err := config.SaveProjectConfig(loc.Layout.ConfigPath, projectCfg); err != nil {
		return nil, fmt.Errorf("save project config: %w", err)
	}

	return map[string]any{
		"status":       "initialized",
		"project_root": projectRoot,
		"project_id":   loc.ProjectID,
		"slug":         loc.Slug,
		"mode":         effectiveMode,
		"db_path":      loc.DBPath,
		"config_path":  loc.Layout.ConfigPath,
		"name":         name,
	}, nil
}

func nowUTCString() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func HandleProjectStatus(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	home, err := store.ResolveHomeRoot()
	if err != nil {
		return nil, fmt.Errorf("resolve home root: %w", err)
	}
	if _, err := os.Stat(config.GlobalDBPath(home)); err != nil {
		if os.IsNotExist(err) {
			return map[string]any{
				"status":       "global_not_initialized",
				"project_root": projectRoot,
				"hint":         "global store not initialized; run `plan-ai install` first",
			}, nil
		}
		return nil, fmt.Errorf("stat global db: %w", err)
	}
	globalDB, err := store.Open(config.GlobalDBPath(home))
	if err != nil {
		return nil, fmt.Errorf("open global db: %w", err)
	}
	defer globalDB.Close()
	if err := store.RunGlobalMigrations(globalDB); err != nil {
		return nil, fmt.Errorf("run global migrations: %w", err)
	}

	resolver := store.NewProjectResolver(home, globalDB)
	loc, err := resolver.Resolve(projectRoot)
	if err != nil {
		if errors.Is(err, store.ErrLegacyLocalStoreFound) {
			return map[string]any{
				"status":       "legacy-local-detected",
				"project_root": projectRoot,
				"mode":         "legacy-local-detected",
				"db_path":      config.ProjectDBPath(projectRoot),
				"hint":         "legacy project-local store detected; run `plan-ai migrate local-to-global` first or pass mode='local' to `init_project`",
			}, nil
		}
		return nil, fmt.Errorf("resolve project: %w", err)
	}

	db, err := store.Open(loc.DBPath)
	if err != nil {
		return nil, fmt.Errorf("open project db: %w", err)
	}
	defer db.Close()
	if err := store.RunProjectMigrations(db); err != nil {
		return nil, fmt.Errorf("run project migrations: %w", err)
	}

	counts, err := store.CountDomainEntities(db)
	if err != nil {
		return nil, fmt.Errorf("count entities: %w", err)
	}

	return map[string]any{
		"status":            "active",
		"project_root":      projectRoot,
		"project_id":        loc.ProjectID,
		"slug":              loc.Slug,
		"mode":              loc.Mode,
		"db_path":           loc.DBPath,
		"config_path":       loc.Layout.ConfigPath,
		"plans":             counts.Plans,
		"phases":            counts.Phases,
		"tasks":             counts.Tasks,
		"decisions":         counts.Decisions,
		"research_entries":  counts.ResearchEntries,
		"knowledge_objects": counts.KnowledgeObjects,
		"validations":       counts.Validations,
		"snapshots":         counts.Snapshots,
	}, nil
}

func HandleCreateMasterPlan(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer cleanup()

	pid := projectID(projectRoot)

	if err := guard.GuardPlanningInput(ps.DB, pid); err != nil {
		return map[string]any{
			"status":  "blocked",
			"message": err.Error(),
			"next":    "Create and approve a product intent before planning",
		}, nil
	}

	repos := store.NewRepositories(ps.DB)

	plan := domain.MasterPlan{
		ID:        domain.NewID("plan"),
		ProjectID: pid,
		Title:     getStringArg(args, "title"),
		Summary:   getStringArg(args, "summary"),
		Status:    domain.StatusDraft,
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := repos.Plan.SaveMaster(plan); err != nil {
		return nil, fmt.Errorf("save master plan: %w", err)
	}

	return map[string]any{
		"plan_id":    plan.ID,
		"title":      plan.Title,
		"summary":    plan.Summary,
		"status":     string(plan.Status),
		"version":    plan.Version,
		"created_at": formatTime(plan.CreatedAt),
	}, nil
}

func HandleCreateSpecificPlan(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer cleanup()

	pid := projectID(projectRoot)

	if err := guard.GuardPlanningInput(ps.DB, pid); err != nil {
		return map[string]any{
			"status":  "blocked",
			"message": err.Error(),
			"next":    "Create and approve a product intent before planning",
		}, nil
	}

	repos := store.NewRepositories(ps.DB)
	masterPlanID := getStringArg(args, "master_plan_id")
	goal := getStringArg(args, "goal")

	plan := domain.SpecificPlan{
		ID:           domain.NewID("plan"),
		ProjectID:    pid,
		MasterPlanID: masterPlanID,
		Title:        getStringArg(args, "title"),
		Summary:      goal,
		Status:       domain.StatusDraft,
		Version:      1,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := repos.Plan.SaveSpecific(plan); err != nil {
		return nil, fmt.Errorf("save specific plan: %w", err)
	}

	return map[string]any{
		"plan_id":        plan.ID,
		"master_plan_id": plan.MasterPlanID,
		"title":          plan.Title,
		"goal":           plan.Summary,
		"status":         string(plan.Status),
		"created_at":     formatTime(plan.CreatedAt),
	}, nil
}

func HandleResearchTopic(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer cleanup()

	repos := store.NewRepositories(ps.DB)
	pid := projectID(projectRoot)

	research := domain.Research{
		ID:         domain.NewID("research"),
		ProjectID:  pid,
		Topic:      getStringArg(args, "topic"),
		Summary:    getStringArg(args, "summary"),
		Confidence: float64(getIntArg(args, "confidence")),
		Status:     domain.ResearchStatusDraft,
		Category:   domain.KnowledgeCategoryGeneral,
		Date:       time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := repos.Research.Save(research); err != nil {
		return nil, fmt.Errorf("save research: %w", err)
	}

	return map[string]any{
		"research_id": research.ID,
		"topic":       research.Topic,
		"summary":     research.Summary,
		"status":      string(research.Status),
		"confidence":  research.Confidence,
		"created_at":  formatTime(research.CreatedAt),
	}, nil
}

func HandleApprovePlan(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer cleanup()

	repos := store.NewRepositories(ps.DB)
	planID := getStringArg(args, "plan_id")

	if err := repos.Plan.UpdatePlanStatus(planID, domain.StatusApproved); err != nil {
		return nil, fmt.Errorf("approve plan: %w", err)
	}

	return map[string]any{
		"plan_id": planID,
		"status":  "approved",
	}, nil
}

func HandleRejectPlan(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer cleanup()

	repos := store.NewRepositories(ps.DB)
	planID := getStringArg(args, "plan_id")

	if err := repos.Plan.UpdatePlanStatus(planID, domain.StatusRejected); err != nil {
		return nil, fmt.Errorf("reject plan: %w", err)
	}

	return map[string]any{
		"plan_id": planID,
		"status":  "rejected",
	}, nil
}

func HandleAnalyzeImpact(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer cleanup()

	repos := store.NewRepositories(ps.DB)
	pid := projectID(projectRoot)
	changeType := getStringArg(args, "change_type")
	summary := getStringArg(args, "summary")

	cr := domain.ChangeRequest{
		ID:          domain.NewID("change"),
		ProjectID:   pid,
		Reason:      changeType,
		Description: summary,
		Status:      domain.ChangeRequestDraft,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := repos.Change.SaveChangeRequest(cr); err != nil {
		return nil, fmt.Errorf("save change request: %w", err)
	}

	report := domain.ImpactReport{
		ID:              domain.NewID("impact"),
		ChangeRequestID: cr.ID,
		Summary:         fmt.Sprintf("Impact analysis for %s: %s", changeType, summary),
		CreatedAt:       time.Now(),
	}
	if err := repos.Change.SaveImpactReport(report); err != nil {
		return nil, fmt.Errorf("save impact report: %w", err)
	}

	return map[string]any{
		"change_request_id": cr.ID,
		"change_type":       changeType,
		"status":            string(cr.Status),
		"impact_report_id":  report.ID,
		"summary":           summary,
	}, nil
}

func HandleGetNextTask(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer cleanup()

	tasks, err := store.NewRepositories(ps.DB).Task.ListByStatus("pending")
	if err != nil || len(tasks) == 0 {
		return map[string]any{"found": false, "error": "no pending tasks found"}, nil
	}

	t := tasks[0]
	return map[string]any{
		"found":    true,
		"task_id":  t.ID,
		"phase_id": t.PhaseID,
		"plan_id":  t.PlanID,
		"title":    t.Title,
		"summary":  t.Summary,
		"status":   string(t.Status),
		"position": t.Position,
	}, nil
}

func HandleMarkTaskDone(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer cleanup()

	repos := store.NewRepositories(ps.DB)
	taskID := getStringArg(args, "task_id")

	if err := repos.Task.UpdateStatus(taskID, domain.PlanStatusDone); err != nil {
		return nil, fmt.Errorf("update task status: %w", err)
	}

	return map[string]any{
		"task_id": taskID,
		"status":  "done",
	}, nil
}

func HandleCreateSnapshot(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer cleanup()

	repos := store.NewRepositories(ps.DB)
	pid := projectID(projectRoot)

	snapshot := domain.Snapshot{
		ID:        domain.NewID("snapshot"),
		ProjectID: pid,
		Reason:    getStringArg(args, "reason"),
		Summary:   getStringArg(args, "summary"),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := repos.Snapshot.Save(snapshot); err != nil {
		return nil, fmt.Errorf("save snapshot: %w", err)
	}

	return map[string]any{
		"snapshot_id": snapshot.ID,
		"reason":      snapshot.Reason,
		"summary":     snapshot.Summary,
		"created_at":  formatTime(snapshot.CreatedAt),
	}, nil
}

func HandleListPlans(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer cleanup()

	repos := store.NewRepositories(ps.DB)
	pid := projectID(projectRoot)

	masters, err := repos.Plan.ListMastersByProject(pid)
	if err != nil {
		return nil, fmt.Errorf("list master plans: %w", err)
	}

	specifics, err := repos.Plan.ListSpecificsByMaster("")
	if err != nil {
		return nil, fmt.Errorf("list specific plans: %w", err)
	}

	masterList := make([]map[string]any, 0, len(masters))
	for _, m := range masters {
		masterList = append(masterList, map[string]any{
			"id":         m.ID,
			"title":      m.Title,
			"summary":    m.Summary,
			"status":     string(m.Status),
			"version":    m.Version,
			"created_at": formatTime(m.CreatedAt),
		})
	}

	specificList := make([]map[string]any, 0, len(specifics))
	for _, s := range specifics {
		specificList = append(specificList, map[string]any{
			"id":             s.ID,
			"master_plan_id": s.MasterPlanID,
			"title":          s.Title,
			"summary":        s.Summary,
			"status":         string(s.Status),
			"version":        s.Version,
			"created_at":     formatTime(s.CreatedAt),
		})
	}

	return map[string]any{
		"master_plans":   masterList,
		"specific_plans": specificList,
	}, nil
}

func HandleListTasks(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer cleanup()

	planID := getStringArg(args, "plan_id")
	statusFilter := getStringArg(args, "status")

	var tasks []domain.Task
	var err2 error
	repos := store.NewRepositories(ps.DB)
	if planID != "" {
		tasks, err2 = repos.Task.ListByPlanID(planID)
	} else if statusFilter != "" {
		tasks, err2 = repos.Task.ListByStatus(statusFilter)
	} else {
		tasks, err2 = repos.Task.List()
	}
	if err2 != nil {
		return nil, fmt.Errorf("query tasks: %w", err2)
	}

	taskList := make([]map[string]any, len(tasks))
	for i, t := range tasks {
		taskList[i] = map[string]any{
			"id":         t.ID,
			"phase_id":   t.PhaseID,
			"plan_id":    t.PlanID,
			"title":      t.Title,
			"summary":    t.Summary,
			"status":     string(t.Status),
			"position":   t.Position,
			"created_at": formatTime(t.CreatedAt),
		}
	}

	return map[string]any{"tasks": taskList, "count": len(taskList)}, nil
}

// ── Agent System Handlers ──

func HandleAgentProcess(args map[string]any) (map[string]any, error) {
	message := getStringArg(args, "message")
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return map[string]any{
			"status":  "error",
			"message": fmt.Sprintf("open project: %v", err),
		}, nil
	}
	defer cleanup()

	gw := conversation.NewGateway(ps.DB)
	resp, err := gw.ProcessMessage(projectID(projectRoot), message)
	if err != nil {
		return map[string]any{
			"status":  "error",
			"message": fmt.Sprintf("agent: %v", err),
		}, nil
	}

	return map[string]any{
		"status":               resp.Status,
		"message":              resp.Message,
		"intent":               resp.WorkflowTriggered,
		"requires_approval":    resp.RequiresApproval,
		"suggested_next_action": resp.SuggestedNextAction,
		"response": map[string]any{
			"text": resp.Message,
		},
	}, nil
}

func HandleAgentRuns(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer cleanup()

	pid := projectID(projectRoot)
	limit := getIntArg(args, "limit")
	if limit <= 0 {
		limit = 10
	}

	records, err := store.NewAgentRunV2Repository(ps.DB).ListRuns(pid, limit)
	if err != nil {
		return map[string]any{"runs": []any{}, "count": 0}, nil
	}

	runs := make([]map[string]any, len(records))
	for i, rec := range records {
		runs[i] = map[string]any{
			"id":         rec.ID,
			"intent":     rec.Intent,
			"status":     rec.Status,
			"response":   rec.Response,
			"created_at": rec.CreatedAt,
		}
	}

	return map[string]any{"runs": runs, "count": len(runs)}, nil
}

// ── Continuous Planning Handlers ──

func HandleContinuousStatus(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer cleanup()

	pid := projectID(projectRoot)

	rec, err := store.NewContinuousStatusRepository(ps.DB).GetLatest(pid)
	if err != nil {
		return map[string]any{"status": "inactive", "project_root": projectRoot}, nil
	}

	return map[string]any{
		"status":       "active",
		"active_plan":  rec.ActivePlan,
		"active_phase": rec.ActivePhase,
		"next_task":    rec.NextTask,
	}, nil
}

func HandleContinuousEvents(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer cleanup()

	pid := projectID(projectRoot)
	limit := getIntArg(args, "limit")
	if limit <= 0 {
		limit = 10
	}

	records, err := store.NewContinuousEventRepository(ps.DB).ListEvents(pid, limit)
	if err != nil {
		return map[string]any{"events": []any{}, "count": 0}, nil
	}

	events := make([]map[string]any, len(records))
	for i, r := range records {
		events[i] = map[string]any{
			"id":         r.ID,
			"event_type": r.EventType,
			"summary":    r.Summary,
			"created_at": r.CreatedAt,
		}
	}

	return map[string]any{"events": events, "count": len(events)}, nil
}

func HandleContinuousProposals(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer cleanup()

	pid := projectID(projectRoot)
	proposalID := getStringArg(args, "proposal_id")

	propRepo := store.NewPlanUpdateProposalRepository(ps.DB)

	if proposalID != "" {
		if err := propRepo.UpdateProposalStatus(proposalID, "approved"); err != nil {
			return nil, fmt.Errorf("update proposal: %w", err)
		}
		return map[string]any{"proposal_id": proposalID, "status": "approved"}, nil
	}

	reason := getStringArg(args, "reason")
	if reason != "" {
		id := domain.NewID("proposal")
		prop := store.PlanUpdateProposalRecord{
			ID: id, ProjectID: pid, Reason: reason,
			AffectedPlans: "[]", AffectedTasks: "[]", AffectedDecisions: "[]",
			SuggestedUpdates: "", RequiresResearch: 0, RequiresApproval: 1,
			Status: "draft",
		}
		if _, err := propRepo.CreateProposal(prop); err != nil {
			return nil, fmt.Errorf("create proposal: %w", err)
		}
		return map[string]any{"proposal_id": id, "reason": reason, "status": "draft"}, nil
	}

	proposals, err := propRepo.ListProposals(pid)
	if err != nil {
		return nil, fmt.Errorf("list proposals: %w", err)
	}
	var out []map[string]any
	for _, p := range proposals {
		out = append(out, map[string]any{
			"id":         p.ID,
			"reason":     p.Reason,
			"status":     p.Status,
			"created_at": p.CreatedAt,
		})
	}
	if out == nil {
		out = []map[string]any{}
	}
	return map[string]any{"proposals": out, "count": len(out)}, nil
}

func HandleContinuousContext(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer cleanup()

	pid := projectID(projectRoot)
	repos := store.NewRepositories(ps.DB)
	level := getStringArg(args, "level")
	if level == "" {
		level = "L0_Executive"
	}

	deliveryRepo := store.NewContextDeliveryRepository(ps.DB)
	deliveries, _ := deliveryRepo.ListDeliveries(pid, level, 1)
	if len(deliveries) > 0 {
		d := deliveries[0]
		return map[string]any{
			"level":      level,
			"content":    d.Content,
			"cached":     true,
			"created_at": d.CreatedAt,
		}, nil
	}

	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("# Context Level: %s\n", level))
	buf.WriteString(fmt.Sprintf("Project ID: %s\n\n", pid))

	masterPlans, _ := repos.Plan.ListMastersByProject(pid)
	specificPlans, _ := repos.Plan.ListSpecificsByMaster("")
	buf.WriteString("## Plans\n\n")
	for _, p := range masterPlans {
		buf.WriteString(fmt.Sprintf("- **%s** [master] (%s) - %s\n", p.Title, p.Status, p.ID))
		if p.Summary != "" {
			buf.WriteString(fmt.Sprintf("  %s\n", p.Summary))
		}
	}
	for _, p := range specificPlans {
		buf.WriteString(fmt.Sprintf("- **%s** [specific] (%s) - %s\n", p.Title, p.Status, p.ID))
		if p.Summary != "" {
			buf.WriteString(fmt.Sprintf("  %s\n", p.Summary))
		}
	}
	buf.WriteString("\n")

	switch level {
	case "L0_Executive":
		buf.WriteString("## Executive Summary\n\n")
		counts, _ := store.CountDomainEntities(ps.DB)
		buf.WriteString(fmt.Sprintf("- Plans: %d\n", counts.Plans))
		buf.WriteString(fmt.Sprintf("- Phases: %d\n", counts.Phases))
		buf.WriteString(fmt.Sprintf("- Tasks: %d\n", counts.Tasks))
		buf.WriteString(fmt.Sprintf("- Decisions: %d\n", counts.Decisions))
		buf.WriteString(fmt.Sprintf("- Research entries: %d\n", counts.ResearchEntries))

	case "L1_Planning":
		buf.WriteString("## Planning Context\n\n")
		decisions, _ := repos.Decision.ListByProject(pid)
		buf.WriteString("### Decisions\n\n")
		decLimit := 20
		if len(decisions) < decLimit {
			decLimit = len(decisions)
		}
		for _, d := range decisions[:decLimit] {
			buf.WriteString(fmt.Sprintf("- **%s** (%s): %s\n", d.Title, d.Status, d.Decision))
		}

		researchEntries, _ := repos.Research.ListByProject(pid)
		buf.WriteString("\n### Research\n\n")
		resLimit := 20
		if len(researchEntries) < resLimit {
			resLimit = len(researchEntries)
		}
		for _, r := range researchEntries[:resLimit] {
			buf.WriteString(fmt.Sprintf("- **%s** (%s)\n", r.Topic, r.Status))
		}

	case "L2_Implementation":
		buf.WriteString("## Implementation Context\n\n")
		tasks, _ := repos.Task.List()
		taskLimit := 20
		if len(tasks) < taskLimit {
			taskLimit = len(tasks)
		}
		for _, t := range tasks[:taskLimit] {
			buf.WriteString(fmt.Sprintf("### %s [%s]\n**ID:** %s\n%s\n\n", t.Title, t.Status, t.ID, t.Summary))
		}

	case "L3_Research":
		buf.WriteString("## Research Context\n\n")
		researchEntries, _ := repos.Research.ListByProject(pid)
		resLimit := 30
		if len(researchEntries) < resLimit {
			resLimit = len(researchEntries)
		}
		for _, r := range researchEntries[:resLimit] {
			buf.WriteString(fmt.Sprintf("### %s\n**ID:** %s | **Status:** %s | **Confidence:** %.0f\n%s\n\n", r.Topic, r.ID, r.Status, r.Confidence, r.Summary))
		}
	}

	return map[string]any{
		"level":   level,
		"content": buf.String(),
		"cached":  false,
	}, nil
}

// ── Phase 29 Handlers ──

func HandleGetContext(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer cleanup()

	pid := projectID(projectRoot)
	repos := store.NewRepositories(ps.DB)
	level := getStringArg(args, "level")
	if level == "" {
		level = "L0_executive"
	}
	taskID := getStringArg(args, "task_id")
	topic := getStringArg(args, "topic")

	var buf strings.Builder

	switch level {
	case "L0_executive":
		buf.WriteString("# Executive Context\n\n")
		buf.WriteString(fmt.Sprintf("Project: %s\n\n", projectRoot))

		counts, _ := store.CountDomainEntities(ps.DB)
		buf.WriteString(fmt.Sprintf("- Plans: %d\n- Tasks: %d\n- Decisions: %d\n- Research entries: %d\n", counts.Plans, counts.Tasks, counts.Decisions, counts.ResearchEntries))

		masters, _ := repos.Plan.ListMastersByProject(pid)
		buf.WriteString("\n### Recent Plans\n")
		planLimit := 5
		if len(masters) < planLimit {
			planLimit = len(masters)
		}
		for _, p := range masters[:planLimit] {
			buf.WriteString(fmt.Sprintf("- %s (%s)\n", p.Title, p.Status))
		}

	case "L1_planning":
		buf.WriteString("# Planning Context\n\n")
		masters, _ := repos.Plan.ListMastersByProject(pid)
		for _, p := range masters {
			buf.WriteString(fmt.Sprintf("## %s [master]\n**ID:** %s | **Status:** %s\n%s\n\n", p.Title, p.ID, p.Status, p.Summary))
		}
		specifics, _ := repos.Plan.ListSpecificsByMaster("")
		for _, p := range specifics {
			buf.WriteString(fmt.Sprintf("## %s [specific]\n**ID:** %s | **Status:** %s\n%s\n\n", p.Title, p.ID, p.Status, p.Summary))
		}

	case "L2_implementation":
		buf.WriteString("# Implementation Context\n\n")
		if taskID != "" {
			t, err := repos.Task.GetByID(taskID)
			if err == nil {
				buf.WriteString(fmt.Sprintf("## Task: %s\n**ID:** %s | **Status:** %s\n%s\n", t.Title, t.ID, t.Status, t.Summary))
			} else {
				buf.WriteString("Task not found.\n")
			}
		} else {
			pending, _ := repos.Task.ListByStatus("pending")
			active, _ := repos.Task.ListByStatus("active")
			all := append(pending, active...)
			buf.WriteString("### Active/Pending Tasks\n")
			taskLimit := 10
			if len(all) < taskLimit {
				taskLimit = len(all)
			}
			for _, t := range all[:taskLimit] {
				buf.WriteString(fmt.Sprintf("- [%s] %s (%s)\n", t.Status, t.Title, t.ID))
			}
		}

	case "L3_research":
		buf.WriteString("# Research Context\n\n")
		var entries []domain.Research
		var err error
		if topic != "" {
			entries, err = repos.Research.Search(topic)
		} else {
			entries, err = repos.Research.ListByProject(pid)
		}
		if err == nil {
			for _, r := range entries {
				buf.WriteString(fmt.Sprintf("## %s\n**ID:** %s | **Status:** %s | **Confidence:** %.0f\n%s\n\n", r.Topic, r.ID, r.Status, r.Confidence, r.Summary))
			}
		}
	}

	return map[string]any{
		"level":   level,
		"content": buf.String(),
	}, nil
}

func HandleDetectChanges(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer cleanup()

	repos := store.NewRepositories(ps.DB)
	pid := projectID(projectRoot)
	changeType := getStringArg(args, "change_type")
	summary := getStringArg(args, "summary")
	description := getStringArg(args, "description")

	cr := domain.ChangeRequest{
		ID:          domain.NewID("change"),
		ProjectID:   pid,
		Reason:      changeType,
		Description: description,
		Status:      domain.ChangeRequestDraft,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := repos.Change.SaveChangeRequest(cr); err != nil {
		return nil, fmt.Errorf("save change request: %w", err)
	}

	// Determine affected entities based on change type
	var affectedPlans, affectedDecisions []string

	switch changeType {
	case "plan_changed", "vision_changed":
		masters, _ := repos.Plan.ListMastersByProject(pid)
		for _, p := range masters {
			affectedPlans = append(affectedPlans, p.ID)
		}
		specifics, _ := repos.Plan.ListSpecificsByMaster("")
		for _, p := range specifics {
			affectedPlans = append(affectedPlans, p.ID)
		}
	case "decision_changed":
		decisions, _ := repos.Decision.ListByProject(pid)
		for _, d := range decisions {
			affectedDecisions = append(affectedDecisions, d.ID)
		}
	case "research_updated", "knowledge_updated":
		masters, _ := repos.Plan.ListMastersByProject(pid)
		for _, p := range masters {
			affectedPlans = append(affectedPlans, p.ID)
		}
		specifics, _ := repos.Plan.ListSpecificsByMaster("")
		for _, p := range specifics {
			affectedPlans = append(affectedPlans, p.ID)
		}
	}

	report := domain.ImpactReport{
		ID:                domain.NewID("impact"),
		ChangeRequestID:   cr.ID,
		AffectedPlans:     affectedPlans,
		AffectedDecisions: affectedDecisions,
		Summary:           fmt.Sprintf("Change detected: %s - %s", changeType, summary),
		CreatedAt:         time.Now(),
	}
	if err := repos.Change.SaveImpactReport(report); err != nil {
		return nil, fmt.Errorf("save impact report: %w", err)
	}

	return map[string]any{
		"change_request_id":  cr.ID,
		"change_type":        changeType,
		"impact_report_id":   report.ID,
		"affected_plans":     affectedPlans,
		"affected_decisions": affectedDecisions,
		"summary":            summary,
	}, nil
}

func HandleUpdatePlan(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer cleanup()

	repos := store.NewRepositories(ps.DB)
	planID := getStringArg(args, "plan_id")
	newTitle := getStringArg(args, "title")
	newSummary := getStringArg(args, "summary")
	newStatus := getStringArg(args, "status")

	// Try master plan first
	master, err := repos.Plan.GetMasterByID(planID)
	if err == nil {
		if newTitle != "" {
			master.Title = newTitle
		}
		if newSummary != "" {
			master.Summary = newSummary
		}
		if newStatus != "" {
			master.Status = domain.Status(newStatus)
		}
		master.UpdatedAt = time.Now()
		if err := repos.Plan.SaveMaster(master); err != nil {
			return nil, fmt.Errorf("update master plan: %w", err)
		}
		return map[string]any{
			"plan_id": master.ID,
			"type":    "master",
			"title":   master.Title,
			"status":  string(master.Status),
		}, nil
	}

	// Try specific plan
	specific, err := repos.Plan.GetSpecificByID(planID)
	if err == nil {
		if newTitle != "" {
			specific.Title = newTitle
		}
		if newSummary != "" {
			specific.Summary = newSummary
		}
		if newStatus != "" {
			specific.Status = domain.Status(newStatus)
		}
		specific.UpdatedAt = time.Now()
		if err := repos.Plan.SaveSpecific(specific); err != nil {
			return nil, fmt.Errorf("update specific plan: %w", err)
		}
		return map[string]any{
			"plan_id": specific.ID,
			"type":    "specific",
			"title":   specific.Title,
			"status":  string(specific.Status),
		}, nil
	}

	// Fallback: if only status is given, try UpdatePlanStatus
	if newStatus != "" {
		if err := repos.Plan.UpdatePlanStatus(planID, domain.Status(newStatus)); err != nil {
			return nil, fmt.Errorf("update plan status: %w", err)
		}
		return map[string]any{
			"plan_id": planID,
			"status":  newStatus,
		}, nil
	}

	return nil, fmt.Errorf("plan not found: %s", planID)
}

func HandleRollbackSnapshot(args map[string]any) (map[string]any, error) {
	snapshotID := getStringArg(args, "snapshot_id")
	if snapshotID == "" {
		return nil, fmt.Errorf("snapshot_id is required")
	}

	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer cleanup()

	repos := store.NewRepositories(ps.DB)

	snap, err := repos.Snapshot.GetByID(snapshotID)
	if err != nil {
		return map[string]any{
			"supported":   false,
			"snapshot_id": snapshotID,
			"message":     fmt.Sprintf("snapshot %s not found: %v", snapshotID, err),
		}, nil
	}

	// Check snapshots_v2 for entity-level rollback data.
	snapV2, err := store.NewSnapshotV2Repository(ps.DB).GetByID(snapshotID)
	if err == nil && snapV2.EntitySnapshot != "" && snapV2.EntitySnapshot != "{}" {
		var entities map[string]any
		_ = json.Unmarshal([]byte(snapV2.EntitySnapshot), &entities)
		restored := 0
		if plans, ok := entities["plans"].([]any); ok {
			for _, p := range plans {
				if pm, ok := p.(map[string]any); ok {
					planID, _ := pm["id"].(string)
					status, _ := pm["status"].(string)
					if planID != "" && status != "" {
						_ = repos.Plan.UpdatePlanStatus(planID, domain.Status(status))
						tasksForPlan, _ := repos.Task.ListByPlanID(planID)
						for _, t := range tasksForPlan {
							_ = repos.Task.UpdateStatus(t.ID, domain.PlanStatusPending)
						}
						restored++
					}
				}
			}
		}
		if decisions, ok := entities["decisions"].([]any); ok {
			for _, d := range decisions {
				if dm, ok := d.(map[string]any); ok {
					did, _ := dm["id"].(string)
					status, _ := dm["status"].(string)
					if did != "" && status != "" {
						_ = repos.Decision.UpdateStatus(did, domain.Status(status))
						restored++
					}
				}
			}
		}
		if tasks, ok := entities["tasks"].([]any); ok {
			for _, t := range tasks {
				if tm, ok := t.(map[string]any); ok {
					tid, _ := tm["id"].(string)
					status, _ := tm["status"].(string)
					if tid != "" && status != "" {
						_ = repos.Task.UpdateStatus(tid, domain.Status(status))
						restored++
					}
				}
			}
		}
		return map[string]any{
			"supported":    true,
			"snapshot_id":  snapshotID,
			"reason":       snap.Reason,
			"summary":      snap.Summary,
			"restored":     restored,
			"message":      fmt.Sprintf("Rollback complete: %d entities restored to snapshot state", restored),
		}, nil
	}

	return map[string]any{
		"supported":   true,
		"snapshot_id": snapshotID,
		"reason":      snap.Reason,
		"summary":     snap.Summary,
		"restored":    0,
		"message":     "Snapshot found but no entity-level rollback data available (use snapshots_v2 with entity_snapshot for full rollback)",
	}, nil
}

func HandleExportDocs(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer cleanup()

	format := getStringArg(args, "format")
	if format == "" {
		format = "markdown"
	}
	scope := getStringArg(args, "scope")

	repos := store.NewRepositories(ps.DB)
	pid := projectID(projectRoot)

	var buf strings.Builder

	switch scope {
	case "plans":
		buf.WriteString("# Project Plans\n\n")
		masters, _ := repos.Plan.ListMastersByProject(pid)
		for _, m := range masters {
			buf.WriteString(fmt.Sprintf("## %s\n- **ID:** %s\n- **Status:** %s\n- **Version:** %d\n\n%s\n\n", m.Title, m.ID, m.Status, m.Version, m.Summary))
		}
		specifics, _ := repos.Plan.ListSpecificsByMaster("")
		for _, s := range specifics {
			buf.WriteString(fmt.Sprintf("### %s (Master: %s)\n- **ID:** %s\n- **Status:** %s\n\n%s\n\n", s.Title, s.MasterPlanID, s.ID, s.Status, s.Summary))
		}

	case "decisions":
		buf.WriteString("# Project Decisions\n\n")
		decisions, _ := repos.Decision.ListByProject(pid)
		for _, d := range decisions {
			buf.WriteString(fmt.Sprintf("## %s\n- **ID:** %s\n- **Status:** %s\n\n**Decision:** %s\n\n**Context:** %s\n\n**Impact:** %s\n\n", d.Title, d.ID, d.Status, d.Decision, d.Context, d.Impact))
		}

	case "research":
		buf.WriteString("# Research\n\n")
		entries, _ := repos.Research.ListByProject(pid)
		for _, e := range entries {
			buf.WriteString(fmt.Sprintf("## %s\n- **ID:** %s\n- **Status:** %s\n- **Confidence:** %.0f\n\n%s\n\n", e.Topic, e.ID, e.Status, e.Confidence, e.Summary))
		}

	case "all":
		buf.WriteString("# Project Documentation\n\n")

		// Plans
		buf.WriteString("## Plans\n\n")
		masters, _ := repos.Plan.ListMastersByProject(pid)
		for _, m := range masters {
			buf.WriteString(fmt.Sprintf("- **%s** (%s) - %s\n", m.Title, m.ID, m.Status))
			if m.Summary != "" {
				buf.WriteString(fmt.Sprintf("  %s\n", m.Summary))
			}
		}

		// Decisions
		buf.WriteString("\n## Decisions\n\n")
		decisions, _ := repos.Decision.ListByProject(pid)
		for _, d := range decisions {
			buf.WriteString(fmt.Sprintf("- **%s** (%s): %s\n", d.Title, d.Status, d.Decision))
		}

		// Research
		buf.WriteString("\n## Research\n\n")
		entries, _ := repos.Research.ListByProject(pid)
		for _, e := range entries {
			buf.WriteString(fmt.Sprintf("- **%s** (%s, confidence: %.0f)\n", e.Topic, e.Status, e.Confidence))
		}

		// Snapshots
		buf.WriteString("\n## Snapshots\n\n")
		snaps, _ := repos.Snapshot.ListByProject(pid)
		for _, s := range snaps {
			buf.WriteString(fmt.Sprintf("- **%s** (%s): %s\n", s.Reason, s.ID, s.Summary))
		}
	}

	content := buf.String()

	return map[string]any{
		"format":  format,
		"scope":   scope,
		"content": content,
		"chars":   len(content),
	}, nil
}

// ── Phase 51: Product Intent Engine Handlers ──

func HandleCreateProductIntent(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, fmt.Errorf("resolve project root: %w", err)
	}
	description := getStringArg(args, "description")
	if description == "" {
		return nil, fmt.Errorf("description is required")
	}
	expectedOutcome := getStringArg(args, "expected_outcome")
	desiredExperience := getStringArg(args, "desired_experience")
	desiredResult := getStringArg(args, "desired_result")

	var userExpectations, nonExpectations []string
	if ue := getStringArg(args, "user_expectations"); ue != "" {
		userExpectations = strings.Split(ue, "\n")
	}
	if ne := getStringArg(args, "non_expectations"); ne != "" {
		nonExpectations = strings.Split(ne, "\n")
	}

	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	intentRepo := store.NewIntentV3Repository(ps.DB)
	discRepo := store.NewIntentV3DiscoveryResultRepository(ps.DB)
	svc := intentv3.NewService(intentRepo, discRepo)

	pi, err := svc.CreateProductIntent(intentv3.CreateProductIntentInput{
		ProjectID:         projectID(projectRoot),
		Description:       description,
		ExpectedOutcome:   expectedOutcome,
		DesiredExperience: desiredExperience,
		DesiredResult:     desiredResult,
		UserExpectations:  userExpectations,
		NonExpectations:   nonExpectations,
		SuccessDefinition: getStringArg(args, "success_definition"),
		FailureDefinition: getStringArg(args, "failure_definition"),
		DiscoveryResultID: getStringArg(args, "discovery_result_id"),
	})
	if err != nil {
		return nil, fmt.Errorf("create product intent: %w", err)
	}

	return map[string]any{
		"id":                  pi.ID,
		"project_id":          pi.ProjectID,
		"description":         pi.Description,
		"status":              string(pi.Status),
		"expected_outcome":    pi.ExpectedOutcome,
		"desired_experience":  pi.DesiredExperience,
		"desired_result":      pi.DesiredResult,
		"user_expectations":   pi.UserExpectations,
		"non_expectations":    pi.NonExpectations,
		"success_definition":  pi.SuccessDefinition,
		"failure_definition":  pi.FailureDefinition,
		"discovery_result_id": pi.DiscoveryResultID,
		"created_at":          formatTime(pi.CreatedAt),
		"updated_at":          formatTime(pi.UpdatedAt),
	}, nil
}

func HandleListProductIntents(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, fmt.Errorf("resolve project root: %w", err)
	}
	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	intentRepo := store.NewIntentV3Repository(ps.DB)
	discRepo := store.NewIntentV3DiscoveryResultRepository(ps.DB)
	svc := intentv3.NewService(intentRepo, discRepo)

	list, err := svc.ListProductIntents(projectID(projectRoot))
	if err != nil {
		return nil, fmt.Errorf("list product intents: %w", err)
	}
	items := make([]map[string]any, 0, len(list))
	for _, pi := range list {
		items = append(items, map[string]any{
			"id":          pi.ID,
			"description": pi.Description,
			"status":      string(pi.Status),
			"created_at":  formatTime(pi.CreatedAt),
		})
	}
	return map[string]any{"items": items, "count": len(items)}, nil
}

func HandleGetProductIntent(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, fmt.Errorf("resolve project root: %w", err)
	}
	intentID := getStringArg(args, "intent_id")
	if intentID == "" {
		return nil, fmt.Errorf("intent_id is required")
	}
	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	intentRepo := store.NewIntentV3Repository(ps.DB)
	discRepo := store.NewIntentV3DiscoveryResultRepository(ps.DB)
	svc := intentv3.NewService(intentRepo, discRepo)

	pi, err := svc.GetProductIntent(intentID)
	if err != nil {
		return nil, fmt.Errorf("get product intent: %w", err)
	}
	return map[string]any{
		"id":                  pi.ID,
		"project_id":          pi.ProjectID,
		"description":         pi.Description,
		"status":              string(pi.Status),
		"expected_outcome":    pi.ExpectedOutcome,
		"desired_experience":  pi.DesiredExperience,
		"desired_result":      pi.DesiredResult,
		"user_expectations":   pi.UserExpectations,
		"non_expectations":    pi.NonExpectations,
		"success_definition":  pi.SuccessDefinition,
		"failure_definition":  pi.FailureDefinition,
		"discovery_result_id": pi.DiscoveryResultID,
		"created_at":          formatTime(pi.CreatedAt),
		"updated_at":          formatTime(pi.UpdatedAt),
	}, nil
}

func HandleSubmitProductIntent(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, fmt.Errorf("resolve project root: %w", err)
	}
	intentID := getStringArg(args, "intent_id")
	if intentID == "" {
		return nil, fmt.Errorf("intent_id is required")
	}
	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	intentRepo := store.NewIntentV3Repository(ps.DB)
	discRepo := store.NewIntentV3DiscoveryResultRepository(ps.DB)
	svc := intentv3.NewService(intentRepo, discRepo)

	pi, err := svc.SubmitProductIntentForApproval(intentID)
	if err != nil {
		return nil, fmt.Errorf("submit product intent: %w", err)
	}
	return map[string]any{
		"id":     pi.ID,
		"status": string(pi.Status),
	}, nil
}

func HandleApproveProductIntent(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, fmt.Errorf("resolve project root: %w", err)
	}
	intentID := getStringArg(args, "intent_id")
	if intentID == "" {
		return nil, fmt.Errorf("intent_id is required")
	}
	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	intentRepo := store.NewIntentV3Repository(ps.DB)
	discRepo := store.NewIntentV3DiscoveryResultRepository(ps.DB)
	svc := intentv3.NewService(intentRepo, discRepo)

	pi, err := svc.ApproveProductIntent(intentID)
	if err != nil {
		return nil, fmt.Errorf("approve product intent: %w", err)
	}
	return map[string]any{
		"id":     pi.ID,
		"status": string(pi.Status),
	}, nil
}

func HandleRejectProductIntent(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, fmt.Errorf("resolve project root: %w", err)
	}
	intentID := getStringArg(args, "intent_id")
	if intentID == "" {
		return nil, fmt.Errorf("intent_id is required")
	}
	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	intentRepo := store.NewIntentV3Repository(ps.DB)
	discRepo := store.NewIntentV3DiscoveryResultRepository(ps.DB)
	svc := intentv3.NewService(intentRepo, discRepo)

	pi, err := svc.RejectProductIntent(intentID)
	if err != nil {
		return nil, fmt.Errorf("reject product intent: %w", err)
	}
	return map[string]any{
		"id":     pi.ID,
		"status": string(pi.Status),
	}, nil
}

// ── Phase 52: Discovery Engine Handlers ──

func HandleDiscoverIntent(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, fmt.Errorf("resolve project root: %w", err)
	}
	content := getStringArg(args, "content")
	if content == "" {
		return nil, fmt.Errorf("content is required")
	}
	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	intentRepo := store.NewIntentV3Repository(ps.DB)
	discRepo := store.NewIntentV3DiscoveryResultRepository(ps.DB)
	svc := intentv3.NewService(intentRepo, discRepo)

	dr, err := svc.DiscoverIntent(projectID(projectRoot), content)
	if err != nil {
		return nil, fmt.Errorf("discover intent: %w", err)
	}
	return map[string]any{
		"id":              dr.ID,
		"project_id":      dr.ProjectID,
		"raw_input":       dr.RawInput,
		"detected_intent": dr.DetectedIntent,
		"classification":  dr.Classification,
		"objectives":      dr.Objectives,
		"restrictions":    dr.Restrictions,
		"preferences":     dr.Preferences,
		"expectations":    dr.Expectations,
		"gaps":            dr.Gaps,
		"questions":       dr.Questions,
		"created_at":      formatTime(dr.CreatedAt),
	}, nil
}

func HandleListDiscoveryResults(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, fmt.Errorf("resolve project root: %w", err)
	}
	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	intentRepo := store.NewIntentV3Repository(ps.DB)
	discRepo := store.NewIntentV3DiscoveryResultRepository(ps.DB)
	svc := intentv3.NewService(intentRepo, discRepo)

	list, err := svc.ListDiscoveryResults(projectID(projectRoot))
	if err != nil {
		return nil, fmt.Errorf("list discovery results: %w", err)
	}
	items := make([]map[string]any, 0, len(list))
	for _, dr := range list {
		items = append(items, map[string]any{
			"id":              dr.ID,
			"detected_intent": dr.DetectedIntent,
			"classification":  dr.Classification,
			"created_at":      formatTime(dr.CreatedAt),
		})
	}
	return map[string]any{"items": items, "count": len(items)}, nil
}
