package ronri

import (
	"testing"
)

func TestEval(t *testing.T) {
	tests := []struct {
		expression string
		expected   bool
	}{
		{`((active && logged == true) || !undefined)`, false},
		//{"(active && logged) || undefined", false},
		//{"active && logged", false},
		//{"active", true},
	}
	environment := NewContext()
	environment.Set("active", true)
	environment.Set("logged", false)
	for ix, test := range tests {
		result, err := Eval(test.expression, environment)
		if err != nil {
			t.Fatalf("error '%s' in expression[%d]: %s", err, ix, test.expression)
		}
		if result != test.expected {
			t.Fatalf("expected '%t', got '%t' on expression[%d]: %s", test.expected, result, ix, test.expression)
		}
	}
}
