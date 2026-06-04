package reference

import "testing"

func TestBuildStripeReferenceClassifiesUX(t *testing.T) {
	ref := Build("project", SourceURL, "https://stripe.com", "Stripe", "")
	if ref.Category != CategoryUX || ref.Status != StatusNeedsReview {
		t.Fatalf("unexpected reference: %#v", ref)
	}
}
