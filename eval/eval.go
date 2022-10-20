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
			tokens = tokens[lastCloseParenIdx+1:] // TODO: is this correct?
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
	// next pass: solve unary minus
	// now should only be numbers and operators--no parentheses
	reduced2 := make([]string, 0, len(reduced))
	for i := 0; i < len(reduced); i++ {
		// only if not last token as we'll use i+1 as a number (we'll catch trailing minus as invalid later on...)
		if reduced[i] == "-" && i < len(reduced)-1 {
			// starting with minus or having one after another operator is considerd a unary minus
			if i == 0 || (i > 0 && isOperator(reduced[i-1])) {
				num, err := ParseHexOrDec(reduced[i+1])
				if err != nil {
					return 0, err
				}
				// NOTE: num itself could be negative which is fine, just flip the sign.
				num *= -1
				reduced[i+1] = strconv.Itoa(int(num))
				continue
			}
		}
		// keep as-is for next pass
		reduced2 = append(reduced2, reduced[i])
	}
	// next pass: solve exponents
	stack1 := make([]string, 0, len(reduced2))
	for i := 0; i < len(reduced2); i++ {
		switch token := reduced2[i]; token {
		case "^":
			if i == len(reduced2)-1 {
				return 0, errors.New("dangling '^'")
			}
			if len(stack1) == 0 {
				return 0, errors.New("mismatched '^'")
			}
			// pop exponent's base:
			baseStr := stack1[len(stack1)-1]
			stack1 = stack1[:len(stack1)-1]
			base, err := ParseHexOrDec(baseStr)
			if err != nil {
				return 0, err
			}
			exponent, err := ParseHexOrDec(reduced2[i+1])
			if err != nil {
				return 0, err
			}
			// stack result:
			result := strconv.Itoa(int(math.Pow(float64(base), float64(exponent))))
			stack1 = append(stack1, result)
			i++ // skip over consumed exponent number
		default:
			stack1 = append(stack1, token)
		}
	}
	// next pass: solve multiply/divide/add/subtract
	if len(stack1) == 0 {
		return 0, errors.New("missing tokens")
	}
	stack2 := make([]int64, 0, len(stack1))
	num, err := ParseHexOrDec(stack1[0])
	if err != nil {
		return 0, err
	}
	stack1 = stack1[1:]
	stack2 = append(stack2, num)

	for {
		if len(stack1) < 2 {
			break
		}
		op := stack1[0]
		if !isOperator(op) {
			return 0, fmt.Errorf("expected operator, got: %s", op)
		}
		num, err := ParseHexOrDec(stack1[1])
		if err != nil {
			return 0, err
		}
		stack1 = stack1[2:]
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
	case "+", "-", "*", "/", "^":
		return true
	default:
		return false
	}
}

func tokenize(s string) []string {
	tokens := []string{}
	curToken := ""
	for _, r := range s {
		if r == ' ' || r == '+' || r == '-' || r == '*' || r == '/' || r == '(' || r == ')' || r == '^' {
			if len(curToken) > 0 {
				// flush accumulated token if any data
				tokens = append(tokens, curToken)
				curToken = ""
			}
			if r != ' ' {
				// operator or parenthesis--add as own token:
				tokens = append(tokens, string(r))
			}
		} else {
			// assume part of a number--accumulate data
			curToken += string(r)
		}
	}
	if len(curToken) > 0 {
		// flush any trailing data
		tokens = append(tokens, curToken)
	}
	return tokens
}
