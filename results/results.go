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
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	oktres "gitlab.tech.orange/dbmsprivate/operators/okt/resources"
)

// OperationResult is the action result of a CreateOrUpdate call
//type OperationResult controllerutil.OperationResult
type OperationResult string

const ( // They should complete the sentence "Deployment default/foo has been ..."
	///// APPLICATION // OBSOLETE
	/*
		// OperationResultAppMutationError xx
		OperationResultAppMutationError OperationResult = "application resource mutatation on error"
		// OperationResultAppMutationSuccess xx
		OperationResultAppMutationSuccess OperationResult = "success of application resource mutatation"
		// OperationResultAppUpdateError xx
		OperationResultAppUpdateError OperationResult = "application update on error"
		// OperationResultAppUpdated xx
		OperationResultAppUpdated OperationResult = "application updated"
		// OperationResultAppUnchanged xx
		OperationResultAppUnchanged OperationResult = "application unchanged"
	*/
	///// CR

	// OperationResultCRNotExist means that we can not pickup the CR on the Cluster
	OperationResultCRNotExist OperationResult = "CR not found"
	// OperationResultCRTypeInvalid means that the CR object type is invalid
	OperationResultCRTypeInvalid OperationResult = "CR type invalid"
	// OperationResultCRUnreadable means the CR exists but can't be read
	OperationResultCRUnreadable OperationResult = "CR is not readable"
	// OperationResultFetchCRSuccess means the CR exists and sucessfully picked on cluster
	OperationResultFetchCRSuccess OperationResult = "CR is succesfully picked up on Cluster"
	// OperationResultCRSemanticError means the CR is semanticaly wrong
	OperationResultCRSemanticError OperationResult = "CR is sematicaly wrong"
	// OperationResultCRIsFinalizing CR is being deleted and finalizing (has finalizer)
	OperationResultCRIsFinalizing OperationResult = "CR is being deleted and finalizing (has finalizer)"
	// OperationResultCRFinalizationUpdateError the end of CR finalization failed on update
	OperationResultCRFinalizationUpdateError OperationResult = "the end of CR finalization failed on update"

	///// REGISTRATION for resources

	// OperationResultRegistrationAborted means that a resource registration in OKT has aborted
	OperationResultRegistrationAborted OperationResult = "reconciler registration aborted on error"
	// OperationResultResourceUnreadable means that we can not pickup the resource on the Cluster
	OperationResultResourceUnreadable OperationResult = "unreadable resource"
	// OperationResultRegistrationSuccess means that we all the resources are registered in the OKT registry
	OperationResultRegistrationSuccess OperationResult = "resource registration success"

	///// MUTATION for resources

	// OperationResultMutationSuccess means that a resource mutation has aborted
	OperationResultMutationSuccess OperationResult = "resource mutation success"
	// OperationResultMutateWithCRError means that there's a pb to mutate the resource with the CR values
	OperationResultMutateWithCRError OperationResult = "resource.MutateWithCR() on error"
	// OperationResultMutateWithCRAskRequeue means that the resource with the CR values is done and a requeue is requested
	OperationResultMutateWithCRAskRequeue OperationResult = "resource.MutateWithCR() done with a requeing result"

	///// Cluster CRUD for resources

	// OperationResultNone means that the resource has not been changed
	OperationResultNone OperationResult = "resource unchanged"
	// OperationResultCreated means that a new resource is created
	OperationResultCreated OperationResult = "resource created"
	// OperationResultCreateDelayed means that a new resource creation is delayed (not an error)
	OperationResultCreateDelayed OperationResult = "resources creation delayed"
	// OperationResultUpdated means that an existing resource is updated
	OperationResultUpdated OperationResult = "resource updated"
	// OperationResultDeleted means that an existing resource is deleted
	OperationResultDeleted OperationResult = "resource deleted"
	// OperationResultCRUDError means that a Create Update or Delete has failed
	OperationResultCRUDError OperationResult = "crud error"

	///// MISC
	///// Alarming errors that should raise a Giveup error

	// OperationResultImplementationConcern xx
	OperationResultImplementationConcern OperationResult = "an implementation concern raised an error"
	// OperationResultInfiniteLoop Error on infinite loop detection
	OperationResultInfiniteLoop OperationResult = "infinite loop"

	///// Status

	// OperationResultStatusEnabled means that an OKT status has been detected in CR and enabled to be updated
	OperationResultStatusEnabled OperationResult = "okt status enabled"

	// OperationResultStatusUpdated means that the CR status is updated
	OperationResultStatusUpdated OperationResult = "status updated"

	// OperationResultStatusUpdateError means that the CR status is updated
	OperationResultStatusUpdateError OperationResult = "status update error"

	// OperationResultSameStatusError appears when we detect a same error raised at each reconciliation cycle. Implies that OKT Status is enabled in CR
	OperationResultSameStatusError OperationResult = "same error at each reconciliation cycle"
)

// Stats report several counter on operation results (error, operations) and a display method
type Stats interface {
	// OpsCount return the current count of operations for a specified ResultOperation type
	OpsCount(operation OperationResult) uint16
	// TotalOpsCount return the total count of operations whatever the ResultOperation type
	TotalOpsCount() (opsTypeCount, opsCount uint16)
	ErrorsCount() uint16
	DisplayCounters(logger logr.Logger)
}

// Results manage a list of results for a reconciler that cumulates some operation's results during execution
type Results interface {
	AddOp(resource oktres.ResourceInfo, result OperationResult, err error, requeueAfterSeconds uint16) error
	AddOpSuccess(resource oktres.ResourceInfo, result OperationResult)
	AddGiveupError(resource oktres.ResourceInfo, result OperationResult, err error) error

	DisplayOpList(logger logr.Logger)

	// ConsolidatedError Return consolidated error
	// Unlike ConsolidatedSigsK8S(), this method returns the ErrGiveUpReconciliation status (raised or not) and the current error of the AlarmingReason error
	ConsolidatedError() (giveup bool, err error)

	// ConsolidatedSigsK8S returns:
	//  nil if no error
	//  It is possible to requeue with a delay (in nanoseconds). The delay returned here is computed regarding the delays reported in the Results.
	//  In case of a same status error reported in the Results, the delay is increased of a period of time growing exponentially at each cycle (up to 6 hours)
	//  In a reconciler (as expected by sigs.k8s.io), 3 outputs are possible:
	//  - Return no error and don't requeue, same as: reconcile.Result{}, nil
	//  - Return no error and requeue with a specified delay, same as: reconcile.Result{Requeue: true, RequeueAfter: delay}, nil
	//  - Return an error and requeue, same as: reconcile.Result{}, err
	ConsolidatedSigsK8S() (reconcile.Result, error)

	// Some counters on the reconciliation process
	Stats

	ResetAllResults()
}
