// Package context implements the Context Engine — approved context
// management, context views, FTS-backed search, and delivery (L0-L4).
// It enforces v4 Principle 6: "El contexto siempre se deriva de
// entidades aprobadas."
//
// Main types: AuthorityService, Builder, DeliveryEngine, Registry.
// Main entry: AuthorityService.Add for deduplicated approved context insertion.
package context
