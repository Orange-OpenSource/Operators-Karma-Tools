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

package engines

import (
	"errors"
	"fmt"
	"testing"

	//"k8s.io/apimachinery/pkg/runtime"
	//	"github.com/stretchr/testify/assert"
	//	corev1 "k8s.io/api/core/v1"
	//	meta "k8s.io/apimachinery/pkg/api/meta"
	//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/go-logr/zapr"
	"github.com/stretchr/testify/require"
	oktreconciler "gitlab.tech.orange/dbmsprivate/operators/okt/reconciler"
	okterr "gitlab.tech.orange/dbmsprivate/operators/okt/results"
	"go.uber.org/zap"
)

type testReconciler struct {
	oktreconciler.AdvancedObject
}

// Blank assignement to check type
var _ StepperEngineHook = &testReconciler{}

//var _ StateHooks = &testReconciler{} // Deprecated

var errMyAlarmingError = errors.New("my alarming error")

func (rec *testReconciler) EnterInState(engine *Stepper) {
	//var err error
	fmt.Println("Enter in ?" + engine.GetState())

	switch engine.GetState() {
	// Main course states
	case "CRChecker":
	case "ObjectsGetter":
	case "Mutator":
		rec.AddOpSuccess(nil, okterr.OperationResultMutationSuccess)
		rec.AddOpSuccess(nil, okterr.OperationResultMutationSuccess)
	case "Updater":
		rec.AddGiveupError(nil, okterr.OperationResultImplementationConcern, errMyAlarmingError) // An alarming error! We'll manage error (in ErrorManager) then giveup (in GiveupManager without action)!
	case "SuccessManager":
		rec.AddOpSuccess(nil, okterr.OperationResultMutationSuccess)
	// Debranching states
	case "CRFinalizer":
	case "ErrorManager":
		myErr := errors.New("My dummy error")
		rec.AddOp(nil, okterr.OperationResultMutateWithCRError, myErr, 99) // Add a second error
	default:
	}
	//_, err = rec.Results.ConsolidatedError()
}

func checkCounts(t *testing.T, reconciler oktreconciler.Basic) {
	require.Equal(t, uint16(2), reconciler.ErrorsCount(), "2 error raised (a Giveup in Update and 1 error in ErrorManager)")
	require.Equal(t, uint16(1), reconciler.OpsCount(okterr.OperationResultImplementationConcern), "A Giveup error raised")
	require.NotEqual(t, uint16(3), reconciler.OpsCount(okterr.OperationResultMutationSuccess), "Should not passed in SuccessManager")
	require.Equal(t, uint16(2), reconciler.OpsCount(okterr.OperationResultMutationSuccess), "2 Mutations success occured")
	require.Equal(t, uint16(1), reconciler.OpsCount(okterr.OperationResultMutateWithCRError), "Should not passed in ErrorManager")

	giveup, err := reconciler.ConsolidatedError()
	require.True(t, giveup, "We gave up!")
	require.Equal(t, errMyAlarmingError, err, "The consolidated error should be the Alarming error that drove us to giveup")
}

func TestResultsNew(t *testing.T) {
	/*
		var (
			buf       bytes.Buffer
			legLogger = log.New(&buf, "logger: ", log.Lshortfile)
		)

		logger := &myLogger{legLogger: legLogger}
	*/
	zapLog, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("who watches the watchmen (%v)?", err))
	}
	logger := zapr.NewLogger(zapLog)

	reconcilerWithHooks := &testReconciler{}
	reconcilerWithHooks.Log = logger
	reconcilerWithHooks.Init("prod", nil, nil)

	// Stepper V2
	engine := NewStepper(reconcilerWithHooks)
	reconcilerWithHooks.SetEngine(engine)

	reconcilerWithHooks.Results.ResetAllResults()
	engine.Run()
	checkCounts(t, reconcilerWithHooks)
	//fmt.Println(engine.machine.GetPathInGraph())
	engine.DisplayPathOfStates(logger)

	reconcilerWithHooks.Results.ResetAllResults()
	engine.Run()
	checkCounts(t, reconcilerWithHooks)
}

/*
// LOGGER
type myLogger struct {
	legLogger *log.Logger
}

var _ logr.Logger = &myLogger{}

func (l *myLogger) Info(msg string, kvs ...interface{}) {
}

func (l *myLogger) Enabled() bool {
	return true
}

func (l *myLogger) Error(err error, msg string, kvs ...interface{}) {
	//l.legLogger
}

func (l *myLogger) V(_ int) logr.InfoLogger {
	return l
}

func (l *myLogger) WithName(name string) logr.Logger {
	return &myLogger{}
}

func (l *myLogger) WithValues(kvs ...interface{}) logr.Logger {
	return &myLogger{}
}
*/
