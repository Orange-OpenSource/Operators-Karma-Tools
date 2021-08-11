// Copyright 2021 Orange SA
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package apis

package statemachine

import (
	"strconv"
)

/** The StateMachine allows to build a LCG (Life Cycle Graph) to represent the multiple states an aplication
    will go through.
	At each state, an action can be executed through a Hook Caller.
**/

type LCGState int

func (s LCGState) String() string {
	return strconv.Itoa(int(s))
}

const DefaultState LCGState = -1

// LCGChildren defines the children nodes oredered by priority. The first has the most important priority.
// The last is de default, the one to choose when DefaultState has been added in the event list and
// no other event match with a child.
type LCGChildren []LCGState

func (t LCGChildren) Default() (exists bool, dft LCGState) {
	length := len(t)
	if length == 0 {
		return false, 0
	}

	return true, t[length-1]
}

func (t LCGChildren) IsDefault(state LCGState) bool {
	exists, dft := t.Default()
	return (exists && state == dft)
}

type LCGNodeInfo struct {
	Name     string
	Children LCGChildren
}

type LCGGraph map[LCGState]LCGNodeInfo

func (g LCGGraph) IsLeafNode(state LCGState) bool {
	return len(g[state].Children) == 0
}

func (g LCGGraph) StateName(state LCGState) (name string) {
	name = g[state].Name
	if name == "" {
		name = state.String()
	}
	return name
}

// LCGStateAction calls
// The Hook "Enter" is called each time the machine enter in a state.
// This is the action to do on a state.
type LCGStateAction interface {
	Enter(state LCGState) error
}

type Machine struct {
	Graph           LCGGraph
	Actions         LCGStateAction
	curState        LCGState
	prevState       LCGState
	pathInGraph     []string
	pathLoopsCount  uint
	pathLengthLimit uint
}

func (m Machine) IsOFF() bool {
	// 0 when the machine is instanciated without any browsing or when set to OFF after traversing a leaf node
	return m.prevState == m.curState
}

func (m *Machine) setOFF() {
	m.prevState = DefaultState
	m.curState = DefaultState
}

// SetState Set current state for this mathine
// Can not set the same state twice
func (m *Machine) SetState(state LCGState) (entered bool) {
	if m.curState != state || m.IsOFF() {
		if m.IsPathInGraphEnabled() {
			m.EnablePathInGraph() // Reset path for a new browsing
		}
		m.enterInState(state)
		return true
	}
	return false
}

// GetState Get current state for this machine
func (m *Machine) GetState() LCGState {
	return m.curState
}

type LCGEvents []LCGState

func (e LCGEvents) Contains(state LCGState) bool {
	for _, elem := range e {
		if elem == state {
			return true
		}
	}
	return false
}

func (e LCGEvents) IsTriggeringState(state LCGState, isDefaultState bool) bool {
	if e.Contains(state) ||
		(isDefaultState && e.Contains(DefaultState)) {
		return true
	}
	return false
}

// enterInState Set the new current state
// The machine is set to OFF if browsing a leaf node
func (m *Machine) enterInState(state LCGState) (err error) {
	var loopPathLength uint

	if m.IsOFF() {
		// Hack to ensure that a never used machine will appear ON
		// after entering in state 0 (because all is initialized at 0 at struct creation)
		m.prevState = DefaultState
	} else {
		if m.prevState == state {
			loopPathLength = 2 // The length of the path in the loop is 2 nodes
		}
		m.prevState = m.curState
	}
	m.curState = state

	m.addStateToPath(state, loopPathLength)

	if m.Actions != nil {
		err = m.Actions.Enter(m.curState)
	}

	// Browsing is at its end (in a leaf node), thus set the machine to OFF
	if m.Graph.IsLeafNode(state) {
		m.setOFF()
	}

	return err
}

// EnterNextState Trigger each event up to the first allowing to throw a new state.
func (m *Machine) EnterNextState(events LCGEvents) (entered bool, err error) {
	children := m.Graph[m.curState].Children
	exists, defaultChild := children.Default()
	if !exists {
		return false, nil
	}

	for _, nextState := range children {
		if events.IsTriggeringState(nextState, (nextState == defaultChild)) {
			// Enter into next state !!
			err = m.enterInState(nextState)
			return true, err
		}
	}

	return false, err
}

func (m *Machine) EnablePathInGraph() {
	/*
		m.pathInGraph = make([]string, 0)
		m.appendToPath(">") // Initialize with the first element
	*/
	m.pathInGraph = []string{">"}
	if m.pathLengthLimit == 0 {
		m.pathLengthLimit = 512
	}
}

func (m *Machine) DisablePathInGraph() {
	m.pathInGraph = nil
}

func (m *Machine) IsPathInGraphEnabled() bool {
	return m.pathInGraph != nil
}

// SetPathLengthLimit Defines the size of the slice storing the path in graph, i.e. the maximum states to store. Nnote that loops in graph count for 2 states.
// This slice store also the separator (">") in addtion to the states.
// If not set, the max is by default limited to 512. You can define more if needed.
// Min is 5 and Maximum is 1024
func (m *Machine) SetPathLengthLimit(max uint) {
	if max < 5 {
		max = 5
	} else if max > 1024 {
		max = 1024
	}
	m.pathLengthLimit = max
}

func (m *Machine) GetPathInGraph() (path string) {
	path = ""
	for _, p := range m.pathInGraph {
		path += p
	}
	return path
}

func (m *Machine) appendToPath(elem ...string) {
	m.pathInGraph = append(m.pathInGraph, elem...)

	if len(m.pathInGraph) > int(m.pathLengthLimit) {
		m.pathInGraph = m.pathInGraph[2:]
	}
}

/*
func (m *Machine) compactAndAppendToPath(elem ...string) {
	newLength := len(m.pathInGraph) - int(m.pathCompactionNeeded) - int(m.pathCompactionNeeded/2)
	m.pathInGraph = m.pathInGraph[:newLength]
	m.pathInGraph = append(m.pathInGraph, elem...)
}
*/

// TODO manage maximum path length to limit its size ?
func (m *Machine) addStateToPath(state LCGState, loopPathLength uint) {
	if m.pathInGraph == nil {
		return
	}

	stateName := m.Graph.StateName(state)

	if m.Graph.IsLeafNode(state) {
		m.appendToPath(stateName)
		m.pathLoopsCount = 0
	} else {
		if loopPathLength == 0 { // There's no loop
			m.pathLoopsCount = 0
			m.appendToPath(stateName, ">")
			return
		}
		m.pathLoopsCount++

		pathLen := len(m.pathInGraph)
		m.pathInGraph[pathLen-int(loopPathLength*2+1)] = ">("
		m.pathInGraph[pathLen-3] = "<<>>"
		count := float32(m.pathLoopsCount)/2.0 + 1.0
		//loops := fmt.Sprintf("%.1f", count)
		if float32(int(count)) == count {
			m.pathInGraph[pathLen-1] = ")x" + strconv.Itoa(int(count)) + ">"
		} else {
			m.pathInGraph[pathLen-1] = ")x" + strconv.Itoa(int(count)) + ">" + stateName + ">"
		}
	}

}
