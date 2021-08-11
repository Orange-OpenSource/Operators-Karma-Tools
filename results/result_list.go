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
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	oktres "github.com/Orange-OpenSource/Operators-Karma-Tools/resources"
)

// opResInfo xx
type opResInfo struct {
	operation OperationResult
	error
	resource            oktres.ResourceInfo
	requeue             bool
	requeueAfterSeconds uint16
}

// opStats Cumulated indicators on operations
type opStats struct {
	opMap         map[OperationResult]uint16
	errorsCount   uint16
	totalOpsCount uint16
}

// Blank assignement to chect type
var _ Stats = &opStats{}

// ErrorsCount Get errors count
func (s *opStats) ErrorsCount() uint16 {
	return s.errorsCount
}

// OpsCount return the current count of operations for a specified ResultOperation type
func (s *opStats) OpsCount(operation OperationResult) uint16 {
	return s.opMap[operation]
}

// TotalOpsCount return the total count of operations whatever the ResultOperation type
func (s *opStats) TotalOpsCount() (opsTypeCount, opsCount uint16) {
	opsTypeCount = uint16(len(s.opMap))
	return opsTypeCount, s.totalOpsCount
}

// DisplayCounters Display Stats counters
func (s *opStats) DisplayCounters(logger logr.Logger) {
	var msg string
	for op, count := range s.opMap {
		msg = fmt.Sprintf("%d %s", count, op)
		logger.Info(msg)
	}
	msg = fmt.Sprintf("%d Ops(s)", s.totalOpsCount)
	logger.Info(msg)
	msg = fmt.Sprintf("%d Error(s)", s.errorsCount)
	logger.Info(msg)
}

func (s *opStats) resetCounters() {
	s.opMap = make(map[OperationResult]uint16, 0)
	s.errorsCount = 0
	s.totalOpsCount = 0
}

func (s *opStats) addToCounters(err error, operation OperationResult) {
	s.opMap[operation]++
	s.totalOpsCount++
	if err != nil {
		s.errorsCount++
	}
}

// resultList a stack of results
type resultList struct {
	giveup        bool
	opResInfoList []*opResInfo
	consolidated  opResInfo
	opStats
}

// blank assignment to verify that ResultList implements Result
var _ Results = &resultList{}

// setRequeue Set result with or without error and with or without requeueing.
// Note that in case of error, requeueing if forced (0 seconds by default), except for the GivUp error.
// ErrGiveUpReconciliation means we'll try to abort the reconciliation processs. No requeing is possible on such "error"
// Max duration is 6 hours so 21600 seconds
func (r *opResInfo) setRequeue(requeueAfterSeconds uint16) {
	// Treat special case of GiveUp error and return
	if r.error == ErrGiveUpReconciliation {
		r.requeue = false
		r.requeueAfterSeconds = 0
		return
	}

	if requeueAfterSeconds > 21600 {
		requeueAfterSeconds = 21600
	}

	r.requeueAfterSeconds = requeueAfterSeconds

	// As soon as it is an error, a requeue is requested.
	// A requeue is requested too, in case of NO error and the requeue time is not 0 => No immediate requeue without error!
	if r.error != nil || r.requeueAfterSeconds > 0 {
		r.requeue = true
	}
}

// getSigsK8SResult Return Result in it sigs.k8s.io reconcile form
func (r opResInfo) getSigsK8SResult() (reconcile.Result, error) {
	if r.error == ErrGiveUpReconciliation {
		return reconcile.Result{Requeue: false, RequeueAfter: 0}, nil // Do not return the reason/alarming error, to avoid a requeue request.
	}
	return reconcile.Result{Requeue: r.requeue, RequeueAfter: time.Duration(r.requeueAfterSeconds) * time.Second}, r.error
}

// addEntry add a new result and build (as we go) the consolidated result as well
// Return (pass) the entry's error
func (r *resultList) addEntry(entry *opResInfo) error {
	r.opResInfoList = append(r.opResInfoList, entry)

	// Compute on-the-go, consolidated result
	// Track first error only. GiveUpReconciliation is prioritary so it is never overrided
	if r.consolidated.error == nil || entry.error == ErrGiveUpReconciliation {
		r.consolidated.error = entry.error
		r.consolidated.resource = entry.resource
	}

	// Set Requeue state, generally Keep minimal requeue duration only but on some errors (GiveUp), the requeue state will be false
	// If the current operation is a notification that it is a permanent error, we take the requeue duration provided
	// (probably growing up after each occurence)
	if entry.requeue {
		if !r.consolidated.requeue ||
			entry.requeueAfterSeconds < r.consolidated.requeueAfterSeconds ||
			entry.operation == OperationResultSameStatusError {
			r.consolidated.setRequeue(entry.requeueAfterSeconds)
		}
	}

	// Update stats
	r.addToCounters(entry.error, entry.operation)

	return entry.error
}

// AddOp Add a result operation to the maintained list of results and maintain a consolidated state
// The consolidated's requeue state can be reset to False on some errors (GiveUp)
// Return (pass) the added result's error passed as parameter
func (r *resultList) AddOp(resource oktres.ResourceInfo, result OperationResult, err error, requeueAfterSeconds uint16) error {
	if err == ErrGiveUpReconciliation {
		r.giveup = true
	}
	entry := opResInfo{
		resource:  resource,
		operation: result,
		error:     err,
	}
	entry.setRequeue(requeueAfterSeconds)

	return r.addEntry(&entry)
}

// AddOpSuccess Add a result operation in case of success (without requeueing!). Use Add if you need to requeue on success.
// Return (pass) the added result's error
func (r *resultList) AddOpSuccess(resource oktres.ResourceInfo, result OperationResult) {
	r.AddOp(resource, result, nil, 0)
}

// AddGiveupError Add a GiveUp Reconciliation result to the maintained list of results and maintain a consolidated state
// Return (pass) the added result's error
func (r *resultList) AddGiveupError(resource oktres.ResourceInfo, result OperationResult, alarmingReason error) error {
	ErrGiveUpReconciliation.Reason(alarmingReason)
	return r.AddOp(resource, result, ErrGiveUpReconciliation, 0)
}

// ResetOpList Delete all elements and re-init the List to 0 element
func (r *resultList) ResetAllResults() {
	r.consolidated.resource = nil
	r.consolidated.operation = OperationResultNone
	r.consolidated.error = nil
	r.consolidated.requeue = false
	r.consolidated.setRequeue(0)

	r.opResInfoList = nil
	r.opResInfoList = make([]*opResInfo, 0)
	r.resetCounters()
}

// ConsolidatedError Return consolidated error
// Unlike ConsolidatedSigsK8S(), this method returns the ErrGiveUpReconciliation status (raised or not) and the current error of the AlarmingReason error
func (r *resultList) ConsolidatedError() (giveup bool, err error) {
	err = r.consolidated.error
	if err == ErrGiveUpReconciliation {
		giveup = true
		err = ErrGiveUpReconciliation.AlarmingReasonToGiveup()
	}
	return giveup, err
}

// ConsolidatedSigsK8S Return consolidated Result in its sigs.k8s.io reconcile version and the error
//  In a reconciler (as expected by sigs.k8s.io), 3 outputs are possible:
//  - Return no error and don't requeue: reconcile.Result{}, nil
//  - Return no error and requeue: reconcile.Result{Requeue: true}, nil
//  - Return an error and requeue: reconcile.Result{}, err
func (r *resultList) ConsolidatedSigsK8S() (reconcile.Result, error) {
	return r.consolidated.getSigsK8SResult()
}

// DisplayOpList Write results list to logger
func (r *resultList) DisplayOpList(logger logr.Logger) {
	for _, entry := range r.opResInfoList {
		resName := "undef"
		if entry.resource != nil {
			resName = entry.resource.KindName()
		}
		if entry.error == nil {
			logger.WithValues("res", resName).Info("Op: " + string(entry.operation))
			continue
		}
		logger.WithValues("res", resName).Error(entry.error, "Op: "+string(entry.operation))
	}
	logger.Info("Consolidated requeue duration: " + fmt.Sprint(r.consolidated.requeueAfterSeconds) + " seconds")
}

// NewResultList Allocates a new registry
func NewResultList() Results {
	list := &resultList{}
	list.ResetAllResults()

	return list
}
