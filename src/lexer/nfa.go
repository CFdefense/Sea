package lexer

import (
	"fmt"
	"regexp"
	"strings"

	debugger "github.com/CFdefense/compiler/src/debug"
)

// NFA state structure
type NFAState struct {
	id                 int
	isAccepting        bool
	tokenType          TokenType
	transitions        map[string][]*NFAState
	epsilonTransitions []*NFAState
}

// NFA structure
type NFA struct {
	start *NFAState
	end   *NFAState
}

var stateIDCounter int = 0

func generateStateID() int {
	stateIDCounter++
	return stateIDCounter
}

// RegexToken represents a single regex atom
type RegexToken struct {
	Value string
	Type  string
}

// tokenizeRegex splits a regex string into regex atoms
func tokenizeRegex(pattern string) []RegexToken {
	var tokens []RegexToken
	i := 0
	for i < len(pattern) {
		switch pattern[i] {
		case '\\':
			// Handle escape sequence - keep as single token
			if i+1 < len(pattern) {
				tokens = append(tokens, RegexToken{pattern[i : i+2], "escape"})
				i += 2
			} else {
				tokens = append(tokens, RegexToken{string(pattern[i]), "literal"})
				i++
			}
		case '[':
			// Handle character class - keep entire class as single token
			j := i + 1
			for j < len(pattern) && pattern[j] != ']' {
				if pattern[j] == '\\' && j+1 < len(pattern) {
					j += 2
				} else {
					j++
				}
			}
			if j < len(pattern) {
				tokens = append(tokens, RegexToken{pattern[i : j+1], "class"})
				i = j + 1
			} else {
				tokens = append(tokens, RegexToken{string(pattern[i]), "literal"})
				i++
			}
		case '(':
			// Handle grouping - keep parentheses as operators
			tokens = append(tokens, RegexToken{string(pattern[i]), "operator"})
			i++
		case ')':
			// Handle grouping - keep parentheses as operators
			tokens = append(tokens, RegexToken{string(pattern[i]), "operator"})
			i++
		case '*', '+', '?':
			// Handle quantifiers - keep as operators
			tokens = append(tokens, RegexToken{string(pattern[i]), "operator"})
			i++
		case '|':
			// Handle alternation - keep as operator
			tokens = append(tokens, RegexToken{string(pattern[i]), "operator"})
			i++
		case '.':
			// Handle dot - keep as operator
			tokens = append(tokens, RegexToken{string(pattern[i]), "operator"})
			i++
		case '^', '$':
			// Handle anchors - keep as operators
			tokens = append(tokens, RegexToken{string(pattern[i]), "operator"})
			i++
		default:
			// Handle literals - collect consecutive letters/digits as single token
			if isLetterOrDigit(pattern[i]) {
				start := i
				for i < len(pattern) && isLetterOrDigit(pattern[i]) {
					i++
				}
				tokens = append(tokens, RegexToken{pattern[start:i], "literal"})
			} else {
				// Single character literal
				tokens = append(tokens, RegexToken{string(pattern[i]), "literal"})
				i++
			}
		}
	}
	return tokens
}

// Helper function to check if character is letter or digit
func isLetterOrDigit(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

// postfix converts a tokenized regex to postfix (RPN) using the shunting yard algorithm
func postfix(regex string, tokenName string, debug *debugger.Debug) string {
	tokens := tokenizeRegex(regex)
	// Operator precedence
	precedence := map[string]int{
		"*": 3,
		"+": 3,
		"?": 3,
		"·": 2, // explicit concatenation
		"|": 1,
	}
	isOperator := func(t RegexToken) bool {
		return t.Type == "operator" && (t.Value == "*" || t.Value == "+" || t.Value == "?" || t.Value == "·" || t.Value == "|")
	}
	// Step 1: Insert explicit concatenation operators
	var explicit []RegexToken
	for i := 0; i < len(tokens); i++ {
		t := tokens[i]
		explicit = append(explicit, t)
		if i+1 < len(tokens) {
			next := tokens[i+1]
			// Insert concatenation if:
			// (literal/class/escape/close paren/quantifier) followed by (literal/class/escape/open paren)
			if (t.Type == "literal" || t.Type == "class" || t.Type == "escape" ||
				(t.Type == "operator" && (t.Value == ")" || t.Value == "*" || t.Value == "+" || t.Value == "?"))) &&
				(next.Type == "literal" || next.Type == "class" || next.Type == "escape" ||
					(next.Type == "operator" && next.Value == "(")) {
				explicit = append(explicit, RegexToken{"·", "operator"})
			}
		}
	}
	// Step 2: Shunting Yard
	var output []string
	var stack []RegexToken
	for _, t := range explicit {
		switch {
		case t.Type == "operator" && t.Value == "(":
			stack = append(stack, t)
		case t.Type == "operator" && t.Value == ")":
			for len(stack) > 0 && !(stack[len(stack)-1].Type == "operator" && stack[len(stack)-1].Value == "(") {
				output = append(output, stack[len(stack)-1].Value)
				stack = stack[:len(stack)-1]
			}
			if len(stack) > 0 {
				stack = stack[:len(stack)-1] // pop '('
			}
		case isOperator(t):
			for len(stack) > 0 && isOperator(stack[len(stack)-1]) && precedence[stack[len(stack)-1].Value] >= precedence[t.Value] {
				output = append(output, stack[len(stack)-1].Value)
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, t)
		default:
			output = append(output, t.Value)
		}
	}
	for len(stack) > 0 {
		output = append(output, stack[len(stack)-1].Value)
		stack = stack[:len(stack)-1]
	}
	if debug != nil {
		debug.DebugLog(fmt.Sprintf("[%s] Regex: %s", tokenName, regex), false)
		debug.DebugLog(fmt.Sprintf("[%s] Postfix: %s", tokenName, strings.Join(output, " ")), false)
	}
	return strings.Join(output, " ")
}

// thompson's construction
// TODO: Implement true Thompson construction with regex parsing
// For now, create placeholder NFAs that will be replaced with proper construction
// TODO: Parse regex to postfix notation and build proper NFA
// For now, create a placeholder NFA structure
func thompsonConstruct(regexPattern string, tokenType TokenType) *NFA {
	single_literal := regexp.MustCompile(`^(\\.|[^\\.*+?|()[\]{}^$])$`)
	// concatenation :=
	// alternation :=
	// kleen_star :=
	if single_literal.MatchString(regexPattern) {
		s1 := &NFAState{
			id:                 generateStateID(),
			isAccepting:        false,
			tokenType:          tokenType,
			transitions:        make(map[string][]*NFAState),
			epsilonTransitions: []*NFAState{},
		}
		s2 := &NFAState{
			id:                 generateStateID(),
			isAccepting:        true,
			tokenType:          tokenType,
			transitions:        make(map[string][]*NFAState),
			epsilonTransitions: []*NFAState{},
		}
		add_transition(s1, regexPattern, s2)

		return &NFA{s1, s2}

	} else {
		return &NFA{}
	}
}

func add_transition(from_state *NFAState, input string, to_state *NFAState) {
	from_state.transitions[input] = append(from_state.transitions[input], to_state)
}
