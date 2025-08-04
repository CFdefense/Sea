package lexer

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type DFAState struct {
	id          int
	isAccepting bool
	tokenType   TokenType
	transitions map[string]*DFAState
	nfaStates   []*NFAState // represented nfa state
}

type DFA struct {
	start    *DFAState
	states   []*DFAState
	alphabet map[string]bool // epsilon-exclusive set of symbols
}

var generateDFAStateID = func() func() int {
	var counter int
	return func() int {
		counter++
		return counter
	}
}()

// i think this is why it was hanging, it was trying to make massive dfas
// this is temporary, maybe christian has a solution :)
const MAX_DFA_STATES = 1000
const MAX_CONVERSION_TIME_MS = 5000

// convert nfa to dfa using subset construction
func ConvertNFAtoDFA(nfa *NFA) *DFA {
	alphabet := getAlphabet(nfa)

	initialNFAStates := nfa.epsilonClosure([]*NFAState{nfa.start})
	initialDFAState := createDFAState(initialNFAStates)

	dfa := &DFA{
		start:    initialDFAState,
		states:   []*DFAState{initialDFAState},
		alphabet: alphabet,
	}

	stateMap := make(map[string]*DFAState)
	stateMap[getNFAStateSetID(initialNFAStates)] = initialDFAState

	// queue of states to process
	queue := []*DFAState{initialDFAState}

	// track start time for timeout (again, temporary)
	startTime := time.Now()

	for len(queue) > 0 {
		if len(dfa.states) > MAX_DFA_STATES {
			fmt.Printf("warn: DFA conversion stopped due to state explosion (> %d states)\n", MAX_DFA_STATES)
			return createSimplifiedDFA(nfa)
		}
		
		if time.Since(startTime).Milliseconds() > MAX_CONVERSION_TIME_MS {
			fmt.Printf("warn: DFA conversion timeout after %d ms\n", MAX_CONVERSION_TIME_MS)
			return createSimplifiedDFA(nfa)
		}

		currentState := queue[0]
		queue = queue[1:] // remove first element

		for symbol := range alphabet {
			moveResult := move(currentState.nfaStates, symbol)
			if len(moveResult) == 0 {
				continue
			}

			nextNFAStates := nfa.epsilonClosure(moveResult)
			if len(nextNFAStates) == 0 {
				continue
			}

			nextStateID := getNFAStateSetID(nextNFAStates)

			var nextDFAState *DFAState
			if existingState, exists := stateMap[nextStateID]; exists {
				nextDFAState = existingState
			} else {
				nextDFAState = createDFAState(nextNFAStates)
				stateMap[nextStateID] = nextDFAState
				dfa.states = append(dfa.states, nextDFAState)
				queue = append(queue, nextDFAState)
			}

			if currentState.transitions == nil {
				currentState.transitions = make(map[string]*DFAState)
			}
			currentState.transitions[symbol] = nextDFAState
		}
	}

	return dfa
}

// fallback dfa, again temporary, because i was struggling
func createSimplifiedDFA(nfa *NFA) *DFA {
	start := &DFAState{
		id:          generateDFAStateID(),
		isAccepting: nfa.start.isAccepting,
		tokenType:   nfa.start.tokenType,
		transitions: make(map[string]*DFAState),
		nfaStates:   []*NFAState{nfa.start},
	}
	
	accept := &DFAState{
		id:          generateDFAStateID(),
		isAccepting: true,
		tokenType:   nfa.end.tokenType,
		transitions: make(map[string]*DFAState),
		nfaStates:   []*NFAState{nfa.end},
	}
	
	alphabet := getAlphabet(nfa)
	for symbol := range alphabet {
		start.transitions[symbol] = accept
	}
	
	return &DFA{
		start:    start,
		states:   []*DFAState{start, accept},
		alphabet: alphabet,
	}
}

func createDFAState(nfaStates []*NFAState) *DFAState {
	dfaState := &DFAState{
		id:          generateDFAStateID(),
		isAccepting: false,
		transitions: make(map[string]*DFAState),
		nfaStates:   nfaStates,
	}

	// a DFA state is accepting if any of its NFA states is accepting
	// if multiple accepting NFA states with different token types,
	// choose the one with highest precedence (lowest enum value)
	for _, nfaState := range nfaStates {
		if nfaState.isAccepting {
			if !dfaState.isAccepting || nfaState.tokenType < dfaState.tokenType {
				dfaState.isAccepting = true
				dfaState.tokenType = nfaState.tokenType
			}
		}
	}

	return dfaState
}

func getNFAStateSetID(states []*NFAState) string {
	ids := make([]int, len(states))
	for i, state := range states {
		ids[i] = state.id
	}
	sort.Ints(ids)

	var sb strings.Builder
	for i, id := range ids {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf("%d", id))
	}
	return sb.String()
}

func getAlphabet(nfa *NFA) map[string]bool {
	alphabet := make(map[string]bool)
	visited := make(map[int]bool)

	var collectSymbols func(*NFAState)
	collectSymbols = func(state *NFAState) {
		if visited[state.id] {
			return
		}
		visited[state.id] = true

		for symbol := range state.transitions {
			if symbol != "ε" && !isSpecialSymbol(symbol) {
				alphabet[symbol] = true
			}
		}

		for _, targets := range state.transitions {
			for _, target := range targets {
				collectSymbols(target)
			}
		}
	}

	collectSymbols(nfa.start)
	
	if len(alphabet) > 128 {
		fmt.Printf("warning: alphabet size %d is too large, limiting to ASCII range\n", len(alphabet))
		alphabet = make(map[string]bool)
		for i := 0; i < 128; i++ {
			alphabet[string(byte(i))] = true
		}
	}
	
	return alphabet
}

func isSpecialSymbol(symbol string) bool {
	specialSymbols := map[string]bool{
		"word_boundary": true,
		"start_anchor":  true,
		"end_anchor":    true,
		"any":           true,
	}
	return specialSymbols[symbol]
}

func move(states []*NFAState, symbol string) []*NFAState {
	var result []*NFAState
	seen := make(map[int]bool)

	for _, state := range states {
		if targets, exists := state.transitions[symbol]; exists {
			for _, target := range targets {
				if !seen[target.id] {
					seen[target.id] = true
					result = append(result, target)
				}
			}
		}
	}

	return result
}

func (dfa *DFA) SimulateDFA(input string) (bool, TokenType) {
	currentState := dfa.start

	for _, char := range input {
		symbol := string(char)
		nextState, exists := currentState.transitions[symbol]
		if !exists {
			nextState, exists = currentState.transitions["any"]
			if !exists {
				return false, 0 // reject
			}
		}
		currentState = nextState
	}

	// accept if the final state is accepting
	return currentState.isAccepting, currentState.tokenType
}

func (dfa *DFA) PrintDFA() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("DFA with %d states:\n", len(dfa.states)))

	for _, state := range dfa.states {
		status := fmt.Sprintf("State %d", state.id)
		if state.isAccepting {
			status += fmt.Sprintf(" [accepting, type=%s]", state.tokenType.String())
		}
		sb.WriteString(status + "\n")

		for symbol, target := range state.transitions {
			sb.WriteString(fmt.Sprintf("  %d --%s--> %d\n", state.id, symbol, target.id))
		}
	}

	return sb.String()
}
