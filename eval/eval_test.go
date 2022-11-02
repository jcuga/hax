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
		testCase{input: "*", expectedVal: []string{"*"}},
		testCase{input: "**", expectedVal: []string{"**"}},
		testCase{input: "* **", expectedVal: []string{"**", "*"}},
		testCase{input: " ****", expectedVal: []string{"**", "**"}},
		testCase{input: "* * * * *", expectedVal: []string{"**", "**", "*"}},
		testCase{input: "-1", expectedVal: []string{"-", "1"}},
		testCase{input: "-234", expectedVal: []string{"-", "234"}},
		testCase{input: "1+2*3/4", expectedVal: []string{"1", "+", "2", "*", "3", "/", "4"}},
		testCase{input: "1+2**3*4/4", expectedVal: []string{"1", "+", "2", "**", "3", "*", "4", "/", "4"}},
		testCase{input: " - 234  ", expectedVal: []string{"-", "234"}},
		testCase{input: " 1+ 2* 3/ 4 ", expectedVal: []string{"1", "+", "2", "*", "3", "/", "4"}},
		testCase{input: "(1+2)*(3/4)", expectedVal: []string{"(", "1", "+", "2", ")", "*", "(", "3", "/", "4", ")"}},
		testCase{input: "(~1+2)*~(3/4)", expectedVal: []string{"(", "~", "1", "+", "2", ")", "*", "~", "(", "3", "/", "4", ")"}},
		testCase{input: "-(1+2)*-(3/4)", expectedVal: []string{"-", "(", "1", "+", "2", ")", "*", "-", "(", "3", "/", "4", ")"}},
		testCase{input: "---3", expectedVal: []string{"-", "-", "-", "3"}},
	}
	for _, c := range cases {
		if val := tokenize(c.input); !strSliceEqual(c.expectedVal, val) {
			t.Errorf("Unexpected value, input: %v, expect: %v, got: %v", c.input, c.expectedVal, val)
		}
	}
}

func Test_Eval_ParseHexDecOrBin(t *testing.T) {
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
		testCase{input: "xab", expectedVal: 171, expectedErr: nil},
		testCase{input: "\\xAbCd", expectedVal: 43981, expectedErr: nil},
		testCase{input: "\\x0a", expectedVal: 10, expectedErr: nil},
		testCase{input: "0x0B", expectedVal: 11, expectedErr: nil},
		testCase{input: "-\\x0B", expectedVal: -11, expectedErr: nil},
		testCase{input: "B1", expectedVal: 1, expectedErr: nil},
		testCase{input: "0b11", expectedVal: 3, expectedErr: nil},
		testCase{input: "\\b101", expectedVal: 5, expectedErr: nil},
		testCase{input: "-b101", expectedVal: -5, expectedErr: nil},
		testCase{input: "B101", expectedVal: 5, expectedErr: nil},
		testCase{input: "b0000 0000 0000 0011", expectedVal: 3, expectedErr: nil},
		// TODO: more cases here... including error cases
	}
	for _, c := range cases {
		val, err := ParseHexDecOrBin(c.input)
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

		// Unary bitwise not/flip bits
		testCase{input: "~1", expectedVal: -2, expectedErr: nil},
		testCase{input: "~2", expectedVal: -3, expectedErr: nil},
		testCase{input: "~-~1", expectedVal: -3, expectedErr: nil},
		testCase{input: "~-~-1", expectedVal: -1, expectedErr: nil},
		testCase{input: "~-~-~1", expectedVal: -4, expectedErr: nil},
		testCase{input: "~~1", expectedVal: 1, expectedErr: nil},
		testCase{input: "~~-1", expectedVal: -1, expectedErr: nil},
		testCase{input: "-~~-1", expectedVal: 1, expectedErr: nil},
		testCase{input: "-~-~-1", expectedVal: 1, expectedErr: nil},
		testCase{input: "---~~1", expectedVal: -1, expectedErr: nil},
		testCase{input: "~~---1", expectedVal: -1, expectedErr: nil},
		testCase{input: "~(2*3)", expectedVal: -7, expectedErr: nil},
		testCase{input: "~(2*3)*4", expectedVal: -28, expectedErr: nil},
		testCase{input: "~(2*3)*~4", expectedVal: 35, expectedErr: nil},
		testCase{input: "-~(2*3)*~4", expectedVal: -35, expectedErr: nil},
		testCase{input: "~(~(2*3)*~4)", expectedVal: -36, expectedErr: nil},
		testCase{input: "~(~(2*~3)*~4)", expectedVal: 34, expectedErr: nil},
		testCase{input: "~(~(~2*~3)*~4)", expectedVal: -66, expectedErr: nil},
		testCase{input: "2**~-3", expectedVal: 4, expectedErr: nil},
		testCase{input: "~2**~-3", expectedVal: -5, expectedErr: nil},
		testCase{input: "~(~2**~-3)", expectedVal: 4, expectedErr: nil},
		testCase{input: "-~(~2**~-3)", expectedVal: -4, expectedErr: nil},
		testCase{input: "~-(~2**~-3)", expectedVal: -6, expectedErr: nil},

		testCase{input: "1+1", expectedVal: 2, expectedErr: nil},
		testCase{input: "1--1", expectedVal: 2, expectedErr: nil},
		testCase{input: "1+-1", expectedVal: 0, expectedErr: nil},
		testCase{input: "1+--1", expectedVal: 2, expectedErr: nil},
		testCase{input: "1---1", expectedVal: 0, expectedErr: nil},
		testCase{input: "  1    +   1 ", expectedVal: 2, expectedErr: nil},
		testCase{input: "1+2+3-4", expectedVal: 2, expectedErr: nil},
		testCase{input: "1+(2*3)", expectedVal: 7, expectedErr: nil},
		testCase{input: "1,000+(2_00*3 0)", expectedVal: 7000, expectedErr: nil},
		testCase{input: "1+2*3-4", expectedVal: 3, expectedErr: nil},
		testCase{input: "1+2*3**2-4", expectedVal: 15, expectedErr: nil},
		testCase{input: "2**((5+3)/2)", expectedVal: 16, expectedErr: nil},
		testCase{input: "-2**((5+3)/2)", expectedVal: -16, expectedErr: nil},
		testCase{input: "-2**4", expectedVal: -16, expectedErr: nil},
		testCase{input: "(-2)**4", expectedVal: 16, expectedErr: nil},
		testCase{input: "-2**-4", expectedVal: 0, expectedErr: nil},
		testCase{input: "(-2)**-4", expectedVal: 0, expectedErr: nil},
		testCase{input: "(2)", expectedVal: 2, expectedErr: nil},
		testCase{input: "-(2)", expectedVal: -2, expectedErr: nil},
		testCase{input: "-(-2)", expectedVal: 2, expectedErr: nil},
		testCase{input: "-(1+2*3**2-4)", expectedVal: -15, expectedErr: nil},
		testCase{input: "-(-(1+2*3**2-4))", expectedVal: 15, expectedErr: nil},
		testCase{input: "-(-(-(1+2*3**2-4)))", expectedVal: -15, expectedErr: nil},
		testCase{input: "-(-(-(1+2*-3**2-4)))", expectedVal: 21, expectedErr: nil},
		testCase{input: "-(-(-(1+2*-3**2-4)))-(-(-(1+2*3**2-4)))", expectedVal: 6, expectedErr: nil},
		testCase{input: "-(-(-(1+2*-3**2-4))) + - (-(-(1+2*3**2-4)))", expectedVal: 6, expectedErr: nil},
		testCase{input: "-(-(-(1+2*-3**2-4)))+(-(-(1+2*3**2-4)))", expectedVal: 36, expectedErr: nil},
		testCase{input: "b1000*2**2**2/2+3", expectedVal: 67, expectedErr: nil},
		// allow space or underscore or comma to format numbers:
		testCase{input: "b1_0_0_0*2**2**2/2+3", expectedVal: 67, expectedErr: nil},
		testCase{input: "b10 00*2**2**2/2+3", expectedVal: 67, expectedErr: nil},
		testCase{input: "b10,00*2**2**2/2+3", expectedVal: 67, expectedErr: nil},
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
