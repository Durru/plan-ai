package knowledge

import (
	"strings"

	"github.com/Durru/plan-ai/internal/domain"
)

// Classify assigns a category to a knowledge topic using deterministic
// keyword rules. The rules are intentionally simple, ordered, and not
// mutually exclusive - the first match wins. When no rule matches the
// topic is reported as CategoryGeneral.
func Classify(topic string) domain.KnowledgeCategory {
	lower := strings.ToLower(strings.TrimSpace(topic))
	if lower == "" {
		return domain.KnowledgeCategoryGeneral
	}

	switch {
	case containsAny(lower, "postgres", "postgresql", "mysql", "mariadb", "sqlite", "sql server", "mongodb", "redis", "dynamodb", "cassandra", "bigquery", "snowflake", "supabase", "prisma", "drizzle", "knex", "sequelize", "typeorm"):
		return domain.KnowledgeCategoryDatabase
	case containsAny(lower, "oauth", "oidc", "jwt", "auth", "login", "signin", "sign-in", "session", "cookie", "rbac", "abac", "passkey", "webauthn", "sso", "saml", "magic link"):
		return domain.KnowledgeCategoryAuthentication
	case containsAny(lower, "billing", "stripe", "paddle", "subscription", "invoice", "invoicing", "payment", "checkout", "pricing", "usage based", "metering"):
		return domain.KnowledgeCategoryBilling
	case containsAny(lower, "react", "next.js", "nextjs", "vue", "nuxt", "svelte", "sveltekit", "remix", "astro", "solid", "frontend", "ui", "ux", "css", "tailwind", "shadcn", "radix", "design system"):
		return domain.KnowledgeCategoryFrontend
	case containsAny(lower, "backend", "api", "rest", "graphql", "trpc", "grpc", "server", "endpoint", "controller", "service layer", "express", "fastify", "nestjs", "hono", "django", "fastapi", "spring", "rails", "laravel", "gin", "fiber", "echo"):
		return domain.KnowledgeCategoryBackend
	case containsAny(lower, "security", "xss", "csrf", "sql injection", "encryption", "vulnerab", "threat model", "owasp", "tls", "ssl", "secret rotation", "vault"):
		return domain.KnowledgeCategorySecurity
	case containsAny(lower, "deploy", "deployment", "ci/cd", "ci ", " cd ", "github actions", "vercel", "netlify", "fly.io", "render", "railway", "docker", "kubernetes", "k8s", "helm", "terraform", "pulumi", "ansible"):
		return domain.KnowledgeCategoryDeployment
	case containsAny(lower, "architecture", "design pattern", "hexagonal", "clean architecture", "ddd", "domain driven", "event driven", "cqrs", "microservice", "monolith", "modular monolith", "ports and adapters", "screaming architecture"):
		return domain.KnowledgeCategoryArchitecture
	case containsAny(lower, "testing", "test", "tdd", "bdd", "coverage", "e2e", "end to end", "integration test", "unit test", "playwright", "cypress", "vitest", "jest"):
		return domain.KnowledgeCategoryTesting
	case containsAny(lower, "mcp", "model context protocol", "mcp server", "mcp client"):
		return domain.KnowledgeCategoryMCP
	case containsAny(lower, "agent", "agents", "autonomous", "tool use", "function calling", "orchestrator"):
		return domain.KnowledgeCategoryAgents
	case containsAny(lower, "ai", "llm", "gpt", "claude", "embedding", "vector", "rag", "prompt", "fine tune", "fine-tuning", "transformer"):
		return domain.KnowledgeCategoryAI
	case containsAny(lower, "devops", "monitoring", "observability", "logging", "tracing", "alerting", "prometheus", "grafana", "datadog", "sentry", "opentelemetry"):
		return domain.KnowledgeCategoryDevops
	case containsAny(lower, "integration", "webhook", "sync", "import", "export", "etl", "connector"):
		return domain.KnowledgeCategoryIntegration
	default:
		return domain.KnowledgeCategoryGeneral
	}
}

func containsAny(s string, terms ...string) bool {
	for _, term := range terms {
		if term == "" {
			continue
		}
		if strings.Contains(s, term) {
			return true
		}
	}
	return false
}
