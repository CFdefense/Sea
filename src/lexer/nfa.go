package lexer

// NFA state structure
type NFAState struct {
	id                 int
	isAccepting        bool
	tokenType          TokenType
	transitions        map[rune][]*NFAState
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

// thompson's construction
// TODO: Implement true Thompson construction with regex parsing
// For now, create placeholder NFAs that will be replaced with proper construction
func thompsonConstruct(regexPattern string, tokenType TokenType) *NFA {
	// TODO: Parse regex to postfix notation and build proper NFA
	// For now, create a placeholder NFA structure
	startState := &NFAState{
		id:                 generateStateID(),
		isAccepting:        false,
		tokenType:          tokenType,
		transitions:        make(map[rune][]*NFAState),
		epsilonTransitions: []*NFAState{},
	}

	endState := &NFAState{
		id:                 generateStateID(),
		isAccepting:        true,
		tokenType:          tokenType,
		transitions:        make(map[rune][]*NFAState),
		epsilonTransitions: []*NFAState{},
	}

	nfa := &NFA{
		start: startState,
		end:   endState,
	}

	return nfa
}

// build a placeholder NFA for a token that doesn't have a regex pattern
func buildSimpleNFA(tokenType TokenType, isAccepting bool) *NFA {
	start := &NFAState{
		id:                 generateStateID(),
		isAccepting:        isAccepting,
		tokenType:          tokenType,
		transitions:        make(map[rune][]*NFAState),
		epsilonTransitions: []*NFAState{},
	}

	end := &NFAState{
		id:                 generateStateID(),
		isAccepting:        isAccepting,
		tokenType:          tokenType,
		transitions:        make(map[rune][]*NFAState),
		epsilonTransitions: []*NFAState{},
	}

	// TODO: Add proper transitions based on token type
	// For now, no transitions (placeholder)

	return &NFA{start: start, end: end}
}

// function to replace thompsonConstruct with true Thompson construction
func parseRegexToPostfix(regexPattern string) string {
	// This is a placeholder for a proper regex to postfix parser.
	// In a real implementation, you would use a regex engine to parse the pattern
	// and convert it to postfix notation.
	// For now, we'll just return the pattern as a placeholder.
	return regexPattern
}

// function to replace thompsonConstruct with true Thompson construction
func buildNFAFromPostfix(postfix string, tokenType TokenType) *NFA {
	// This is a placeholder for a proper postfix to NFA builder.
	// In a real implementation, you would use a regex engine to parse the postfix
	// and build the NFA state machine.
	// For now, we'll just return a placeholder NFA.
	startState := &NFAState{
		id:                 generateStateID(),
		isAccepting:        false,
		tokenType:          tokenType,
		transitions:        make(map[rune][]*NFAState),
		epsilonTransitions: []*NFAState{},
	}

	endState := &NFAState{
		id:                 generateStateID(),
		isAccepting:        true,
		tokenType:          tokenType,
		transitions:        make(map[rune][]*NFAState),
		epsilonTransitions: []*NFAState{},
	}

	nfa := &NFA{
		start: startState,
		end:   endState,
	}

	return nfa
}

func epsilonClosure(states []*NFAState) []*NFAState {
	// 1. Start with input states
	closure := make(map[int]*NFAState)
	stack := make([]*NFAState, len(states))
	copy(stack, states)

	// Add initial states to closure
	for _, state := range states {
		closure[state.id] = state
	}

	// 2. Add all states reachable via ε-transitions
	// 3. Repeat until no new states added
	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		// Add all ε-transitions
		for _, nextState := range current.epsilonTransitions {
			if _, exists := closure[nextState.id]; !exists {
				closure[nextState.id] = nextState
				stack = append(stack, nextState)
			}
		}
	}

	// 4. Return closure set
	result := make([]*NFAState, 0, len(closure))
	for _, state := range closure {
		result = append(result, state)
	}
	return result
}
