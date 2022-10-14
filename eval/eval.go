package eval

import (
	"errors"
	"fmt"
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

// operator order:
// mult and div from left to right
// add and sub from left to right
func EvalExpression(s string) (int64, error) {
	items := split(s)
	// fmt.Println(items)
	ans, err := eval(items)
	if err != nil {
		return 0, err
	}
	return ans, nil
}

func eval(tokens []string) (int64, error) {
	// TODO: refactor so corner case special handling not necessary by virtue of way processed
	// TODO: still use recursive, or convert to iterative? (stack of stacks?)
	// CORNER CASE: starts with "(" not a number+operator pair
	if len(tokens) > 0 && tokens[0] == "(" {
		stub := []string{"0", "+"}
		// fix by padding with number+operator that the code below expects in order to process
		tokens = append(stub, tokens...)
	} else if len(tokens) > 1 && tokens[0] == "-" && tokens[1] == "(" { // CORNER CASE 2: unary negative in front of parens
		stub := []string{"-1", "*"}
		// fix by padding with number+operator that the code below expects in order to process
		tokens = append(stub, tokens[1:]...)
	}

	num, tokens, err := consumeNumber(tokens)
	// fmt.Println("eval", num, tokens, err)

	if err != nil {
		return 0, err
	}
	stack := make([]int64, 0, len(tokens))
	stack = append(stack, num)
	for {
		if len(tokens) < 2 {
			break
		}
		op := tokens[0]
		if tokens[1] == "(" {
			// find matching ")" and then recursively solve that sub problem
			lastCloseParenIdx := -1
			depth := 0
			for i := 2; i < len(tokens); i++ {
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
			num, err = eval(tokens[2:lastCloseParenIdx])
			tokens = tokens[lastCloseParenIdx+1:] // TODO: is this correct?
		} else {
			num, tokens, err = consumeNumber(tokens[1:])
		}
		if err != nil {
			return 0, err
		}
		switch op {
		case "*":
			top := stack[len(stack)-1]
			stack = stack[:len(stack)-1]   // pop
			stack = append(stack, top*num) // push
		case "/":
			top := stack[len(stack)-1]
			stack = stack[:len(stack)-1]   // pop
			stack = append(stack, top/num) // push
		case "+":
			stack = append(stack, num)
		case "-":
			stack = append(stack, -1*num)
		default:
			return 0, fmt.Errorf("unsupported operator: %s", op)
		}
	}

	if len(tokens) != 0 {
		return 0, fmt.Errorf("dangling token(s): %v", tokens)
	}

	val := int64(0)
	for _, st := range stack {
		val += st
	}
	return val, nil
}

// consume number from start of tokens, handling optional minus(-) sign in front
// NOTE: not allowing mulitple minuses in a row (but calling code can handle "2 - -1" as the first minus will be consumed an operator first)
// nor allowing unary "+" as that is implied.
func consumeNumber(tokens []string) (int64, []string, error) {
	if len(tokens) == 0 {
		return 0, nil, errors.New("no more tokens")
	}
	sign := int64(1)
	numCandidate := tokens[0]
	remainingTokens := tokens[1:]
	if tokens[0] == "-" {
		if len(tokens) < 2 {
			return 0, nil, errors.New("no more tokens after '-'")
		}
		sign = -1
		numCandidate = tokens[1]
		remainingTokens = tokens[2:]
	}
	if num, err := ParseHexOrDec(numCandidate); err == nil {
		return sign * num, remainingTokens, nil
	} else {
		return 0, nil, err
	}
}

// TODO: handle 1+-1, 3*1+-1 etc. in other words: handling negative sign as part of the number
// TODO: one approach is to consume subsequents that just change sign and default to addition after?
// TODO: handle ^ and << >>
// TODO: also logical operators? ^xor and exponent^ would conflict....
// TODO: maybe more basic math for expression and a more bit math dedicated eval for other scenarios?
func split(s string) []string {
	items := []string{}
	curToken := ""
	for _, r := range s {
		if r == ' ' || r == '+' || r == '-' || r == '*' || r == '/' || r == '(' || r == ')' {
			if len(curToken) > 0 {
				// flush accumulated token if any data
				items = append(items, curToken)
				curToken = ""
			}
			if r != ' ' {
				items = append(items, string(r))
			}
		} else { // assume number as constraints say will be a valid exp
			curToken += string(r)
		}
	}

	if len(curToken) > 0 {
		// flush accumulated token if any data
		items = append(items, curToken)
	}

	return items
}
