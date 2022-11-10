package eval

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

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
	BitwiseAnd
	BitwiseOr
	BitwiseXor
)

// String provided for better error messages:
func (t *Token) String() string {
	if t == nil {
		return "nil"
	}
	switch *t {
	case OpenParentheses:
		return "("
	case CloseParentheses:
		return ")"
	case Number:
		return "number"
	case Plus:
		return "+"
	case Minus:
		return "-"
	case UnaryBitwiseNot:
		return "~"
	case Multiply:
		return "*"
	case Exponent:
		return "**"
	case Divide:
		return "/"
	case Modulo:
		return "%"
	case LeftShift:
		return "<<"
	case RightShift:
		return ">>"
	case BitwiseAnd:
		return "&"
	case BitwiseOr:
		return "|"
	case BitwiseXor:
		return "^"
	default:
		return "unknown"
	}
}

type tokenValue struct {
	token Token
	value int64
}

// String provided for better error messages:
func (tv *tokenValue) String() string {
	if tv == nil {
		return "nil"
	}
	if tv.token == Number {
		return strconv.Itoa(int(tv.value))
	}
	return tv.token.String()
}

type opFuncMap map[Token]func(int64, int64) (int64, error)

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
				return 0, 0, fmt.Errorf("expected number, got token: %s", tokens[i].token.String())
			}
			break
		}
	}
	if numTokensConsumed >= len(tokens) {
		return 0, 0, fmt.Errorf("dangling operator: %s", tokens[len(tokens)-1].String())
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
			return 0, 0, fmt.Errorf("unhandled unary token: %s", u.String())
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

func DisplayEvalResult(val int64) error {
	fmt.Fprintf(os.Stdout, "%d\n", val)
	if (val >= math.MinInt8 && val <= math.MaxInt8) || (val >= 0 && val < math.MaxUint8) {
		displayHexAndBinaryValue(val, 8)
	} else if (val >= math.MinInt16 && val <= math.MaxInt16) || (val >= 0 && val < math.MaxUint16) {
		displayHexAndBinaryValue(val, 16)
	} else if (val >= math.MinInt32 && val <= math.MaxInt32) || (val >= 0 && val < math.MaxUint32) {
		displayHexAndBinaryValue(val, 32)
	} else {
		displayHexAndBinaryValue(val, 64)
	}
	return nil // TODO: any things that can fail to cause error here?
}

func displayHexAndBinaryValue(val int64, bits int) {
	// fmt.Fprintf(os.Stdout, "\n%d Bits\n", bits)
	for i := (bits / 8) - 1; i >= 0; i-- {
		fmt.Fprintf(os.Stdout, "%02X ", val>>(8*i)&0xff)
	}
	fmt.Fprintf(os.Stdout, "\n")

	for i := (bits - 1); i >= 0; i-- {
		if val&(1<<i) != 0 {
			fmt.Fprintf(os.Stdout, "%d", 1)
		} else {
			fmt.Fprintf(os.Stdout, "%d", 0)
		}
		if i%4 == 0 {
			fmt.Fprintf(os.Stdout, " ")
		}
		if i%8 == 0 {
			fmt.Fprintf(os.Stdout, " ")
		}
	}
	fmt.Fprintf(os.Stdout, "\n")
}

func eval(tokens []tokenValue) (int64, error) {
	if len(tokens) == 0 {
		return 0, errors.New("not enough tokens")
	}
	if len(tokens) == 1 { // trivial single number
		if tokens[0].token != Number {
			return 0, fmt.Errorf("expected number, got token: %s", tokens[0].token.String())
		}
		return tokens[0].value, nil
	}

	opFuncs := opFuncMap{
		// NOTE: absent are OpenParentheses, CloseParentheses, and UnaryBitwiseNot
		// as the first two are not operators and the latter is a unary-only which
		// will be handled in parseNumWithPossibleUnaryOperators().
		Exponent: func(left, right int64) (int64, error) {
			return int64(math.Pow(float64(left), float64(right))), nil
		},
		Multiply: func(left, right int64) (int64, error) {
			return left * right, nil
		},
		Divide: func(left, right int64) (int64, error) {
			if right == 0 {
				return 0, errors.New("divide by zero")
			}
			return left / right, nil
		},
		Modulo: func(left, right int64) (int64, error) {
			if right == 0 {
				return 0, errors.New("modulo divide by zero")
			}
			return left % right, nil
		},
		Plus: func(left, right int64) (int64, error) {
			return left + right, nil
		},
		Minus: func(left, right int64) (int64, error) {
			return left - right, nil
		},
		RightShift: func(left, right int64) (int64, error) {
			if right < 0 {
				return 0, errors.New("negative bit shift")
			}
			return left >> right, nil
		},
		LeftShift: func(left, right int64) (int64, error) {
			if right < 0 {
				return 0, errors.New("negative bit shift")
			}
			return left << right, nil
		},
		BitwiseAnd: func(left, right int64) (int64, error) {
			return left & right, nil
		},
		BitwiseXor: func(left, right int64) (int64, error) {
			return left ^ right, nil
		},
		BitwiseOr: func(left, right int64) (int64, error) {
			return left | right, nil
		},
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
	// NOTE: since expontent is stronger than unary operators, this is handled first, and
	// differently than subsequent passes through tokens.
	stack := make([]tokenValue, 0, len(reduced))
	exponentFunc, ok := opFuncs[Exponent]
	if !ok {
		return 0, errors.New("no function for exponent operator")
	}
	for i := 0; i < len(reduced); i++ {
		curToken := reduced[i]
		switch curToken.token {
		case Exponent:
			if i == len(reduced)-1 {
				return 0, errors.New("dangling '**'")
			}
			if len(stack) == 0 {
				return 0, errors.New("mismatched '**'")
			}
			// pop exponent's base:
			baseTokenValue := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			// NOTE: not using parseNumWithPossibleUnaryOperators as exponent comes before unary minus
			if baseTokenValue.token != Number {
				return 0, fmt.Errorf("expected number as exponent base, got token: %s", baseTokenValue.token.String())
			}
			// in other words -2^4 is -8 since it's really -(2^4) whereas (-2)^4 is 8...
			base := baseTokenValue.value
			// NOTE: handle unary minus etc on the exponentiator
			exponent, tokensConsumed, err := parseNumWithPossibleUnaryOperators(reduced[i+1:])
			if err != nil {
				return 0, err
			}
			// stack result:
			if result, err := exponentFunc(base, exponent); err == nil {
				stack = append(stack, tokenValue{token: Number, value: result})
				i += tokensConsumed // skip over consumed exponent number
			} else {
				return 0, err
			}
		default:
			stack = append(stack, curToken)
		}
	}

	// Apply operators in order of precedence,
	// doing a pass for each set of left-to-right/equal precedence ops.
	var err error
	if stack, err = applyOperators(map[Token]bool{Multiply: true, Divide: true, Modulo: true}, stack, opFuncs); err != nil {
		return 0, err
	}
	if stack, err = applyOperators(map[Token]bool{Plus: true, Minus: true}, stack, opFuncs); err != nil {
		return 0, err
	}
	if stack, err = applyOperators(map[Token]bool{LeftShift: true, RightShift: true}, stack, opFuncs); err != nil {
		return 0, err
	}
	if stack, err = applyOperators(map[Token]bool{BitwiseAnd: true}, stack, opFuncs); err != nil {
		return 0, err
	}
	if stack, err = applyOperators(map[Token]bool{BitwiseXor: true}, stack, opFuncs); err != nil {
		return 0, err
	}
	if stack, err = applyOperators(map[Token]bool{BitwiseOr: true}, stack, opFuncs); err != nil {
		return 0, err
	}

	// Should now have a stack of only numbers, add them up to get answer:
	val := int64(0)
	for _, t := range stack {
		if t.token != Number {
			return 0, fmt.Errorf("expected only numbers, got token: %s", t.String())
		}
		val += t.value
	}
	return val, nil
}

func applyOperators(opSet map[Token]bool, stack []tokenValue, opFuncs opFuncMap) ([]tokenValue, error) {
	updatedStack := make([]tokenValue, 0, len(stack))
	if len(stack) == 0 {
		return updatedStack, errors.New("missing tokens")
	}
	// Now that parentheses are dealt with, should begin with a number or number preceeded by unary operators
	var num int64
	var err error
	var tokensConsumed int
	num, tokensConsumed, err = parseNumWithPossibleUnaryOperators(stack)
	stack = stack[tokensConsumed:]
	if err != nil {
		return updatedStack, err
	}
	updatedStack = append(updatedStack, tokenValue{token: Number, value: num})
	for {
		if len(stack) < 2 {
			break
		}
		// get operator and number that follows it:
		op := stack[0]
		if op.token == Number || op.token == UnaryBitwiseNot {
			return updatedStack, fmt.Errorf("expected operator, got: %s", op.String())
			// NOTE: can be an operator that overlaps with unary like minus or plus, but can't be a unary-only one.
		}
		stack = stack[1:]
		num, tokensConsumed, err = parseNumWithPossibleUnaryOperators(stack)
		if err != nil {
			return updatedStack, err
		}
		stack = stack[tokensConsumed:]

		if op.token == Number {
			return updatedStack, errors.New("got number where expeted operator")
		}
		if _, ok := opSet[op.token]; ok {
			// have an operator we're tackling in this pass--apply it.
			opFunc, ok := opFuncs[op.token]
			if !ok {
				return updatedStack, fmt.Errorf("no function for operator: %s", op.token.String())
			}
			top := updatedStack[len(updatedStack)-1]
			updatedStack = updatedStack[:len(updatedStack)-1] // pop
			if result, err := opFunc(top.value, num); err == nil {
				updatedStack = append(updatedStack, tokenValue{token: Number, value: result}) // push
			} else {
				return updatedStack, err
			}
		} else {
			// Operator to deal with in later pass...
			// stack num and op
			updatedStack = append(updatedStack, op)
			updatedStack = append(updatedStack, tokenValue{token: Number, value: num})
		}
	}
	// Should have consumed all of the original stack:
	if len(stack) != 0 {
		return updatedStack, fmt.Errorf("dangling token: %s", stack[0].String())
	}
	return updatedStack, nil
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
		case '&':
			curToken = BitwiseAnd
		case '|':
			curToken = BitwiseOr
		case '^':
			curToken = BitwiseXor
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
