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

var generateStateID = func() func() int {
	var counter int
	return func() int {
		counter++
		return counter
	}
}()

// RegexToken represents a single regex atom
type RegexToken struct {
	Value string
	Type  string
}

// Exported helper functions for testing

// ConvertToPostfix converts a regex pattern to postfix notation for testing
func ConvertToPostfix(pattern string) string {
	return postfix(pattern, "TEST", nil)
}

// CreateNFA creates an NFA from a postfix regex pattern for testing
func CreateNFA(postfixPattern string, tokenType TokenType) *NFA {
	return thompsonConstruct(postfixPattern, tokenType)
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
			var classContent strings.Builder
			for j < len(pattern) && pattern[j] != ']' {
				if pattern[j] == '\\' && j+1 < len(pattern) {
					// Always treat backslash + any character as a single unit inside class
					classContent.WriteString(pattern[j : j+2])
					j += 2
				} else {
					classContent.WriteByte(pattern[j])
					j++
				}
			}
			if j < len(pattern) {
				tok := RegexToken{"[" + classContent.String() + "]", "class"}
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
		case '{':
			// Handle range quantifiers {n}, {n,}, {n,m}
			j := i + 1
			var quantifier strings.Builder
			quantifier.WriteByte('{')
			for j < len(pattern) && pattern[j] != '}' {
				quantifier.WriteByte(pattern[j])
				j++
			}
			if j < len(pattern) {
				quantifier.WriteByte('}')
				tokens = append(tokens, RegexToken{quantifier.String(), "operator"})
				i = j + 1
			} else {
				// unmatched brace implies literal
				tokens = append(tokens, RegexToken{string(pattern[i]), "literal"})
				i++
			}
		case '^', '$':
			tok := RegexToken{string(pattern[i]), "anchor"}
			tokens = append(tokens, tok)
			i++
		default:
			if isWordChar(pattern[i]) {
				start := i
				for i < len(pattern) && isWordChar(pattern[i]) {
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

// simplifyRegex preprocesses complex regex patterns to make them easier to handle
func simplifyRegex(pattern string) string {
	// Handle common problematic patterns
	simplified := pattern

	// Replace \b (word boundary) with a simpler pattern at the end
	if strings.HasSuffix(simplified, "\\b") {
		simplified = strings.TrimSuffix(simplified, "\\b")
	}

	// Replace anchors at the beginning for simpler processing
	if strings.HasPrefix(simplified, "^") {
		simplified = strings.TrimPrefix(simplified, "^")
	}

	// For very complex alternation patterns, use a fallback
	if strings.Count(simplified, "|") > 10 {
		// Too many alternations, create a simpler pattern
		return "[a-zA-Z_][a-zA-Z0-9_]*" // Generic identifier pattern
	}

	return simplified
}

// validateAndFixPostfix validates and fixes a postfix expression
func validateAndFixPostfix(output []string, debug *debugger.Debug, tokenName string) []string {
	if len(output) == 0 {
		return output
	}

	// Count operands and operators
	operandCount := 0
	operatorCount := 0
	for _, token := range output {
		switch token {
		case "·", "|":
			operatorCount++
		case "*", "+", "?":
			// Unary operators don't change the operand count
		default:
			operandCount++
		}
	}

	// A valid postfix expression should have exactly one more operand than binary operator
	expectedOperands := operatorCount + 1

	if operandCount != expectedOperands {
		if debug != nil {
			debug.DebugLog(fmt.Sprintf("[%s] Invalid postfix: %d operands, %d operators (expected %d operands)",
				tokenName, operandCount, operatorCount, expectedOperands), false)
		}

		// Fix common issues
		fixed := make([]string, 0, len(output))

		// Remove trailing concatenation operators that cause imbalance
		for i, token := range output {
			if token == "·" && i == len(output)-1 && operandCount < expectedOperands {
				// Skip trailing concatenation
				continue
			}
			fixed = append(fixed, token)
		}

		// If still invalid, create a minimal valid expression
		if len(fixed) == 0 {
			return []string{".*"} // Match anything
		}

		return fixed
	}

	return output
}

// createState creates a new NFA state with the given properties
func createState(isAccepting bool, tokenType TokenType) *NFAState {
	return &NFAState{
		id:          generateStateID(),
		isAccepting: isAccepting,
		tokenType:   tokenType,
		transitions: make(map[string][]*NFAState),
	}
}

// addCharRangeTransitions adds transitions for a range of characters
func addCharRangeTransitions(start *NFAState, end *NFAState, fromChar, toChar byte) {
	for c := fromChar; c <= toChar; c++ {
		add_transition(start, string(c), end)
	}
}

// createFallbackNFA creates a simple NFA that matches word characters (fallback)
func createFallbackNFA(tokenType TokenType) *NFA {
	start := createState(false, tokenType)
	end := createState(true, tokenType)

	addCharRangeTransitions(start, end, 'a', 'z')
	addCharRangeTransitions(start, end, 'A', 'Z')
	addCharRangeTransitions(start, end, '0', '9')
	add_transition(start, "_", end)
	add_transition(end, "ε", start)

	return &NFA{start, end}
}

// postfix converts a tokenized regex to postfix (RPN) using the shunting yard algorithm
func postfix(regex string, tokenName string, debug *debugger.Debug) string {
	if debug != nil {
		debug.DebugLog(fmt.Sprintf("[%s] RAW PATTERN: %q", tokenName, regex), false)
	}

	// Simplify complex patterns first
	simplified := simplifyRegex(regex)
	if simplified != regex && debug != nil {
		debug.DebugLog(fmt.Sprintf("[%s] SIMPLIFIED: %q -> %q", tokenName, regex, simplified), false)
	}

	tokens := tokenizeRegex(simplified)

	// Debug: print tokens if debug enabled
	if debug != nil {
		var tokenStrs strings.Builder
		for i, t := range tokens {
			if i > 0 {
				tokenStrs.WriteByte(' ')
			}
			tokenStrs.WriteString(fmt.Sprintf("%s(%s)", t.Value, t.Type))
		}
		debug.DebugLog(fmt.Sprintf("[%s] Tokens: %s", tokenName, tokenStrs.String()), false)
	}

	// Operator precedence
	precedence := map[string]int{
		"*": 3,
		"+": 3,
		"?": 3,
		"{": 3, // range quantifiers
		"·": 2, // explicit concatenation
		"|": 1,
	}
	isOperator := func(t RegexToken) bool {
		return t.Type == "operator" && (t.Value == "*" || t.Value == "+" || t.Value == "?" || t.Value == "·" || t.Value == "|" || t.Value == "(" || t.Value == ")" || strings.HasPrefix(t.Value, "{"))
	}
	// Step 1: Insert explicit concatenation operators
	isValue := func(t RegexToken) bool {
		return t.Type == "literal" || t.Type == "class" || t.Type == "escape" || t.Type == "anchor" ||
			(t.Type == "operator" && (t.Value == ")" || t.Value == "*" || t.Value == "+" || t.Value == "?" || strings.HasPrefix(t.Value, "{")))
	}

	var explicit []RegexToken
	for i := 0; i < len(tokens)-1; i++ {
		t := tokens[i]
		next := tokens[i+1]
		explicit = append(explicit, t)
		// Don't add concatenation if the next token is an alternation
		if isValue(t) && (next.Type == "literal" || next.Type == "class" || next.Type == "escape" ||
			(next.Type == "operator" && next.Value == "(")) &&
			!(next.Type == "operator" && next.Value == "|") {
			explicit = append(explicit, RegexToken{"·", "operator"})
		}
	}
	if len(tokens) > 0 {
		explicit = append(explicit, tokens[len(tokens)-1])
	}

	// Debug: print explicit tokens if debug enabled
	if debug != nil {
		var explicitStrs strings.Builder
		for i, t := range explicit {
			if i > 0 {
				explicitStrs.WriteByte(' ')
			}
			explicitStrs.WriteString(fmt.Sprintf("%s(%s)", t.Value, t.Type))
		}
		debug.DebugLog(fmt.Sprintf("[%s] Explicit: %s", tokenName, explicitStrs.String()), false)
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

	// Verify and fix the postfix expression
	output = validateAndFixPostfix(output, debug, tokenName)

	if debug != nil {
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

	// Validate postfix expression
	validatePostfix := func(tokens []string) bool {
		var stack int
		for _, tok := range tokens {
			switch tok {
			case "·", "|":
				stack--
				if stack < 1 {
					return false
				}
			case "*", "+", "?":
				// Unary operators don't change stack size
			default:
				stack++
			}
		}
		return stack == 1
	}

	if !validatePostfix(tokens) {
		return createFallbackNFA(tokenType)
	}

	for _, tok := range tokens {
		switch tok {
		case "·", "|": // Binary operators
			if len(stack) < 2 {
				panic("thompsonConstruct: not enough operands for binary operator")
			}
			b := stack[len(stack)-1].nfa
			a := stack[len(stack)-2].nfa
			stack = stack[:len(stack)-2]

			if tok == "·" {
				add_transition(a.end, "ε", b.start)
				stack = append(stack, nfaStackElem{&NFA{a.start, b.end}})
			} else { // "|"
				start := createState(false, tokenType)
				end := createState(true, tokenType)
				add_transition(start, "ε", a.start)
				add_transition(start, "ε", b.start)
				add_transition(a.end, "ε", end)
				add_transition(b.end, "ε", end)
				stack = append(stack, nfaStackElem{&NFA{start, end}})
			}
		case "*", "+", "?": // Unary operators
			if len(stack) < 1 {
				panic("thompsonConstruct: not enough operands for unary operator")
			}
			a := stack[len(stack)-1].nfa
			stack = stack[:len(stack)-1]
			start := createState(false, tokenType)
			end := createState(true, tokenType)

			switch tok {
			case "*": // Kleene star
				add_transition(start, "ε", a.start)
				add_transition(start, "ε", end)
				add_transition(a.end, "ε", a.start)
				add_transition(a.end, "ε", end)
			case "+": // One or more
				add_transition(start, "ε", a.start)
				add_transition(a.end, "ε", a.start)
				add_transition(a.end, "ε", end)
			case "?": // Zero or one
				add_transition(start, "ε", a.start)
				add_transition(start, "ε", end)
				add_transition(a.end, "ε", end)
			}
			stack = append(stack, nfaStackElem{&NFA{start, end}})
		case "{": // range quantifiers {n}, {n,}, {n,m}
			if len(stack) < 1 {
				panic("thompsonConstruct: not enough operands for range quantifier")
			}
			a := stack[len(stack)-1].nfa
			stack = stack[:len(stack)-1]

			// Parse quantifier
			parts := strings.Split(tok[1:len(tok)-1], ",")
			min := parseInt(parts[0])
			max := min
			if len(parts) == 2 {
				if parts[1] == "" {
					max = -1
				} else {
					max = parseInt(parts[1])
				}
			}
			if min < 0 || (max != -1 && max < min) {
				panic("thompsonConstruct: invalid range quantifier values")
			}

			// build the NFA for the range quantifier
			start := createState(false, tokenType)
			end := createState(true, tokenType)

			// Handle special case: {0} or {0,0}
			if min == 0 && (max == 0 || max == -1) {
				add_transition(start, "ε", end)
				stack = append(stack, nfaStackElem{&NFA{start, end}})
				continue
			}

			// create exactly min copies
			current := start
			for i := 0; i < min; i++ {
				// copy the NFA
				copied := copyNFA(a)
				add_transition(current, "ε", copied.start)
				current = copied.end
			}

			// max > min
			if max == -1 || max > min {
				// add optional loop for remaining copies
				optionalStart := current
				if max == -1 {
					// For unbounded {n,}, create a single optional loop
					copied := copyNFA(a)
					add_transition(current, "ε", copied.start)
					add_transition(copied.end, "ε", optionalStart)
					current = copied.end
				} else {
					// For bounded {n,m}, create exactly remaining copies
					remaining := max - min
					for i := 0; i < remaining; i++ {
						copied := copyNFA(a)
						add_transition(current, "ε", copied.start)
						current = copied.end
						// add epsilon transition back to optional start
						add_transition(current, "ε", optionalStart)
					}
				}
			}

			add_transition(current, "ε", end)
			stack = append(stack, nfaStackElem{&NFA{start, end}})
		default:
			// Handle different token types
			start := createState(false, tokenType)
			end := createState(true, tokenType)

			if strings.HasPrefix(tok, "[") && strings.HasSuffix(tok, "]") {
				// Character class - parse and create transitions for each character
				classContent := tok[1 : len(tok)-1] // Remove brackets
				if strings.HasPrefix(classContent, "^") {
					// Negated character class
					chars := parseCharacterClass(classContent[1:])

					// Create a set for O(1) lookup
					charSet := make(map[string]bool, len(chars))
					for _, char := range chars {
						charSet[char] = true
					}

					for i := 0; i < 128; i++ {
						char := string(byte(i))
						if !charSet[char] {
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
					case 'n', 't', 'r', '\\', '"', '\'':
						escapeMap := map[byte]string{
							'n': "\n", 't': "\t", 'r': "\r",
							'\\': "\\", '"': "\"", '\'': "'",
						}
						add_transition(start, escapeMap[escapedChar], end)
					case 'b':
						// Word boundary position assertion
						add_transition(start, "word_boundary", end)
					case 'd':
						// Digit class
						addCharRangeTransitions(start, end, '0', '9')
					case 's':
						// Whitespace class
						for _, ws := range []string{" ", "\t", "\n", "\r"} {
							add_transition(start, ws, end)
						}
					case 'w':
						// Word character class
						addCharRangeTransitions(start, end, 'a', 'z')
						addCharRangeTransitions(start, end, 'A', 'Z')
						addCharRangeTransitions(start, end, '0', '9')
						add_transition(start, "_", end)
					default:
						// Any other escaped character
						add_transition(start, string(escapedChar), end)
					}
				}
			} else if tok == "^" || tok == "$" || tok == "\\b" {
				// Position assertions
				anchorMap := map[string]string{
					"^": "start_anchor", "$": "end_anchor", "\\b": "word_boundary",
				}
				add_transition(start, anchorMap[tok], end)
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
		panic(fmt.Sprintf("thompsonConstruct: stack did not end with exactly one NFA, got %d. Postfix: %s", len(stack), postfix))
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
				chars = append(chars, string(start), "-", string(end))
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

// safely convert a string to an integer
func parseInt(s string) int {
	var result int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c-'0')
		}
	}
	return result
}

// create a deep copy of an NFA
func copyNFA(original *NFA) *NFA {
	stateMap := make(map[int]*NFAState)

	// Create new states
	var copyStates func(*NFAState) *NFAState
	copyStates = func(state *NFAState) *NFAState {
		if newState, exists := stateMap[state.id]; exists {
			return newState
		}

		newState := createState(state.isAccepting, state.tokenType)
		stateMap[state.id] = newState

		// copy transitions
		for symbol, targets := range state.transitions {
			for _, target := range targets {
				newTarget := copyStates(target)
				newState.transitions[symbol] = append(newState.transitions[symbol], newTarget)
			}
		}

		return newState
	}

	newStart := copyStates(original.start)
	newEnd := copyStates(original.end)

	return &NFA{start: newStart, end: newEnd}
}

func (nfa *NFA) Print(debug *debugger.Debug) {
	visited := make(map[int]bool)
	queue := []*NFAState{nfa.start}
	debug.DebugLog("NFA Structure:", false)

	for len(queue) > 0 {
		state := queue[0]
		queue = queue[1:]
		if visited[state.id] {
			continue
		}
		visited[state.id] = true

		status := fmt.Sprintf("State %d", state.id)
		if state.isAccepting {
			status += fmt.Sprintf(" [accepting, type=%s]", state.tokenType.String())
		}
		debug.DebugLog(status, false)

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
	closure := make(map[int]*NFAState, len(states))
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
