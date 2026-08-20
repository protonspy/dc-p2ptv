package version

import "testing"

func TestDescribe(t *testing.T) {
	cases := map[string]struct {
		role string
		want string
	}{
		"role is kept":        {role: "node", want: "node/" + Current},
		"blank role reported": {role: "  ", want: "unknown/" + Current},
		"empty role reported": {role: "", want: "unknown/" + Current},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if got := Describe(c.role); got != c.want {
				t.Errorf("Describe(%q) = %q, want %q", c.role, got, c.want)
			}
		})
	}
}
