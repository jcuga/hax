package eval

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// TODO: make a string so can print more meaningful errors? Then define consts using appropriate string...
type Token int

const (
	OpenParentheses Token = iota
	CloseParentheses
	Number
	Plus
	Minus
	UnaryBitwiseNot
	Multiply
	Exponent
	Divide
	Modulo
	LeftShift
	RightShift
)

type tokenValue struct {
	token Token
	value int64
}

func ParseHexDecOrBin(input string) (int64, error) {
	if len(input) == 0 {
		return 0, nil
	}
	// If sarts with "x", "0x", or "\\x" then interpret as hex
	// If sarts with "b", "0b", or "\\b" then interpret as binary
	// else interpret as decimal.
	// In all cases, ignore space/comma/underscores as formatting
	input = strings.ToLower(input)
	input = strings.TrimSpace(input)
	input = strings.Replace(input, " ", "", -1)
	input = strings.Replace(input, "_", "", -1)
	input = strings.Replace(input, ",", "", -1)
	sign := int64(1)
	if strings.HasPrefix(input, "-") {
		input = input[1:]
		sign = -1
	}
	parsed := int64(0)
	var err error
	if strings.HasPrefix(input, "x") || strings.HasPrefix(input, "0x") || strings.HasPrefix(input, "\\x") {
		// trim off leading "0x", "\\x", or "x"
		xIndex := strings.Index(input, "x")
		if xIndex != -1 {
			input = input[xIndex+1:]
		}
		parsed, err = strconv.ParseInt(input, 16, 64)
	} else if strings.HasPrefix(input, "b") || strings.HasPrefix(input, "0b") || strings.HasPrefix(input, "\\b") {
		// trim off leading "0b", "\\b", or "b"
		bIndex := strings.Index(input, "b")
		if bIndex != -1 {
			input = input[bIndex+1:]
		}
		parsed, err = strconv.ParseInt(input, 2, 64)
	} else {
		parsed, err = strconv.ParseInt(input, 10, 64)
	}
	return sign * parsed, err
}

// Returns parsed num, number of tokens consumed in parsing, error.
func parseNumWithPossibleUnaryOperators(tokens []tokenValue) (int64, int, error) {
	if len(tokens) == 0 {
		return 0, 0, errors.New("no tokens")
	}
	// NOTE: have to buffer unary operators as order matters.
	// ex: ~-1 == 0 which is not the same as -~ == 2
	unaries := make([]Token, 0)
	numTokensConsumed := 0
	for i := 0; i < len(tokens); i++ {
		if tokens[i].token == Minus || tokens[i].token == Plus || tokens[i].token == UnaryBitwiseNot {
			unaries = append(unaries, tokens[i].token)
			numTokensConsumed++
		} else {
			if tokens[i].token != Number {
				return 0, 0, fmt.Errorf("expected number, got token of type: %d", tokens[i].token)
			}
			break
		}
	}
	// Now on a token of type Number
	num := tokens[numTokensConsumed].value
	numTokensConsumed++
	// apply unary operators:
	for i := len(unaries) - 1; i >= 0; i-- {
		switch u := unaries[i]; u {
		case Plus:
			// do nothing, unary + keeps number the same
		case Minus:
			num *= -1
		case UnaryBitwiseNot:
			// NOTE: in golang, there is no "~", one uses the unary xor (^):
			num = ^num
		default:
			// hitting this implies programming error above:
			// not using same set of unary operators here versus above conditional
			return 0, 0, fmt.Errorf("unhandled unary token type: %d", u)
		}

	}
	return num, numTokensConsumed, nil
}

func EvalExpression(s string) (int64, error) {
	tokens, err := tokenize(s)
	if err != nil {
		return 0, err
	}
	ans, err := eval(tokens)
	if err != nil {
		return 0, err
	}
	return ans, nil
}

func eval(tokens []tokenValue) (int64, error) {
	if len(tokens) == 0 {
		return 0, errors.New("not enough tokens")
	}
	if len(tokens) == 1 { // trivial single number
		if tokens[0].token != Number {
			return 0, fmt.Errorf("expected number, got token of type: %d", tokens[0].token)
		}
		return tokens[0].value, nil
	}

	// First pass--solve parenthesis recursively, stack everything else for subsequent processing
	reduced := make([]tokenValue, 0, len(tokens))
	for {
		if len(tokens) < 1 {
			break
		}
		curToken := tokens[0]
		switch curToken.token {
		case OpenParentheses:
			// find matching ")" and then recursively solve that sub problem
			lastCloseParenIdx := -1
			depth := 0
			for i := 1; i < len(tokens); i++ {
				if tokens[i].token == CloseParentheses {
					if depth == 0 {
						lastCloseParenIdx = i
						break
					}
					depth--
				} else if tokens[i].token == OpenParentheses {
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
			reduced = append(reduced, tokenValue{token: Number, value: num})
		case CloseParentheses:
			// should always consume matching ")" when we first see the opening "("
			// getting here is a sign of mismatched parentheses
			return 0, errors.New("mismatched ')'")
		default:
			reduced = append(reduced, curToken)
			tokens = tokens[1:]
		}
	}

	// next pass: solve exponents
	stack1 := make([]tokenValue, 0, len(reduced))
	for i := 0; i < len(reduced); i++ {
		curToken := reduced[i]
		switch curToken.token {
		case Exponent:
			if i == len(reduced)-1 {
				return 0, errors.New("dangling '**'")
			}
			if len(stack1) == 0 {
				return 0, errors.New("mismatched '**'")
			}
			// pop exponent's base:
			baseTokenValue := stack1[len(stack1)-1]
			stack1 = stack1[:len(stack1)-1]
			// NOTE: not using parseNumWithPossibleUnaryOperators as exponent comes before unary minus
			if baseTokenValue.token != Number {
				return 0, fmt.Errorf("expected number as exponent base, got token type: %d", baseTokenValue.token)
			}
			// in other words -2^4 is -8 since it's really -(2^4) whereas (-2)^4 is 8...
			base := baseTokenValue.value
			// NOTE: handle unary minus etc on the exponentiator
			exponent, tokensConsumed, err := parseNumWithPossibleUnaryOperators(reduced[i+1:])
			if err != nil {
				return 0, err
			}
			// stack result:
			result := tokenValue{token: Number, value: int64(math.Pow(float64(base), float64(exponent)))}
			stack1 = append(stack1, result)
			i += tokensConsumed // skip over consumed exponent number
		default:
			stack1 = append(stack1, curToken)
		}
	}

	// next pass: solve multiply/divide/modulo
	if len(stack1) == 0 {
		return 0, errors.New("missing tokens")
	}
	stack2 := make([]tokenValue, 0, len(stack1))
	// Now that parentheses are dealt with, should begin with a number or number preceeded by unary operators
	var num int64
	var err error
	var tokensConsumed int
	num, tokensConsumed, err = parseNumWithPossibleUnaryOperators(stack1)
	stack1 = stack1[tokensConsumed:]
	if err != nil {
		return 0, err
	}
	stack2 = append(stack2, tokenValue{token: Number, value: num})
	for {
		if len(stack1) < 2 {
			break
		}
		// get operator and number that follows it:
		op := stack1[0]
		if op.token == Number || op.token == UnaryBitwiseNot {
			return 0, fmt.Errorf("expected operator, got: %v", op) // TOOD: better message here.
			// NOTE: can be an operator that overlaps with unary like minus or plus
		}
		stack1 = stack1[1:]
		num, tokensConsumed, err = parseNumWithPossibleUnaryOperators(stack1)
		if err != nil {
			return 0, err
		}
		stack1 = stack1[tokensConsumed:]
		switch op.token {
		case Multiply:
			if len(stack2) < 1 {
				return 0, fmt.Errorf("operator %v without preceeding number", op)
			}
			top := stack2[len(stack2)-1]
			stack2 = stack2[:len(stack2)-1]                                            // pop
			stack2 = append(stack2, tokenValue{token: Number, value: top.value * num}) // push
		case Divide:
			if len(stack2) < 1 {
				return 0, fmt.Errorf("operator %v without preceeding number", op)
			}
			top := stack2[len(stack2)-1]
			stack2 = stack2[:len(stack2)-1] // pop
			if num == 0 {
				return 0, errors.New("divide by zero")
			}
			stack2 = append(stack2, tokenValue{token: Number, value: top.value / num}) // push
		case Modulo:
			if len(stack2) < 1 {
				return 0, fmt.Errorf("operator %v without preceeding number", op)
			}
			top := stack2[len(stack2)-1]
			stack2 = stack2[:len(stack2)-1] // pop
			// NOTE: golang modulo works differently than python when negatives are involved.
			// See: https://stackoverflow.com/a/43018347
			// TLDR: when using a negative, ex: -9%10, python gives 1 whereas golang gives -9.
			// Keeping the golang behavior but worth pointing out.
			if num == 0 {
				return 0, errors.New("modulo divide by zero")
			}
			stack2 = append(stack2, tokenValue{token: Number, value: top.value % num}) // push
		case Number:
			return 0, errors.New("got number where expeted operator")
		default:
			// Operator to deal with in later pass (ex: +/- or <<, >>, etc)
			// stack num and op
			stack2 = append(stack2, op)
			stack2 = append(stack2, tokenValue{token: Number, value: num})
		}
	}

	// next pass: add/subtract
	stack3 := make([]tokenValue, 0, len(stack2))
	// Now that parentheses are dealt with, should begin with a number or number preceeded by unary operators
	num, tokensConsumed, err = parseNumWithPossibleUnaryOperators(stack2)
	stack2 = stack2[tokensConsumed:]
	if err != nil {
		return 0, err
	}
	stack3 = append(stack3, tokenValue{token: Number, value: num})
	for {
		if len(stack2) < 2 {
			break
		}
		// get operator and number that follows it:
		op := stack2[0]
		if op.token == Number || op.token == UnaryBitwiseNot {
			return 0, fmt.Errorf("expected operator, got: %v", op) // TOOD: better message here.
			// NOTE: can be an operator that overlaps with unary like minus or plus
		}
		stack2 = stack2[1:]
		num, tokensConsumed, err = parseNumWithPossibleUnaryOperators(stack2)
		if err != nil {
			return 0, err
		}
		stack2 = stack2[tokensConsumed:]
		switch op.token {
		case Minus:
			if len(stack3) < 1 {
				return 0, fmt.Errorf("operator %v without preceeding number", op)
			}
			top := stack3[len(stack3)-1]
			stack3 = stack3[:len(stack3)-1]                                            // pop
			stack3 = append(stack3, tokenValue{token: Number, value: top.value - num}) // push
		case Plus:
			if len(stack3) < 1 {
				return 0, fmt.Errorf("operator %v without preceeding number", op)
			}
			top := stack3[len(stack3)-1]
			stack3 = stack3[:len(stack3)-1]                                            // pop
			stack3 = append(stack3, tokenValue{token: Number, value: top.value + num}) // push
		case Number:
			return 0, errors.New("got number where expeted operator")
		default:
			// Operator to deal with in later pass (ex: <<, >>, etc)
			// stack num and op
			stack3 = append(stack3, op)
			stack3 = append(stack3, tokenValue{token: Number, value: num})
		}
	}

	// next pass: << and >>
	stack4 := make([]tokenValue, 0, len(stack3))
	// Now that parentheses are dealt with, should begin with a number or number preceeded by unary operators
	num, tokensConsumed, err = parseNumWithPossibleUnaryOperators(stack3)
	stack3 = stack3[tokensConsumed:]
	if err != nil {
		return 0, err
	}
	stack4 = append(stack4, tokenValue{token: Number, value: num})
	for {
		if len(stack3) < 2 {
			break
		}
		// get operator and number that follows it:
		op := stack3[0]
		if op.token == Number || op.token == UnaryBitwiseNot {
			return 0, fmt.Errorf("expected operator, got: %v", op) // TOOD: better message here.
			// NOTE: can be an operator that overlaps with unary like minus or plus
		}
		stack3 = stack3[1:]
		num, tokensConsumed, err = parseNumWithPossibleUnaryOperators(stack3)
		if err != nil {
			return 0, err
		}
		stack3 = stack3[tokensConsumed:]
		switch op.token {
		case LeftShift:
			if len(stack4) < 1 {
				return 0, fmt.Errorf("operator %v without preceeding number", op)
			}
			top := stack4[len(stack4)-1]
			stack4 = stack4[:len(stack4)-1] // pop
			if num < 0 {
				return 0, errors.New("negative bit shift")
			}
			stack4 = append(stack4, tokenValue{token: Number, value: top.value << num}) // push
		case RightShift:
			if len(stack4) < 1 {
				return 0, fmt.Errorf("operator %v without preceeding number", op)
			}
			top := stack4[len(stack4)-1]
			stack4 = stack4[:len(stack4)-1] // pop
			if num < 0 {
				return 0, errors.New("negative bit shift")
			}
			stack4 = append(stack4, tokenValue{token: Number, value: top.value >> num}) // push
		case Number:
			return 0, errors.New("got number where expeted operator")
		default:
			// Operator to deal with in later pass (ex: <<, >>, etc)
			// stack num and op
			stack4 = append(stack4, op)
			stack4 = append(stack4, tokenValue{token: Number, value: num})
		}
	}

	// TODO: binary operators--may need to be broken up into multiple passes...

	if len(stack3) != 0 {
		return 0, fmt.Errorf("dangling token(s): %v", stack3)
	}
	val := int64(0)
	for _, t := range stack4 {
		val += t.value
	}
	return val, nil
}

func tokenize(s string) ([]tokenValue, error) {
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
	numBuf := ""
	tokens := make([]tokenValue, 0)
	for i := 0; i < len(s); i++ {
		var curToken Token
		switch cur := s[i]; cur {
		case '(':
			curToken = OpenParentheses
		case ')':
			curToken = CloseParentheses
		case '+':
			curToken = Plus
		case '-':
			curToken = Minus
		case '~':
			curToken = UnaryBitwiseNot
		case '*':
			if i < len(s)-1 && s[i+1] == '*' {
				curToken = Exponent
				i++ // consume 2nd char
			} else {
				curToken = Multiply
			}
		case '/':
			curToken = Divide
		case '>':
			if i < len(s)-1 && s[i+1] == '>' {
				curToken = RightShift
				i++ // consume 2nd char
			} else {
				return nil, fmt.Errorf("Invalid operator: %q", s[i])
			}
		case '<':
			if i < len(s)-1 && s[i+1] == '<' {
				curToken = LeftShift
				i++ // consume 2nd char
			} else {
				return nil, fmt.Errorf("Invalid operator: %q", s[i])
			}
		case '%':
			curToken = Modulo
		default:
			curToken = Number
		}

		if curToken != Number {
			if len(numBuf) > 0 { // flush any buffered number
				val, err := ParseHexDecOrBin(numBuf)
				if err != nil {
					return nil, err
				}
				tokens = append(tokens, tokenValue{token: Number, value: val})
				numBuf = ""
			}

			// add current, non-number token
			tokens = append(tokens, tokenValue{token: curToken})
		} else {
			numBuf += string(s[i])
		}
	}

	if len(numBuf) > 0 { // flush any buffered number
		val, err := ParseHexDecOrBin(numBuf)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, tokenValue{token: Number, value: val})
		numBuf = ""
	}
	return tokens, nil
}
