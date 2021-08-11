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

/**
  This file is intended to replace the reconciler engine file in `reconciler/engine/stepper.go`
  For the moment the test does not pass since a weird behaviour of the Build() function that should be fixed quickly.

  The Graph below is an example of LCGraph and Reconciliation states.
**/
package engines

import (
	"fmt"

	oktreconciler "gitlab.tech.orange/dbmsprivate/operators/okt/reconciler"
	okterr "gitlab.tech.orange/dbmsprivate/operators/okt/results"
	oktsm "gitlab.tech.orange/dbmsprivate/operators/okt/tools/statemachine"

	"github.com/go-logr/logr"
)

// StepperEngineHook Defines the callback used by this engine at each step of the reconciliation.
// Step name is the state name (CRChecker, ObjectGetter, Mutator, Update, ErrorManager, SuccessManager)
type StepperEngineHook interface {
	EnterInState(engine *Stepper)
}

/*
// hookCaller calls
type hookCaller interface {
	callHook(smc interface{}, tx *gorm.DB) error
	//callLocalHook(smc interface{}, tx *gorm.DB) error
}
*/

func giveUpStateHook(ctx interface{}) error {
	smc := ctx.(*Stepper)
	//smc.Logger.Info("------- GIVEN UP reconciliation state reached -------")
	smc.Logger.Info(okterr.ErrGiveUpReconciliation.Error())

	if _, err := smc.ConsolidatedError(); err != nil {
		smc.Logger.Error(err, "Something is missing or definitively wrong. Give up this reconciliation")
	}
	return nil
}

func endStateHook(ctx interface{}) error {
	smc := ctx.(*Stepper)
	smc.DisplayPathOfStates(smc.Logger)
	return nil
}

// Stepper is an OKT reconciler engine which reates a Reconciliation state machine.
// It can be used to implement an idempotent Reconcile function for an Operator Controller
//
// Reconciliation states :
//      CRChecker           // Check the Custom Resource received by the Controller (from the queue)
//      ObjectsGetter       // Create all OKT resources types manipulated by this Operator, register them to the OKT Reconciler
//      Mutator			    // Mutates OKT resources to the desired state
//      Updater				// Create or Updates OKT resources on Cluster
//		SuccessManager		// Is reached after the Updater state in case of success
//
//  Here are the debranching steps occuring on Finalization, Error, or Give-up events:
//		CRFinalizer			// The CR is being deleted and has finalizer. This stage allows you to manage your own cleanup logic. The CR update is managed at the OKT reconciler level.
//		ErrorManager		// Is reached after any step that raise an error different than GiveUpError. Just add an error in your Reconciler's results.
//
//  Here is a special debranching step that is managed internaly by the Stepper, no need to handle it at your end
//      GiveUpManager		// A Terminal state. Can be reached at any time/in any state (even in SuccessManager) when an unrecoverable error happen or simply when the reconciliation can't go further for the moment
//
// Besides that, use the "Free style" engine if you want a full control on the reconiliation process.
type Stepper struct {
	logr.Logger
	okterr.Results

	machine *oktsm.Machine

	hook StepperEngineHook
}

// blank assignment to verify that ReconcileCockroachDB implements reconcile.Reconciler
var _ oktreconciler.Engine = &Stepper{}
var _ oktsm.LCGStateAction = &Stepper{}

// GetState transition.Stater implementation
func (smc *Stepper) GetState() string {
	return recGraph[smc.machine.GetState()].Name
}

func (smc *Stepper) Enter(state oktsm.LCGState) error {
	//state := smc.machine.GetState()

	switch state {
	case GiveupManager:
		giveUpStateHook(smc)
	case End:
		endStateHook(smc)
	default:
		smc.hook.EnterInState(smc)
	}

	return nil
}

const (
	//Initial oktsm.LCGState
	CRChecker = iota // The entry point of a reconciliation
	CRFinalizer
	ObjectsGetter
	Mutator
	Updater
	ErrorManager
	SuccessManager
	GiveupManager // This special state to deal with a not alarming error (i.e. no CR available,...) or with an alarming error (implementation problem,...)
	End           // Leaf node: End of reconciliation
)

/**
*** The Life Cycle Graph for the Reconciliation state machine
**/
var recGraph = oktsm.LCGGraph{
	CRChecker: oktsm.LCGNodeInfo{
		Name: "CRChecker",
		Children: oktsm.LCGChildren{
			ErrorManager,
			GiveupManager,
			CRFinalizer,
			ObjectsGetter,
		},
	},
	ObjectsGetter: oktsm.LCGNodeInfo{
		Name: "ObjectsGetter",
		Children: oktsm.LCGChildren{
			ErrorManager,
			GiveupManager,
			Mutator,
		},
	},
	Mutator: oktsm.LCGNodeInfo{
		Name: "Mutator",
		Children: oktsm.LCGChildren{
			ErrorManager,
			GiveupManager,
			Updater,
		},
	},
	Updater: oktsm.LCGNodeInfo{
		Name: "Updater",
		Children: oktsm.LCGChildren{
			ErrorManager,
			GiveupManager,
			SuccessManager,
		},
	},
	CRFinalizer: oktsm.LCGNodeInfo{
		Name: "CRFinalizer",
		Children: oktsm.LCGChildren{
			ErrorManager,
			GiveupManager,
			End, // This is the end of life!!
		},
	},
	SuccessManager: oktsm.LCGNodeInfo{
		Name: "SuccessManager",
		Children: oktsm.LCGChildren{
			ErrorManager,
			GiveupManager,
			End, // A real success!!
		},
	},
	ErrorManager: oktsm.LCGNodeInfo{
		Name: "ErrorManager",
		Children: oktsm.LCGChildren{
			GiveupManager,
			End, // End as a simple error case
		},
	},
	GiveupManager: oktsm.LCGNodeInfo{
		Name: "GiveupManager",
		Children: oktsm.LCGChildren{
			End, // End a Giveup with or without an alarming error
		},
	},
	End: oktsm.LCGNodeInfo{Name: "End"}, // A leaf node
}

// Determine which is the next state after the current one and regarding the case of error or not and the Givenup case if any.
// Return the event name that will be triggered to go to the next state.
func (smc *Stepper) eventsList(curState oktsm.LCGState) (events oktsm.LCGEvents) {
	giveup, err := smc.Results.ConsolidatedError()
	finalizing := smc.Results.OpsCount(okterr.OperationResultCRIsFinalizing) == 1

	events = make(oktsm.LCGEvents, 0)

	// An error raised at any step ?
	if err != nil {
		events = append(events, ErrorManager)
	}
	// Giveup raised at any step ?
	if giveup {
		events = append(events, GiveupManager)
	}
	// CR Finalizing raised ?
	if finalizing {
		events = append(events, CRFinalizer)
	}

	// Normal course or end
	events = append(events, oktsm.DefaultState)

	return events
}

// DisplayPathOfStates log the path of states
func (smc *Stepper) DisplayPathOfStates(logs logr.Logger) {
	path := "History: " + smc.machine.GetPathInGraph()
	logs.Info(path)
}

// SetLogger xx
func (smc *Stepper) SetLogger(logs logr.Logger) {
	smc.Logger = logs
}

// SetResults xx
func (smc *Stepper) SetResults(res okterr.Results) {
	smc.Results = res
}

// Run starts from the first reconciliation step (CRChecker).
// At each an action hook is called with a context parameter which is the Stepper engine itself.
// It stops either if an error is returned or if the "End" state is reached
// All results (errors, operations, ...) have to be collected in Reconciler's results (Results)data
// Compute next state. The principle is to generate a list of events based on watched variable (error, giveup, finalizing) and
// build a list of events. To be sure to go to the next step, the normal course event "DefaultState" is always added to the list.
func (smc *Stepper) Run() {
	okterr.ErrGiveUpReconciliation.Reason(nil)

	var infiniteLoopBreaker = 1000 // Security in case of wrong machine state model

	smc.machine.EnablePathInGraph() // Enable path function or reset it to zero /!\
	smc.machine.SetState(CRChecker) // Enter into First state (without any event condition)

	// Loop on each step of the state machine and call hooks until an error is raised or an ending state is reached
	// Ending state: "End"
	for state := smc.machine.GetState(); !recGraph.IsLeafNode(state); state = smc.machine.GetState() {
		events := smc.eventsList(state)
		if entered, _ := smc.machine.EnterNextState(events); !entered {
			return
		}

		// Very special case (wrong graph definition ?)
		if infiniteLoopBreaker--; infiniteLoopBreaker == 0 {
			smc.AddGiveupError(nil, okterr.OperationResultInfiniteLoop, fmt.Errorf("infinite loop in reconciliation machine state"))
			return
		}
	}
}

// NewStepper Allocates a new reconciler Engine that will course several steps that maps a classical Reconciling process
// See Stepper type documentation
func NewStepper(hook StepperEngineHook) *Stepper {
	s := &Stepper{hook: hook}
	s.machine = &oktsm.Machine{Graph: recGraph, Actions: s}

	return s
}
