package eval

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

func ParseHexOrDec(input string) (int64, error) {
	if len(input) == 0 {
		return 0, nil
	}
	// If sarts with "0", "x", or "0x", "\x" OR has a-fA-F, then interpret as hex
	// also interpret as hex if has spaces between nubmers which would be if one
	// copy-pasted a value from a previous run's output (ex: "AA BB CC")
	input = strings.ToLower(input)
	input = strings.TrimSpace(input)
	if strings.HasPrefix(input, "0") || strings.HasPrefix(input, "0x") || strings.HasPrefix(input, "x") ||
		strings.HasPrefix(input, "\\x") || strings.ContainsAny(input, "abcdef ") {
		// assume hex
		// trim off leading "0x" or "x" if found (could be just 0 in front)
		xIndex := strings.Index(input, "x")
		if xIndex != -1 {
			input = input[xIndex+1:]
		}
		// remove any spaces between bytes ("AA BB" --> "AABB")
		input = strings.Replace(input, " ", "", -1)
		return strconv.ParseInt(input, 16, 64)
	}
	return strconv.ParseInt(input, 10, 64)
}

// Returns parsed num, number of tokens consumed in parsing, error.
func parseNumWithPossibleUnaryMinus(tokens []string) (int64, int, error) {
	if len(tokens) == 0 {
		return 0, 0, errors.New("no tokens")
	}
	if len(tokens) > 1 && tokens[0] == "-" {
		num, err := ParseHexOrDec(tokens[1])
		return -1 * num, 2, err
	}
	num, err := ParseHexOrDec(tokens[0])
	return num, 1, err
}

func EvalExpression(s string) (int64, error) {
	tokens := tokenize(s)
	ans, err := eval(tokens)
	if err != nil {
		return 0, err
	}
	return ans, nil
}

func eval(tokens []string) (int64, error) {
	if len(tokens) == 0 {
		return 0, errors.New("not enough tokens")

	}
	if len(tokens) == 1 { // trivial single number
		return ParseHexOrDec(tokens[0])
	}
	// First pass--solve parenthesis recursively, stack everything else for subsequent processing
	reduced := make([]string, 0, len(tokens))
	for {
		if len(tokens) < 1 {
			break
		}
		switch token := tokens[0]; token {
		case "(":
			// find matching ")" and then recursively solve that sub problem
			lastCloseParenIdx := -1
			depth := 0
			for i := 1; i < len(tokens); i++ {
				if tokens[i] == ")" {
					if depth == 0 {
						lastCloseParenIdx = i
						break
					}
					depth--
				} else if tokens[i] == "(" {
					depth++
				}
			}
			if lastCloseParenIdx == -1 {
				return 0, errors.New("no closing ')'")
			}
			num, err := eval(tokens[1:lastCloseParenIdx])
			if err != nil {
				return 0, err
			}
			tokens = tokens[lastCloseParenIdx+1:]
			reduced = append(reduced, strconv.Itoa(int(num)))
		case ")":
			// should always consume matching ")" when we first see the opening "("
			// getting here is a sign of mismatched parentheses
			return 0, errors.New("mismatched ')'")
		default:
			reduced = append(reduced, token)
			tokens = tokens[1:]
		}
	}
	// next pass: solve exponents
	stack1 := make([]string, 0, len(reduced))
	for i := 0; i < len(reduced); i++ {
		switch token := reduced[i]; token {
		case "**":
			if i == len(reduced)-1 {
				return 0, errors.New("dangling '^'")
			}
			if len(stack1) == 0 {
				return 0, errors.New("mismatched '^'")
			}
			// pop exponent's base:
			baseStr := stack1[len(stack1)-1]
			stack1 = stack1[:len(stack1)-1]
			// NOTE: not using parseNumWithPossibleUnaryMinus as exponent comes before unary minus
			// in other words -2^4 is -8 since it's really -(2^4) whereas (-2)^4 is 8...
			base, err := ParseHexOrDec(baseStr)
			if err != nil {
				return 0, err
			}
			// NOTE: handle unary minus
			exponent, tokensConsumed, err := parseNumWithPossibleUnaryMinus(reduced[i+1:])
			if err != nil {
				return 0, err
			}
			// stack result:
			result := strconv.Itoa(int(math.Pow(float64(base), float64(exponent))))
			stack1 = append(stack1, result)
			i += tokensConsumed // skip over consumed exponent number
		default:
			stack1 = append(stack1, token)
		}
	}
	// next pass: solve multiply/divide/add/subtract
	if len(stack1) == 0 {
		return 0, errors.New("missing tokens")
	}
	stack2 := make([]int64, 0, len(stack1))
	// NOTE: handle unary minus
	var num int64
	var err error
	var tokensConsumed int
	num, tokensConsumed, err = parseNumWithPossibleUnaryMinus(stack1)
	stack1 = stack1[tokensConsumed:]
	if err != nil {
		return 0, err
	}
	stack2 = append(stack2, num)

	for {
		if len(stack1) < 2 {
			break
		}
		op := stack1[0]
		if !isOperator(op) {
			return 0, fmt.Errorf("expected operator, got: %s", op)
		}
		stack1 = stack1[1:]
		// NOTE: handle unary minus
		num, tokensConsumed, err = parseNumWithPossibleUnaryMinus(stack1)
		if err != nil {
			return 0, err
		}
		stack1 = stack1[tokensConsumed:]
		switch op {
		case "*":
			if len(stack2) < 1 {
				return 0, fmt.Errorf("operator %s without preceeding number", op)
			}
			top := stack2[len(stack2)-1]
			stack2 = stack2[:len(stack2)-1]  // pop
			stack2 = append(stack2, top*num) // push
		case "/":
			if len(stack2) < 1 {
				return 0, fmt.Errorf("operator %s without preceeding number", op)
			}
			top := stack2[len(stack2)-1]
			stack2 = stack2[:len(stack2)-1]  // pop
			stack2 = append(stack2, top/num) // push
		case "+":
			stack2 = append(stack2, num)
		case "-":
			stack2 = append(stack2, -1*num)
		default:
			return 0, fmt.Errorf("unsupported operator: %s", op)
		}
	}
	if len(stack1) != 0 {
		return 0, fmt.Errorf("dangling token(s): %v", stack1)
	}
	val := int64(0)
	for _, st := range stack2 {
		val += st
	}
	return val, nil
}

func isOperator(s string) bool {
	switch s {
	case "+", "-", "*", "/", "**":
		return true
	default:
		return false
	}
}

func tokenize(s string) []string {
	// first--eliminate ignored chars
	// this allow for writing numbers like:
	// "1,000", "1 000 000", "1_000_000" etc
	var normS strings.Builder
	for _, r := range s {
		if r == ' ' || r == '_' || r == ',' {
			continue
		}
		normS.WriteRune(r)
	}
	s = normS.String()

	tokens := []string{}
	curToken := ""
	for i := 0; i < len(s); i++ {
		// handle 2-char operators like **, <<, >>, etc
		if i < len(s)-1 && isOperator(s[i:i+2]) {
			if len(curToken) > 0 {
				tokens = append(tokens, curToken)
				curToken = ""
			}
			tokens = append(tokens, s[i:i+2])
			i++ // skip 2nd char in addition to normal loop iteration increment of i
		} else if isOperator(string(s[i])) || s[i] == '(' || s[i] == ')' { // check for single digit operator
			if len(curToken) > 0 {
				tokens = append(tokens, curToken)
				curToken = ""
			}
			tokens = append(tokens, string(s[i]))
		} else { // since not ignored char or operator, assume part of a number
			curToken += string(s[i])
		}
	}
	// flush any trailing data
	if len(curToken) > 0 {
		tokens = append(tokens, curToken)
	}
	return tokens
}
