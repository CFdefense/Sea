package lexer

import "regexp"

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

// An NFA represening a Fragment of a larger NFA
type Fragment struct {
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
// TODO: Parse regex to postfix notation and build proper NFA
// For now, create a placeholder NFA structure
func thompsonConstruct(regexPattern string, tokenType TokenType) *NFA {
	single_literal := regexp.MustCompile(`^(\\.|[^\\.*+?|()[\]{}^$])$`)
	concatenation :=
	alternation := 
	kleen_star :=
	if single_literal.MatchString(regexPattern) {
		s1 := &NFAState{
			id:                 generateStateID(),
			isAccepting:        false,
			tokenType:          tokenType,
			transitions:        make(map[rune][]*NFAState),
			epsilonTransitions: []*NFAState{},
		}
		s2 := &NFAState{
			id:                 generateStateID(),
			isAccepting:        true,
			tokenType:          tokenType,
			transitions:        make(map[rune][]*NFAState),
			epsilonTransitions: []*NFAState{},
		}

	}
}

func add_transition(from_state NFAState, input string, to_state NFAState) {
	from_state.transitions.append(input, to_state)
}
    
func connect(frag1 Fragment, frag2 Fragment) *Fragment {
    add_transition(frag1.accept_state, ε, frag2.start_state)
    return Fragment(frag1.start_state, frag2.accept_state)
}