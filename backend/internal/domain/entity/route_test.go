package entity

import "testing"

func TestNormalizeMethod(t *testing.T) {
	tests := map[string]string{
		"":      HTTPMethodAll,
		"   ":   HTTPMethodAll,
		"get":   "GET",
		" Post": "POST",
		"PUT":   "PUT",
	}
	for in, want := range tests {
		got := Route{Method: in}.NormalizeMethod()
		if got != want {
			t.Errorf("NormalizeMethod(%q) = %q; want %q", in, got, want)
		}
	}
}
