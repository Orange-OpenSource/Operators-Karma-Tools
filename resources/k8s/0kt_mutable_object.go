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

package k8s

import (
	"errors"

	okthash "github.com/Orange-OpenSource/Operators-Karma-Tools/tools/hash"
	//"k8s.io/kubernetes/pkg/apis/apps
)

// MutableResourceObject implements an ObjectResource with mutation facilities
type MutableResourceObject struct {
	ResourceObject // OKT K8S resource

	needResync    bool // false by default
	lastSyncState bool // false by default
}

// Blank assignement to check type
//var _ oktres.MutableResource = &MutableResourceObject{}

// UpdateSyncStatus Apply an annotation fingerprint to the object which permits to follows object modifications.
// Update status telling if yes or no the expected value is Synched with the Cluster Resource
// When it is an object creation, no existing hash value exists, thus the needResync property will be TRUE!
func (r *MutableResourceObject) UpdateSyncStatus(ref okthash.HashableRef) error {
	var err error

	//obj := r.GetExpected()
	obj := r.Object

	if r.lastSyncState, err = okthash.GenerateNew(obj, ref.GetRef()); err != nil {
		return err
	}

	// Update Synch status can be called twice, so don't set to false if it is already true!
	if r.lastSyncState {
		r.needResync = true
	}

	return nil
}

// NeedResync Tells if the expected object has been modified regarding its cluster version.
func (r *MutableResourceObject) NeedResync() bool {
	return r.needResync
}

// LastSyncState Tells if call to UpdateSyncStatus() detected a modification on the resource regarding its cluster version
// Can differ from NeedResync as this last stays TRUE for ever once triggered.
func (r *MutableResourceObject) LastSyncState() bool {
	return r.lastSyncState
}

// UpdatePeer Update peer object. Can NOT succeed if the resource has been marked as "to be created"
// at OKT's registration time.
// Reset NeedResync flag to false in case of success.
func (r *MutableResourceObject) UpdatePeer() error {
	if r.IsCreation() {
		return errors.New("Peer is presumed not yet created:" + r.Index())
	}

	if err := r.Update(); err != nil {
		return err
	}
	r.needResync = false

	return nil
}
