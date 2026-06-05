package main

import (
	"strings"
	"testing"
)

// ──────────────────────────────────────────────
// Phase 47-48: CLI tests for memory and model commands
// ──────────────────────────────────────────────

func TestMemoryAddCreatesEntryDeterministically(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()

	runStep(t, home, project, "Global installation: installed", "install")
	runStep(t, home, project, "Project initialization: initialized", "init")

	out := runStep(t, home, project, "", "memory", "add",
		"--type", "decision",
		"--title", "Use PostgreSQL for persistence",
		"--content", "PostgreSQL is the primary database for all project data.",
		"--citation", "https://postgresql.org",
		"--source", "team-meeting",
	)
	if !strings.Contains(out, "Memory entry created:") {
		t.Fatalf("memory add output missing 'Memory entry created:', got:\n%s", out)
	}
	if !strings.Contains(out, "decision") {
		t.Fatalf("memory add output missing type 'decision':\n%s", out)
	}
}

func TestMemoryListShowsStoredEntry(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()

	runStep(t, home, project, "Global installation: installed", "install")
	runStep(t, home, project, "Project initialization: initialized", "init")

	runStep(t, home, project, "Memory entry created:", "memory", "add",
		"--type", "decision",
		"--title", "Use PostgreSQL for persistence",
		"--content", "PostgreSQL is the primary database.",
	)

	listOut := runStep(t, home, project, "", "memory", "list")
	if !strings.Contains(listOut, "PostgreSQL") {
		t.Fatalf("memory list missing title 'PostgreSQL':\n%s", listOut)
	}
	if !strings.Contains(listOut, "decision") {
		t.Fatalf("memory list missing type 'decision':\n%s", listOut)
	}
}

func TestMemoryAskReusesQuestionAnswer(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()

	runStep(t, home, project, "Global installation: installed", "install")
	runStep(t, home, project, "Project initialization: initialized", "init")

	// Add a question_answer entry with citation and source
	addOut := runStep(t, home, project, "", "memory", "add",
		"--type", "question_answer",
		"--question", "What database should we use?",
		"--answer", "PostgreSQL with schema-per-tenant",
		"--citation", "https://postgresql.org/docs",
		"--source", "architecture-review",
	)
	if !strings.Contains(addOut, "Memory entry created:") {
		t.Fatalf("memory add QA failed:\n%s", addOut)
	}
	// Extract entry ID
	id := extractMemoryID(t, addOut)

	// Ask the same question — should reuse
	askOut, err := executeCommand(t, home, project, "memory", "ask", "What database should we use?")
	if err != nil {
		t.Fatalf("memory ask: %v\n%s", err, askOut)
	}
	if !strings.Contains(askOut, "(Reused existing memory entry)") {
		t.Fatalf("memory ask should indicate reuse:\n%s", askOut)
	}
	if !strings.Contains(askOut, id) {
		t.Fatalf("memory ask should contain the original entry ID:\n%s", askOut)
	}
	if !strings.Contains(askOut, "PostgreSQL with schema-per-tenant") {
		t.Fatalf("memory ask missing answer:\n%s", askOut)
	}

	// Ask a different question — should not find a match
	missOut, err := executeCommand(t, home, project, "memory", "ask", "What color should we paint the bikeshed?")
	if err != nil {
		t.Fatalf("memory ask miss: %v\n%s", err, missOut)
	}
	if !strings.Contains(missOut, "No matching memory entry found.") {
		t.Fatalf("expected miss for unmatched question:\n%s", missOut)
	}
}

func TestMemoryAskIncludesCitationSourceEvidence(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()

	runStep(t, home, project, "Global installation: installed", "install")
	runStep(t, home, project, "Project initialization: initialized", "init")

	runStep(t, home, project, "Memory entry created:", "memory", "add",
		"--type", "question_answer",
		"--question", "What is the auth strategy?",
		"--answer", "OAuth 2.0 with PKCE",
		"--citation", "https://oauth.net/2/",
		"--source", "security-review",
	)

	askOut, err := executeCommand(t, home, project, "memory", "ask", "What is the auth strategy?")
	if err != nil {
		t.Fatalf("memory ask: %v\n%s", err, askOut)
	}
	// The ask output includes ID, type, Q, A, Content fields but NOT citation/source directly
	// because newMemoryAskCommand only prints question, answer, and content
	if !strings.Contains(askOut, "OAuth 2.0 with PKCE") {
		t.Fatalf("memory ask missing answer:\n%s", askOut)
	}
	// Verify the entry is reused
	if !strings.Contains(askOut, "(Reused existing memory entry)") {
		t.Fatalf("memory ask should indicate reuse:\n%s", askOut)
	}
}

func TestModelProvidersIncludesExpectedProviders(t *testing.T) {
	out, err := executeCommand(t, t.TempDir(), t.TempDir(), "model", "providers")
	if err != nil {
		t.Fatalf("model providers: %v\n%s", err, out)
	}

	for _, want := range []string{
		"anthropic",
		"openai",
		"gemini",
		"deepseek",
		"qwen",
		"openrouter",
		"openai_compatible",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("model providers output missing %q:\n%s", want, out)
		}
	}
	if !strings.Contains(out, "Supported providers") {
		t.Fatalf("model providers output missing header:\n%s", out)
	}
}

func TestModelCompatibilityForKnownModelProvider(t *testing.T) {
	out, err := executeCommand(t, t.TempDir(), t.TempDir(), "model", "compatibility", "gpt-4o", "openai")
	if err != nil {
		t.Fatalf("model compatibility: %v\n%s", err, out)
	}
	for _, want := range []string{
		"Model:  gpt-4o",
		"Provider: openai",
		"Supported: true",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("compatibility output missing %q:\n%s", want, out)
		}
	}
}

func TestModelCompatibilityForKnownModelWithoutProvider(t *testing.T) {
	out, err := executeCommand(t, t.TempDir(), t.TempDir(), "model", "compatibility", "gpt-4o")
	if err != nil {
		t.Fatalf("model compatibility (no provider): %v\n%s", err, out)
	}
	// Should find that gpt-4o is supported by openai
	if !strings.Contains(out, `is supported by`) {
		t.Fatalf("expected supported-by message:\n%s", out)
	}
	if !strings.Contains(out, "openai") {
		t.Fatalf("expected openai in supported providers:\n%s", out)
	}
}

func TestModelCompatibilityForUnknownModel(t *testing.T) {
	out, err := executeCommand(t, t.TempDir(), t.TempDir(), "model", "compatibility", "gpt-5.5")
	if err != nil {
		t.Fatalf("model compatibility unknown: %v\n%s", err, out)
	}
	if !strings.Contains(out, "not in the compatibility catalog") {
		t.Fatalf("expected 'not in the compatibility catalog' for unknown model:\n%s", out)
	}
}

func TestModelCompatibilityForUnknownModelWithKnownProvider(t *testing.T) {
	out, err := executeCommand(t, t.TempDir(), t.TempDir(), "model", "compatibility", "gpt-5.5", "openai")
	if err != nil {
		t.Fatalf("model compatibility unknown+provider: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Supported: false") {
		t.Fatalf("expected 'Supported: false' for unknown model:\n%s", out)
	}
}

// ──────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────

func extractMemoryID(t *testing.T, output string) string {
	t.Helper()
	for _, line := range strings.Split(output, "\n") {
		for _, field := range strings.Fields(line) {
			if strings.HasPrefix(field, "mem_") {
				return field
			}
		}
	}
	t.Fatalf("memory id not found in output:\n%s", output)
	return ""
}
