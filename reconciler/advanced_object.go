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

package reconciler

import (
	oktres "github.com/Orange-OpenSource/Operators-Karma-Tools/resources"
	okterr "github.com/Orange-OpenSource/Operators-Karma-Tools/results"
)

// AdvancedObject Implementation of the Advanced Reconciler Object based on the Reconciler
// This reconciler deals with mutable resources.
type AdvancedObject struct {
	BasicObject
}

// Blank assignement to check type
var _ Advanced = &AdvancedObject{}

/*
// SetParams Defines which params to use to initialize registered resources
func (r *AdvancedObject) SetParams(params map[string]string) {
	r.params =
}
*/

// Mutate Mutates a mutable resource by applying defaults and values brougth by the Custom Resource (CR) of the Operator
// After potential mutation or modification, compute new object's fingerprint (Hash) and
// compare it with the existing fingerprint (hash code).
// Store the new fingerprint in the expected object.
// The entry's synch status property (NeedResync) is updated as well.
// This function adds operation's result in Results list
// Returns the error if any.
func (ar *AdvancedObject) Mutate(entry oktres.MutableResourceType) error {
	if !entry.IsCreation() {
		hashableRef := entry.GetHashableRef()
		if err := entry.UpdateSyncStatus(hashableRef); err != nil {
			return ar.AddGiveupError(entry, okterr.OperationResultImplementationConcern, err)
		}
	}

	if err := entry.PreMutate(ar.GetScheme()); err != nil {
		return ar.AddGiveupError(entry, okterr.OperationResultImplementationConcern, err)
	}

	// Decides if yes or no the Intial Data have to be re-applied to the Expected object (in memory)
	// Checks first if the Peer object has been modified against its preceeding version on cluster
	// However, in case of Creation, initial data are mandatory
	if entry.NeedResync() || entry.IsCreation() {
		if err := entry.MutateWithInitialData(); err != nil {
			return ar.AddGiveupError(entry, okterr.OperationResultImplementationConcern, err)
		}
	}

	{ // Treat Mutate with CR values and possible requeuing options
		var err error
		var requeueAfterSeconds uint16
		// Then Mutate expected object with CR values (as defined, or not, in the OKT resource's implementation)
		if requeueAfterSeconds, err = entry.MutateWithCR(); err != nil {
			return ar.AddOp(entry, okterr.OperationResultMutateWithCRError, err, requeueAfterSeconds)
		}
		if requeueAfterSeconds > 0 {
			ar.AddOp(entry, okterr.OperationResultMutateWithCRAskRequeue, nil, requeueAfterSeconds)
		}
	}

	if err := entry.PostMutate(ar.GetCR(), ar.GetScheme()); err != nil {
		return ar.AddGiveupError(entry, okterr.OperationResultImplementationConcern, err)
	}

	// Safe measure on keys (NGVK) data. After mutation(s) forbid any modification of these ID properties!
	if err := entry.CheckExpectedKey(); err != nil {
		return ar.AddGiveupError(entry, okterr.OperationResultImplementationConcern, err)
	}

	// Compute Hash  and detect a modification on the object picked up from the cluster
	hashableRef := entry.GetHashableRef()
	if err := entry.UpdateSyncStatus(hashableRef); err != nil {
		return ar.AddGiveupError(entry, okterr.OperationResultImplementationConcern, err)
	}

	ar.AddOpSuccess(entry, okterr.OperationResultMutationSuccess)
	return nil
}

// MutateAllResources Mutate all Mutable resources
// Return immediatley if a raised or current consolidated error is GiveUpReconciliation
// If stopOnError is true, stop as soon as an error is raised
// Return the last raised error during a resource Mutation
func (ar *AdvancedObject) MutateAllResources(stopOnError bool) error {
	if giveup, err := ar.ConsolidatedError(); giveup {
		return err
	}

	var err error

	for _, res := range ar.GetRegisteredResources() {
		mutRes, ok := res.(oktres.MutableResourceType)
		if ok {
			if err = ar.Mutate(mutRes); err != nil {
				if stopOnError {
					return err
				}
			}
		}
	}

	return err
}

// Update update a mutable resource (so, having a Mutator interface) and modified against its Cluster Peer (NeerResync() = true).
// The resource passed as arguement must be a Mutable resource
// This function adds operation's result in reconciler's Results list
// Returns the error if any.
func (ar *AdvancedObject) Update(resource oktres.MutableResourceType) error {
	// Is a mutable and unsynched resource with its cluster version ? => Do Uppdate
	if resource.NeedResync() {
		if err := resource.UpdatePeer(); err != nil {
			return ar.AddOp(resource, okterr.OperationResultCRUDError, err, requeueDurationOnCRUDError)
		}

		ar.AddOpSuccess(resource, okterr.OperationResultUpdated)
		return nil
	}

	ar.AddOpSuccess(resource, okterr.OperationResultNone)
	return nil
}

// CreateOrUpdateAllResources Utility method to Create or Update all OKT resources (taking care of their types and created/mutation status)
// The parameter maxCreation specified the maximum count of resources to create in one shot
// Return immediatley if a raised or current consolidated error is GiveUpReconciliation
// If stopOnError is true, stop as soon as an error is raised
// Return the last raised error during the updates
func (ar *AdvancedObject) CreateOrUpdateAllResources(maxCreation uint16, stopOnError bool) error {
	if giveup, err := ar.ConsolidatedError(); giveup {
		return err
	}

	var err error

	for _, res := range ar.GetRegisteredResources() {
		if res.IsCreation() {
			if err = ar.Create(res, maxCreation); err != nil {
				if stopOnError {
					return err
				}
			}
			continue
		}
		mutRes, ok := res.(oktres.MutableResourceType)
		if ok {
			if err = ar.Update(mutRes); err != nil {
				if stopOnError {
					return err
				}
			}
		}
	}

	return err
}
