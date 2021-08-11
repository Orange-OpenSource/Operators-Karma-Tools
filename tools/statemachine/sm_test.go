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
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	Start LCGState = iota
	Run
	Service
	Stop
	End
)

var tGraph = LCGGraph{
	Start: LCGNodeInfo{
		Name: "Start",
		Children: LCGChildren{
			End,
			Run,
		},
	},
	Run: LCGNodeInfo{
		Name: "Run",
		Children: LCGChildren{
			Service,
			Stop,
		},
	},
	Service: LCGNodeInfo{
		Name: "Servicing",
		Children: LCGChildren{
			Stop,
			Run,
		},
	},
	Stop: LCGNodeInfo{
		Name: "Stopping",
		Children: LCGChildren{
			End,
		},
	},
	End: LCGNodeInfo{Name: "End"},
}

type OKTDatabase struct {
	//sm *Machine
}

var _ LCGStateAction = &OKTDatabase{}

func (db *OKTDatabase) Enter(state LCGState) error {
	stateName := tGraph.StateName(state)
	fmt.Println("Enter in state: ", stateName)

	switch state {
	case Start:
		fmt.Println("The DB is starting")
	case Run:
		fmt.Println("The DB is running now")
	case Service:
		fmt.Println("The DB is in a servicing operation")
	case Stop:
		fmt.Println("The DB is stopping")
	case End:
		fmt.Println("The livecycle machine of the DB is OFF")
	}

	return nil
}

func TestStateMachine(t *testing.T) {
	db := &OKTDatabase{}
	sm := &Machine{Graph: tGraph, Actions: db}

	off := sm.IsOFF()
	require.True(t, off, "At creation machine should be OFF")

	pStatus := sm.IsPathInGraphEnabled()
	require.False(t, pStatus, "PathInGraph should be disabled by default")
	sm.EnablePathInGraph()
	pStatus = sm.IsPathInGraphEnabled()
	require.True(t, pStatus, "PathInGraph should be enabled afeter Init")

	entered := sm.SetState(Start)
	require.True(t, entered, "Should be entered in Start state")
	off = sm.IsOFF()
	require.False(t, off, "Now, after entering in a state, the machine should be ON")

	var err error

	state := sm.GetState()
	require.Equal(t, Start, state, "Current state must be Start")
	entered, err = sm.EnterNextState(LCGEvents{End})
	require.True(t, entered, "Must be entered in state End")
	require.Nil(t, err, "No error must be raised in End state")

	path := sm.GetPathInGraph()
	fmt.Println(path)

	off = sm.IsOFF()
	require.True(t, off, "The state machine should be OFF after traversing a leaf node in the graph")
	state = sm.GetState()
	require.Equal(t, DefaultState, state, "Current state must be End")

	entered, err = sm.EnterNextState(LCGEvents{Start})
	require.False(t, entered, "Can't go to the next state when the machine is OFF. SetState() must be called before")
	require.Nil(t, err, "No error must be raised when a state has not been browsed")

	entered = sm.SetState(Run)
	require.True(t, entered, "Must be entered in state Run")
	state = sm.GetState()
	require.Equal(t, Run, state, "Current state must be Run")

	entered, err = sm.EnterNextState(LCGEvents{Start})
	require.False(t, entered, "Start is not a child state of Run")
	require.Nil(t, err, "No error must be raised when a state has not been browsed")

	entered, err = sm.EnterNextState(LCGEvents{Service})
	require.True(t, entered, "Must be entered in state of Service")
	require.Nil(t, err, "No error must be raised when a state has not been browsed")

	for count := 11; count > 0; count-- {
		entered, err = sm.EnterNextState(LCGEvents{DefaultState})
		require.True(t, entered, "Must be entered in state of Service")
		require.Nil(t, err, "No error must be raised here")
		state = sm.GetState()
		require.Equal(t, Run, state, "Current state must be Run now it is the Default after Service")

		entered, err = sm.EnterNextState(LCGEvents{Service})
		require.True(t, entered, "Must be entered in state of Service")
		require.Nil(t, err, "No error must be raised here")
	}

	entered, err = sm.EnterNextState(LCGEvents{Run, Stop}) // Priority test
	require.True(t, entered, "Must be entered in Run state")
	require.Nil(t, err, "No error must be raised here")
	state = sm.GetState()
	require.Equal(t, Stop, state, "Current state must be Stop as Run is not the prior state")

	entered, err = sm.EnterNextState(LCGEvents{DefaultState})
	require.True(t, entered, "Must be entered in  state")
	require.Nil(t, err, "No error must be raised here")
	off = sm.IsOFF()
	require.True(t, off, "The state machine should be OFF after traversing a leaf node in the graph")
	state = sm.GetState()
	require.Equal(t, DefaultState, state, "Current state must undefined/default when machine is OFF")

	path = sm.GetPathInGraph()
	fmt.Println(path)

	// Start>Run>Stop>End
	off = sm.IsOFF()
	require.True(t, off, "The state machine should be OFF after traversing a leaf node in the graph")
	entered = sm.SetState(Start)
	require.True(t, entered, "Should be entered in Start state")

	for sm.IsOFF() == false {
		entered, _ = sm.EnterNextState(LCGEvents{DefaultState})
		require.True(t, entered, "Should be entered in Start state")
	}
	require.True(t, entered, "Should be entered in state")

	path = sm.GetPathInGraph()
	fmt.Println(path)
}
