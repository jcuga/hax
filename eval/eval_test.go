package eval

import (
	"testing"
)

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
		testCase{input: "1+1", expectedVal: 2, expectedErr: nil},
		testCase{input: "  1    +   1 ", expectedVal: 2, expectedErr: nil},
		testCase{input: "1+2+3-4", expectedVal: 2, expectedErr: nil},
		testCase{input: "1+(2*3)", expectedVal: 7, expectedErr: nil},
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
