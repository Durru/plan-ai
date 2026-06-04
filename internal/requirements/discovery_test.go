package requirements

import "testing"

func TestDiscoverEcommerceCandidates(t *testing.T) {
	candidates := Discover("project", "quiero ecommerce")
	for _, expected := range []string{"cart", "checkout", "coupons", "SEO", "blog", "analytics"} {
		if !hasCandidate(candidates, expected) {
			t.Fatalf("missing %s in %#v", expected, candidates)
		}
	}
}

func hasCandidate(items []Candidate, name string) bool {
	for _, item := range items {
		if item.Name == name && item.State == StateCandidate {
			return true
		}
	}
	return false
}
