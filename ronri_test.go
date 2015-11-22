package ronri

import (
	"testing"
)

func TestEval(t *testing.T) {
	tests := []struct {
		expression string
		expected   bool
		mustFail   bool
	}{
		{`active`, true, false},
		{`undefined`, false, true},
		{`kind`, false, true},
		{`active == true`, true, false},
		{`active == false`, false, false},
		{`!active`, false, false},
		{`active == !active`, false, false},
		{`!active == false`, true, false},
		{`!active == !false`, false, false},
		{`!active == !true`, true, false},
		{`_`, false, true},
		{`_ == true`, false, true},
		{`kind == "admin"`, true, false},
		{`uid < 200`, false, false},
		{`uid == "200"`, false, true},
		{`uid == 200`, true, false},
		{`uid <= 200`, true, false},
		{`uid > 200`, false, false},
		{`uid >= 200`, true, false},
		{`active && kind == "admin"`, true, false},
		{`active && kind == "admin" && organization == "imdario"`, true, false},
		{`organization != "dariocc"`, true, false},
		{`kind == "admin" || organization == "dariocc"`, true, false},
		{`kind == "admin" || organization == "dariocc" && !active`, true, false},
		{`kind == "admin" || (organization == "dariocc" && !active)`, true, false},
		{`(kind == "admin" || organization == "dariocc") && !active`, false, false},
		{`kind == "admin" || kind == "operator" && active && organization == "imdario"`, true, false},
		{`kind == "admin" || kind == "operator" && !active || organization != "imdario"`, true, false},
		{`kind == "admin" || kind == "operator" && !active || organization != "imdario"`, true, false},
		{`kind == "user" || kind == "operator" && !active || organization != "imdario"`, false, false},
	}
	context := NewContext()
	context.Set("active", true)
	context.Set("kind", "admin")
	context.Set("organization", "imdario")
	context.Set("uid", 200)
	context.Set("karma", 42.314)
	for ix, test := range tests {
		result, err := Eval(test.expression, context)
		if err != nil && !test.mustFail {
			t.Fatalf("error '%s' in expression[%d]: %s", err, ix, test.expression)
		}
		if result != test.expected {
			t.Fatalf("expected '%t', got '%t' on expression[%d]: %s", test.expected, result, ix, test.expression)
		}
		if test.mustFail {
			if err == nil {
				t.Fatalf("expected error but error not returned on expression[%d]: %s", ix, test.expression)
			} else {
				t.Logf("expression[%d] (%s) expected error: %s", ix, test.expression, err)
			}
		}
	}
}
