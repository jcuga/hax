package eval

import (
	"testing"
)

func strSliceEqual(a, b []string) bool {
	if a == nil && b == nil {
		return true
	} else if a == nil || b == nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func Test_tokenize(t *testing.T) {
	type testCase struct {
		input       string
		expectedVal []string
	}
	cases := []testCase{
		testCase{input: "", expectedVal: []string{}},
		testCase{input: "0", expectedVal: []string{"0"}},
		testCase{input: "-1", expectedVal: []string{"-", "1"}},
		testCase{input: "-234", expectedVal: []string{"-", "234"}},
		testCase{input: "1+2*3/4", expectedVal: []string{"1", "+", "2", "*", "3", "/", "4"}},
		testCase{input: " - 234  ", expectedVal: []string{"-", "234"}},
		testCase{input: " 1+ 2* 3/ 4 ", expectedVal: []string{"1", "+", "2", "*", "3", "/", "4"}},
		testCase{input: "(1+2)*(3/4)", expectedVal: []string{"(", "1", "+", "2", ")", "*", "(", "3", "/", "4", ")"}},
		testCase{input: "-(1+2)*-(3/4)", expectedVal: []string{"-", "(", "1", "+", "2", ")", "*", "-", "(", "3", "/", "4", ")"}},
		testCase{input: "---3", expectedVal: []string{"-", "-", "-", "3"}},
	}
	for _, c := range cases {
		if val := tokenize(c.input); !strSliceEqual(c.expectedVal, val) {
			t.Errorf("Unexpected value, input: %v, expect: %v, got: %v", c.input, c.expectedVal, val)
		}
	}
}

func Test_Eval_ParseHexOrDec(t *testing.T) {
	type testCase struct {
		input       string
		expectedVal int64
		expectedErr error
	}
	cases := []testCase{
		testCase{input: "0", expectedVal: 0, expectedErr: nil},
		testCase{input: "1", expectedVal: 1, expectedErr: nil},
		testCase{input: "-1", expectedVal: -1, expectedErr: nil},
		testCase{input: "\\xA", expectedVal: 10, expectedErr: nil},
		testCase{input: "\\xb", expectedVal: 11, expectedErr: nil},
		testCase{input: "0a", expectedVal: 10, expectedErr: nil},
		testCase{input: "0B", expectedVal: 11, expectedErr: nil},
		testCase{input: "-0B", expectedVal: -11, expectedErr: nil},
		// TODO: more cases here... including error cases
	}
	for _, c := range cases {
		val, err := ParseHexOrDec(c.input)
		if val != c.expectedVal {
			t.Errorf("Unexpected value, input: %v, expect: %v, got: %v", c.input, c.expectedVal, val)
		}
		if err != c.expectedErr {
			t.Errorf("Unexpected err, input: %v, expect: %v, got: %v", c.input, c.expectedErr, err)
		}
	}
}

func Test_Eval_EvalExpression(t *testing.T) {
	type testCase struct {
		input       string
		expectedVal int64
		expectedErr error
	}
	cases := []testCase{
		testCase{input: "0", expectedVal: 0, expectedErr: nil},
		testCase{input: "  0   ", expectedVal: 0, expectedErr: nil},
		testCase{input: "-1", expectedVal: -1, expectedErr: nil},
		testCase{input: "1+1", expectedVal: 2, expectedErr: nil},
		testCase{input: "  1    +   1 ", expectedVal: 2, expectedErr: nil},
		testCase{input: "1+2+3-4", expectedVal: 2, expectedErr: nil},
		testCase{input: "1+(2*3)", expectedVal: 7, expectedErr: nil},
		testCase{input: "1+2*3-4", expectedVal: 3, expectedErr: nil},
		testCase{input: "1+2*3^2-4", expectedVal: 15, expectedErr: nil},
		testCase{input: "2^((5+3)/2)", expectedVal: 16, expectedErr: nil},
		testCase{input: "-2^((5+3)/2)", expectedVal: -16, expectedErr: nil},
		testCase{input: "-2^4", expectedVal: -16, expectedErr: nil},
		testCase{input: "(-2)^4", expectedVal: 16, expectedErr: nil},
		testCase{input: "-2^-4", expectedVal: 0, expectedErr: nil},
		testCase{input: "(-2)^-4", expectedVal: 0, expectedErr: nil},
		testCase{input: "(2)", expectedVal: 2, expectedErr: nil},
		testCase{input: "-(2)", expectedVal: -2, expectedErr: nil},
		testCase{input: "-(-2)", expectedVal: 2, expectedErr: nil},
		testCase{input: "-(1+2*3^2-4)", expectedVal: -15, expectedErr: nil},
		testCase{input: "-(-(1+2*3^2-4))", expectedVal: 15, expectedErr: nil},
		testCase{input: "-(-(-(1+2*3^2-4)))", expectedVal: -15, expectedErr: nil},
		testCase{input: "-(-(-(1+2*-3^2-4)))", expectedVal: 21, expectedErr: nil},
		testCase{input: "-(-(-(1+2*-3^2-4)))-(-(-(1+2*3^2-4)))", expectedVal: 6, expectedErr: nil},
		testCase{input: "-(-(-(1+2*-3^2-4))) + - (-(-(1+2*3^2-4)))", expectedVal: 6, expectedErr: nil},
		testCase{input: "-(-(-(1+2*-3^2-4)))+(-(-(1+2*3^2-4)))", expectedVal: 36, expectedErr: nil},

		// TODO: some hex numbers sprinkled in using various formats \x 0x x ab, etc...

		// TODO: more cases here... including error cases
	}
	for _, c := range cases {
		val, err := EvalExpression(c.input)
		if val != c.expectedVal {
			t.Errorf("Unexpected value, input: %v, expect: %v, got: %v", c.input, c.expectedVal, val)
		}
		if err != c.expectedErr {
			t.Errorf("Unexpected err, input: %v, expect: %v, got: %v", c.input, c.expectedErr, err)
		}
	}
}
