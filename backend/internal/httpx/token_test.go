package httpx

import "testing"

func TestTokenRoundtrip(t *testing.T) {
	tok := signToken("secret", "photographer")
	u, ok := parseToken("secret", tok)
	if !ok || u != "photographer" {
		t.Fatalf("%s %v", u, ok)
	}
	if _, ok := parseToken("other", tok); ok {
		t.Fatal("forged")
	}
}
