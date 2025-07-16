package lexer

import (
	"fmt"
	"strings"

	debugger "github.com/CFdefense/compiler/src/debug"
)

// NFA state structure
type NFAState struct {
	id          int
	isAccepting bool
	tokenType   TokenType
	transitions map[string][]*NFAState
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
		// Handle escape sequences: any backslash + character is a single escape token
		if pattern[i] == '\\' && i+1 < len(pattern) {
			// Regular escape: backslash + any character
			tokens = append(tokens, RegexToken{pattern[i : i+2], "escape"})
			i += 2
			continue
		}

		switch pattern[i] {
		case '[':
			// Handle character class
			j := i + 1
			classContent := ""
			for j < len(pattern) && pattern[j] != ']' {
				if pattern[j] == '\\' && j+1 < len(pattern) {
					// Always treat backslash + any character as a single unit inside class
					classContent += pattern[j : j+2]
					j += 2
				} else {
					classContent += string(pattern[j])
					j++
				}
			}
			if j < len(pattern) {
				tok := RegexToken{"[" + classContent + "]", "class"}
				tokens = append(tokens, tok)
				i = j + 1
			} else {
				tok := RegexToken{string(pattern[i]), "literal"}
				tokens = append(tokens, tok)
				i++
			}
		case '(', ')', '*', '+', '?', '|', '.':
			tok := RegexToken{string(pattern[i]), "operator"}
			tokens = append(tokens, tok)
			i++
		case '^', '$':
			tok := RegexToken{string(pattern[i]), "anchor"}
			tokens = append(tokens, tok)
			i++
		default:
			if isLetterOrDigit(pattern[i]) {
				start := i
				for i < len(pattern) && isLetterOrDigit(pattern[i]) {
					i++
				}
				tok := RegexToken{pattern[start:i], "literal"}
				tokens = append(tokens, tok)
			} else {
				tok := RegexToken{string(pattern[i]), "literal"}
				tokens = append(tokens, tok)
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

// check if a character is a word character
func isWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

// check if at word position i
func isAtWordBoundary(input string, i int) bool {
	// a word boundary occurs when:
	// string starts with word char
	// string ends with word char
	// current and previous chars are not word chars
	// current char is not word char and previous is word char

	if i == 0 {
		// At start
		return len(input) > 0 && isWordChar(input[0])
	}

	if i >= len(input) {
		// At end
		return len(input) > 0 && isWordChar(input[len(input)-1])
	}

	// In middle
	prevIsWord := isWordChar(input[i-1])
	currIsWord := isWordChar(input[i])

	return prevIsWord != currIsWord
}

// postfix converts a tokenized regex to postfix (RPN) using the shunting yard algorithm
func postfix(regex string, tokenName string, debug *debugger.Debug) string {
	if debug != nil {
		debug.DebugLog(fmt.Sprintf("[%s] RAW PATTERN: %q", tokenName, regex), false)
	}
	tokens := tokenizeRegex(regex)

	// Debug: print tokens
	if debug != nil {
		tokenStrs := make([]string, len(tokens))
		for i, t := range tokens {
			tokenStrs[i] = fmt.Sprintf("%s(%s)", t.Value, t.Type)
		}
		debug.DebugLog(fmt.Sprintf("[%s] Tokens: %s", tokenName, strings.Join(tokenStrs, " ")), false)
	}

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
		// Treat anchors, word boundaries, and character classes as literals for concatenation
		if t.Type == "anchor" || t.Type == "word_boundary" || t.Type == "class" {
			t.Type = "literal"
		}
		explicit = append(explicit, t)
		if i+1 < len(tokens) {
			next := tokens[i+1]
			// Treat anchors, word boundaries, and character classes as literals for next token too
			if next.Type == "anchor" || next.Type == "word_boundary" || next.Type == "class" {
				next.Type = "literal"
			}
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

	// Debug: print explicit tokens
	if debug != nil {
		explicitStrs := make([]string, len(explicit))
		for i, t := range explicit {
			explicitStrs[i] = fmt.Sprintf("%s(%s)", t.Value, t.Type)
		}
		debug.DebugLog(fmt.Sprintf("[%s] Explicit: %s", tokenName, strings.Join(explicitStrs, " ")), false)
		debug.DebugLog(fmt.Sprintf("[%s] Token count: %d", tokenName, len(explicit)), false)
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
func thompsonConstruct(postfix string, tokenType TokenType) *NFA {
	tokens := strings.Fields(postfix)
	type nfaStackElem struct {
		nfa *NFA
	}
	var stack []nfaStackElem

	for _, tok := range tokens {
		switch tok {
		case "·": // Concatenation
			if len(stack) < 2 {
				panic("thompsonConstruct: not enough operands for concatenation")
			}
			b := stack[len(stack)-1].nfa
			a := stack[len(stack)-2].nfa
			stack = stack[:len(stack)-2]
			// Connect a.end to b.start with epsilon
			add_transition(a.end, "ε", b.start)
			stack = append(stack, nfaStackElem{&NFA{a.start, b.end}})
		case "|": // Alternation
			if len(stack) < 2 {
				panic("thompsonConstruct: not enough operands for alternation")
			}
			b := stack[len(stack)-1].nfa
			a := stack[len(stack)-2].nfa
			stack = stack[:len(stack)-2]
			start := &NFAState{id: generateStateID(), transitions: make(map[string][]*NFAState)}
			end := &NFAState{id: generateStateID(), isAccepting: true, tokenType: tokenType, transitions: make(map[string][]*NFAState)}
			add_transition(start, "ε", a.start)
			add_transition(start, "ε", b.start)
			add_transition(a.end, "ε", end)
			add_transition(b.end, "ε", end)
			stack = append(stack, nfaStackElem{&NFA{start, end}})
		case "*": // Kleene star
			if len(stack) < 1 {
				panic("thompsonConstruct: not enough operands for kleene star")
			}
			a := stack[len(stack)-1].nfa
			stack = stack[:len(stack)-1]
			start := &NFAState{id: generateStateID(), transitions: make(map[string][]*NFAState)}
			end := &NFAState{id: generateStateID(), isAccepting: true, tokenType: tokenType, transitions: make(map[string][]*NFAState)}
			add_transition(start, "ε", a.start)
			add_transition(start, "ε", end)
			add_transition(a.end, "ε", a.start)
			add_transition(a.end, "ε", end)
			stack = append(stack, nfaStackElem{&NFA{start, end}})
		case "+": // One or more
			if len(stack) < 1 {
				panic("thompsonConstruct: not enough operands for plus")
			}
			a := stack[len(stack)-1].nfa
			stack = stack[:len(stack)-1]
			start := &NFAState{id: generateStateID(), transitions: make(map[string][]*NFAState)}
			end := &NFAState{id: generateStateID(), isAccepting: true, tokenType: tokenType, transitions: make(map[string][]*NFAState)}
			add_transition(start, "ε", a.start)
			add_transition(a.end, "ε", a.start)
			add_transition(a.end, "ε", end)
			stack = append(stack, nfaStackElem{&NFA{start, end}})
		case "?": // Zero or one
			if len(stack) < 1 {
				panic("thompsonConstruct: not enough operands for question")
			}
			a := stack[len(stack)-1].nfa
			stack = stack[:len(stack)-1]
			start := &NFAState{id: generateStateID(), transitions: make(map[string][]*NFAState)}
			end := &NFAState{id: generateStateID(), isAccepting: true, tokenType: tokenType, transitions: make(map[string][]*NFAState)}
			add_transition(start, "ε", a.start)
			add_transition(start, "ε", end)
			add_transition(a.end, "ε", end)
			stack = append(stack, nfaStackElem{&NFA{start, end}})
		default:
			// Handle different token types
			start := &NFAState{id: generateStateID(), transitions: make(map[string][]*NFAState)}
			end := &NFAState{id: generateStateID(), isAccepting: true, tokenType: tokenType, transitions: make(map[string][]*NFAState)}

			if strings.HasPrefix(tok, "[") && strings.HasSuffix(tok, "]") {
				// Character class - parse and create transitions for each character
				classContent := tok[1 : len(tok)-1] // Remove brackets
				if strings.HasPrefix(classContent, "^") {
					// Negated character class
					// Parse the characters in the class (excluding the ^)
					chars := parseCharacterClass(classContent[1:])
					// Create transitions for all characters NOT in the class
					// For simplicity, we'll create transitions for common ASCII characters
					for i := 0; i < 128; i++ {
						char := string(byte(i))
						found := false
						for _, classChar := range chars {
							if char == classChar {
								found = true
								break
							}
						}
						if !found {
							add_transition(start, char, end)
						}
					}
				} else {
					// Regular character class
					// Parse character ranges and individual characters
					chars := parseCharacterClass(classContent)
					for _, char := range chars {
						add_transition(start, char, end)
					}
				}
			} else if strings.HasPrefix(tok, "\\") {
				// Escape sequence
				if len(tok) == 2 {
					// Single character escape
					escapedChar := tok[1]
					switch escapedChar {
					case 'n':
						add_transition(start, "\n", end)
					case 't':
						add_transition(start, "\t", end)
					case 'r':
						add_transition(start, "\r", end)
					case '\\':
						add_transition(start, "\\", end)
					case '"':
						add_transition(start, "\"", end)
					case '\'':
						add_transition(start, "'", end)
					case 'b':
						// Word boundary position assertion
						add_transition(start, "word_boundary", end)
					case 'd':
						// Digit class
						for i := '0'; i <= '9'; i++ {
							add_transition(start, string(i), end)
						}
					case 's':
						// Whitespace class
						add_transition(start, " ", end)
						add_transition(start, "\t", end)
						add_transition(start, "\n", end)
						add_transition(start, "\r", end)
					case 'w':
						// Word character class
						for i := 'a'; i <= 'z'; i++ {
							add_transition(start, string(i), end)
						}
						for i := 'A'; i <= 'Z'; i++ {
							add_transition(start, string(i), end)
						}
						for i := '0'; i <= '9'; i++ {
							add_transition(start, string(i), end)
						}
						add_transition(start, "_", end)
					default:
						// Any other escaped character
						add_transition(start, string(escapedChar), end)
					}
				}
			} else if tok == "^" || tok == "$" {
				// Anchors - these are position assertions, not character consumers
				// For now, treat as special tokens that need special handling in simulation
				if tok == "^" {
					add_transition(start, "start_anchor", end)
				} else {
					add_transition(start, "end_anchor", end)
				}
			} else if tok == "\\b" {
				// Word boundary - position assertion, not character consumer
				add_transition(start, "word_boundary", end)
			} else if tok == "." {
				// Dot operator - matches any character
				add_transition(start, "any", end)
			} else {
				// Regular literal
				add_transition(start, tok, end)
			}

			stack = append(stack, nfaStackElem{&NFA{start, end}})
		}
	}
	if len(stack) != 1 {
		panic("thompsonConstruct: stack did not end with exactly one NFA")
	}
	return stack[0].nfa
}

// parseCharacterClass parses a character class and returns all characters it matches
func parseCharacterClass(classContent string) []string {
	var chars []string
	i := 0
	for i < len(classContent) {
		// Handle escape sequences
		if classContent[i] == '\\' && i+1 < len(classContent) {
			escapedChar := classContent[i+1]
			chars = append(chars, string(escapedChar))
			i += 2
			continue
		}

		// Handle character ranges
		if i+2 < len(classContent) && classContent[i+1] == '-' {
			start := classContent[i]
			end := classContent[i+2]

			// Validate range (start <= end)
			if start > end {
				// Invalid range, treat as literal characters
				chars = append(chars, string(start))
				chars = append(chars, "-")
				chars = append(chars, string(end))
			} else {
				// Valid range
				for c := start; c <= end; c++ {
					chars = append(chars, string(c))
				}
			}
			i += 3
		} else {
			// Single character
			chars = append(chars, string(classContent[i]))
			i++
		}
	}
	return chars
}

// add_transition adds a defined transition to the NFA
func add_transition(from_state *NFAState, input string, to_state *NFAState) {
	from_state.transitions[input] = append(from_state.transitions[input], to_state)
}

// print the NFA structure
func (nfa *NFA) Print(debug *debugger.Debug) {
	visited := make(map[int]bool)
	var queue []*NFAState
	queue = append(queue, nfa.start)
	debug.DebugLog("NFA Structure:", false)
	for len(queue) > 0 {
		state := queue[0]
		queue = queue[1:]
		if visited[state.id] {
			continue
		}
		visited[state.id] = true
		debug.DebugLog(fmt.Sprintf("State %d", state.id), false)
		if state.isAccepting {
			debug.DebugLog(fmt.Sprintf("[accepting, type=%s]", state.tokenType.String()), false)
		}
		debug.DebugLog("", false)
		for symbol, targets := range state.transitions {
			for _, target := range targets {
				debug.DebugLog(fmt.Sprintf("  %d --%s--> %d", state.id, symbol, target.id), false)
				queue = append(queue, target)
			}
		}
	}
}

// TestNFA validates an NFA by testing it with various inputs
func (nfa *NFA) TestNFA(testCases []string, expectedResults []bool, debug *debugger.Debug) bool {
	if len(testCases) != len(expectedResults) {
		debug.DebugLog("TestNFA: test cases and expected results must have same length", true)
		return false
	}

	allPassed := true
	for i, testInput := range testCases {
		result := nfa.Simulate(testInput)
		expected := expectedResults[i]
		if result != expected {
			debug.DebugLog(fmt.Sprintf("TestNFA FAILED: input '%s', expected %v, got %v", testInput, expected, result), true)
			allPassed = false
		} else {
			debug.DebugLog(fmt.Sprintf("TestNFA PASSED: input '%s' -> %v", testInput, result), false)
		}
	}
	return allPassed
}

// Simulate runs an NFA on a given input string
func (nfa *NFA) Simulate(input string) bool {
	// Start with epsilon closure of start state
	currentStates := nfa.epsilonClosure([]*NFAState{nfa.start})

	// Debug: show initial states
	fmt.Printf("Starting simulation with input: %q\n", input)
	fmt.Printf("Initial states: %v\n", getStateIDs(currentStates))

	// Handle start anchor (^), check at position 0
	nextStates := []*NFAState{}
	for _, state := range currentStates {
		if transitions, exists := state.transitions["start_anchor"]; exists {
			fmt.Printf("  State %d has start_anchor transition to states: %v\n",
				state.id, getStateIDs(transitions))
			nextStates = append(nextStates, transitions...)
		}
	}
	if len(nextStates) > 0 {
		currentStates = nfa.epsilonClosure(nextStates)
		fmt.Printf("After start anchor and epsilon closure: %v\n", getStateIDs(currentStates))
	}

	// Check for word boundary at start of input
	if isAtWordBoundary(input, 0) {
		nextStates = []*NFAState{}
		for _, state := range currentStates {
			if transitions, exists := state.transitions["word_boundary"]; exists {
				fmt.Printf("  State %d has word_boundary transition at start to states: %v\n",
					state.id, getStateIDs(transitions))
				nextStates = append(nextStates, transitions...)
			}
		}
		if len(nextStates) > 0 {
			currentStates = nfa.epsilonClosure(nextStates)
			fmt.Printf("After start word boundary and epsilon closure: %v\n", getStateIDs(currentStates))
		}
	}

	for i := 0; i < len(input); i++ {
		char := string(input[i])
		nextStates = []*NFAState{}

		fmt.Printf("Processing character '%s' at position %d\n", char, i)

		// Check for word boundary at current position before consuming character
		if isAtWordBoundary(input, i) {
			for _, state := range currentStates {
				if transitions, exists := state.transitions["word_boundary"]; exists {
					fmt.Printf("  State %d has word_boundary transition at position %d to states: %v\n",
						state.id, i, getStateIDs(transitions))
					nextStates = append(nextStates, transitions...)
				}
			}
		}

		// For each current state, find all states reachable on this character
		for _, state := range currentStates {
			if transitions, exists := state.transitions[char]; exists {
				fmt.Printf("  State %d has transition on '%s' to states: %v\n",
					state.id, char, getStateIDs(transitions))
				nextStates = append(nextStates, transitions...)
			}
			// Also check for "any" transitions (dot operator)
			if transitions, exists := state.transitions["any"]; exists {
				fmt.Printf("  State %d has 'any' transition to states: %v\n",
					state.id, getStateIDs(transitions))
				nextStates = append(nextStates, transitions...)
			}
		}

		// Take epsilon closure of next states
		currentStates = nfa.epsilonClosure(nextStates)
		fmt.Printf("After epsilon closure: %v\n", getStateIDs(currentStates))

		// If no states, reject
		if len(currentStates) == 0 {
			fmt.Printf("No states remaining - REJECTING\n")
			return false
		}
	}

	// Check for word boundary at end of input
	if isAtWordBoundary(input, len(input)) {
		nextStates = []*NFAState{}
		for _, state := range currentStates {
			if transitions, exists := state.transitions["word_boundary"]; exists {
				fmt.Printf("  State %d has word_boundary transition at end to states: %v\n",
					state.id, getStateIDs(transitions))
				nextStates = append(nextStates, transitions...)
			}
		}
		if len(nextStates) > 0 {
			currentStates = nfa.epsilonClosure(nextStates)
			fmt.Printf("After end word boundary and epsilon closure: %v\n", getStateIDs(currentStates))
		}
	}

	// Handle end anchor ($)
	nextStates = []*NFAState{} // Reset nextStates
	for _, state := range currentStates {
		if transitions, exists := state.transitions["end_anchor"]; exists {
			fmt.Printf("  State %d has end_anchor transition to states: %v\n",
				state.id, getStateIDs(transitions))
			nextStates = append(nextStates, transitions...)
		}
	}
	if len(nextStates) > 0 {
		currentStates = nfa.epsilonClosure(nextStates)
		fmt.Printf("After end anchor and epsilon closure: %v\n", getStateIDs(currentStates))
	}

	// Check if any accepting state is in current states
	for _, state := range currentStates {
		if state.isAccepting {
			fmt.Printf("Found accepting state %d (type: %s) - ACCEPTING\n",
				state.id, state.tokenType.String())
			return true
		}
	}

	fmt.Printf("No accepting states found - REJECTING\n")
	return false
}

// Helper function to get state IDs for debugging
func getStateIDs(states []*NFAState) []int {
	ids := make([]int, len(states))
	for i, state := range states {
		ids[i] = state.id
	}
	return ids
}

// epsilonClosure computes the epsilon closure of a set of states
func (nfa *NFA) epsilonClosure(states []*NFAState) []*NFAState {
	closure := make(map[int]*NFAState)
	var stack []*NFAState

	// Initialize with input states
	for _, state := range states {
		closure[state.id] = state
		stack = append(stack, state)
	}

	// Process stack until empty
	for len(stack) > 0 {
		state := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		// Add all states reachable via epsilon transitions
		if transitions, exists := state.transitions["ε"]; exists {
			for _, nextState := range transitions {
				if _, visited := closure[nextState.id]; !visited {
					closure[nextState.id] = nextState
					stack = append(stack, nextState)
				}
			}
		}
	}

	// Convert map back to slice
	result := make([]*NFAState, 0, len(closure))
	for _, state := range closure {
		result = append(result, state)
	}
	return result
}
