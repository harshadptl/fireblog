package auth

import "testing"

func TestSignerRoundTrip(t *testing.T) {
	s := NewSigner("secret")
	signed, err := s.Sign(OAuthState{State: "abc", Next: "/x"})
	if err != nil {
		t.Fatal(err)
	}
	var got OAuthState
	if err := s.Verify(signed, &got); err != nil || got.State != "abc" || got.Next != "/x" {
		t.Fatalf("verify: %v %+v", err, got)
	}
	if err := NewSigner("other").Verify(signed, &got); err == nil {
		t.Fatal("signature from another key must not verify")
	}
	if err := s.Verify(signed+"x", &got); err == nil {
		t.Fatal("tampered value must not verify")
	}
}

func TestSafeNext(t *testing.T) {
	for in, want := range map[string]string{
		"":                       "/",
		"/posts/x":               "/posts/x",
		"//evil.example":         "/",
		"/\\evil.example":        "/",
		"https://evil.example":   "/",
		"javascript:alert(1)":    "/",
		"/ok?next=1#comments":    "/ok?next=1#comments",
		"/bad\r\nSet-Cookie: x=": "/",
	} {
		if got := SafeNext(in); got != want {
			t.Errorf("SafeNext(%q) = %q, want %q", in, got, want)
		}
	}
}
