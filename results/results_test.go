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

package results

import (
	"errors"
	"fmt"
	"testing"
	"time"

	//	"github.com/stretchr/testify/assert"
	"github.com/go-logr/zapr"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	//	corev1 "k8s.io/api/core/v1"
	//	meta "k8s.io/apimachinery/pkg/api/meta"
	//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestResultsNew(t *testing.T) {
	resInteface := NewResultList()

	results, ok := resInteface.(*resultList)
	require.True(t, ok)

	require.NotNil(t, results)
	require.Equal(t, len(results.opResInfoList), 0)
	require.Equal(t, results.consolidated.operation, OperationResultNone)
	require.Nil(t, results.consolidated.error)
	require.Equal(t, results.ErrorsCount(), uint16(0))

	// 1-Add success
	results.AddOpSuccess(nil, OperationResultMutationSuccess)
	require.Equal(t, len(results.opResInfoList), 1)
	require.Nil(t, results.consolidated.error)
	require.False(t, results.consolidated.requeue)
	require.Equal(t, results.consolidated.operation, OperationResultNone)
	require.Equal(t, results.ErrorsCount(), uint16(0))
	require.Equal(t, results.opMap[OperationResultMutationSuccess], uint16(1))

	giveup, theerr := results.ConsolidatedError()
	require.False(t, giveup)
	require.Nil(t, theerr)

	myErr := errors.New("My dummy error")

	// 2-Add Op with Error and Requeue duration and Check it on Consolidated result
	results.AddOp(nil, OperationResultMutateWithCRError, myErr, 123) // 123 seconds
	require.Equal(t, len(results.opResInfoList), 2)
	require.Equal(t, results.opMap[OperationResultMutateWithCRError], uint16(1))
	rr, e := results.ConsolidatedSigsK8S()
	require.True(t, rr.Requeue)
	require.Equal(t, rr.RequeueAfter, time.Duration(123000000000), "Conversion of seconds to microseconds is wrong or field not set")
	require.Equal(t, e, myErr, "The returned error is wrong")

	giveup, theerr = results.ConsolidatedError()
	require.False(t, giveup)
	require.NotNil(t, theerr)
	require.Equal(t, theerr, myErr)

	// 3-Add a Giveup error with an alarming reason (myErr)
	results.AddGiveupError(nil, OperationResultNone, myErr) // Give up, but there's no error there (no alarming reason)
	require.Equal(t, len(results.opResInfoList), 3)
	rr, e = results.ConsolidatedSigsK8S()
	require.False(t, rr.Requeue)
	require.Nil(t, e, "Do not return the GiveupError or any other error in case of Giveup")

	giveup, theerr = results.ConsolidatedError()
	require.True(t, giveup)
	require.NotNil(t, theerr)
	require.Equal(t, theerr, myErr)

	// Check Stats
	require.Equal(t, results.ErrorsCount(), uint16(2))
	require.Equal(t, results.opMap[OperationResultMutationSuccess], uint16(1))
	require.Equal(t, results.opMap[OperationResultMutationSuccess], uint16(1))
	require.Equal(t, results.opMap[OperationResultMutateWithCRError], uint16(1))
	require.Equal(t, results.opMap[OperationResultNone], uint16(1))
	require.Equal(t, results.opMap[OperationResultRegistrationSuccess], uint16(0))

	// 4-Add Op with Requeue duration and Check that after a giveup, the consolidated error remains giveup (thus NO error) without any requeue
	results.AddOp(nil, OperationResultRegistrationSuccess, myErr, 123) // 123 seconds
	require.Equal(t, len(results.opResInfoList), 4)
	require.Equal(t, results.opMap[OperationResultRegistrationSuccess], uint16(1))
	rr, e = results.ConsolidatedSigsK8S()
	require.False(t, rr.Requeue)
	require.Equal(t, rr.RequeueAfter, time.Duration(0))
	require.Nil(t, e, "No error set as consolidated result in case of Giveup error")

	// Reset results
	results.ResetAllResults()
	require.Equal(t, len(results.opResInfoList), 0)
	require.False(t, results.consolidated.requeue)
	require.Nil(t, results.consolidated.error)
	require.Equal(t, results.ErrorsCount(), uint16(0))
	require.Equal(t, len(results.opMap), 0)

	// Test display of
	zapLog, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("who watches the watchmen (%v)?", err))
	}
	logger := zapr.NewLogger(zapLog)

	results.DisplayCounters(logger)
	results.DisplayOpList(logger)

}
