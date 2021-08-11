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

package resources

import (
	okthash "github.com/Orange-OpenSource/Operators-Karma-Tools/tools/hash"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Params brings resources parameters to use for reconciliation
type Params interface {
	SetData(map[string]string)
	GetData() map[string]string
}

// ResourceInfo interface provides resource ID and name
type ResourceInfo interface {
	// A uniq key name
	Index() string
	// Commodity name composed with resource "Kind/Name"
	KindName() string
}

// Resource a resource manageable by the OKT Reconciler. It is synchronized with its peer on cluster and
// offers utilities for its creation.
// A simple Resource can NOT be mutated (see MutableResource), but only created.
type Resource interface {
	ResourceInfo

	IsCreation() bool
	CreatePeer() error
	SyncFromPeer() error
	Params
}

// MutableResource provides all the required tools to process a mutation on a resource having a Mutator (mandatory)
// All things driven with idempotency in mind.
type MutableResource interface {
	// Tell if yes or not this resource will be mutated with its owner reference. Default is True.
	//WithOwnerReference(bool)

	// UpdateSyncStatus Apply an annotation fingerprint to the object which permits to follows object modifications.
	// Compute new object's fingerprint (Hash) and compare it with the existing fingerprint.
	// Update status telling if yes or no the expected value is Synched with the Cluster Resource
	// When it is an object creation, no existing hash value exists, thus the needResync property will be TRUE!
	UpdateSyncStatus(ref okthash.HashableRef) error

	// Is this object (before, after or without mutation) out of sync with its peer ?
	// The NeedResync status is updated as well thanks to a call to UpdateSyncStatus()
	// As soon it is triggered to True it remains at True
	NeedResync() bool
	// If NeedResync() is already true, LastSyncState() allow to know if last call to UpdateSyncStatus has detected NO modification.
	LastSyncState() bool

	PreMutate(scheme *runtime.Scheme) error
	PostMutate(cr client.Object, scheme *runtime.Scheme) error

	// Verify  that object keys (index) remains the same as expected, even after a modifications
	CheckExpectedKey() error

	UpdatePeer() error
}

// Mutator is an interface providing mutation function based on hash computation to determines changes on the resource
// The HashableRef provides all the objects entering in the scope of the Hash computation
type Mutator interface {
	// HashableRef is a reference (address) on a structure we want to use for
	// the Hash calculation performed to detect object's modification
	GetHashableRef() okthash.HashableRef

	MutateWithInitialData() error                          // Initial data are the one set at creation time
	MutateWithCR() (requeueAfterSeconds uint16, err error) // Custom Resource data can evolve during resource life cycle
}

// MutableResourceType is an OKT resource leading a resource type that can be mutated by an OKT Reconciler
type MutableResourceType interface {
	Resource
	MutableResource
	Mutator
}

// MutationHelper provides  pre and post mutation help in order to deal with specific resources beahaviour like Secrets
// When a mutation helper has been defined for a MutableResourceType, it can then be added to the Pre and PostMutate() functions called
// by the OKT Reconciler.
type MutationHelper interface {
	GetObject() client.Object
	GetObjectSpec() interface{}
	PreMutate() error
	PostMutate() error
}
