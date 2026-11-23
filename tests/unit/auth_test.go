package unit

import (
	"testing"

	"redbook-go/internal/auth"
)

func TestPasswordEncodeAndMatch(t *testing.T) {
	encoded := auth.EncodePassword("secret")
	if !auth.MatchesPassword(encoded, "secret") {
		t.Fatal("expected password to match")
	}
	if auth.MatchesPassword(encoded, "wrong") {
		t.Fatal("expected wrong password to fail")
	}
}

func TestTokenRoundTrip(t *testing.T) {
	secret := "01234567890123456789012345678901"
	tokens := auth.NewTokenService(secret, 3600, nil)
	token, err := tokens.CreateToken(42, 3)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := tokens.ParseToken("Bearer " + token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != 42 || claims.TV != 3 {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}
