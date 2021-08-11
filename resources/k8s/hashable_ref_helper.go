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
	"fmt"
	"sort"

	oktres "github.com/Orange-OpenSource/Operators-Karma-Tools/resources"
	okterr "github.com/Orange-OpenSource/Operators-Karma-Tools/results"
	okthash "github.com/Orange-OpenSource/Operators-Karma-Tools/tools/hash"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// HashableRefHelper xx
type HashableRefHelper struct {
	client.Object
	ref []interface{}
	oktres.MutationHelper
}

// blank assignment to verify that HashableRefHelper implements HashableRef
var _ okthash.HashableRef = &HashableRefHelper{}

// GetRef returns the hashable reference (interface{}) used in Hash computation and comparaison
// to detect an object's modification during a reconciliation
func (hr *HashableRefHelper) GetRef() interface{} {
	return hr.ref
}

func (hr *HashableRefHelper) add(ref interface{}) {
	hr.ref = append(hr.ref, ref)
}

func (hr *HashableRefHelper) addSortedStringMap(ref map[string]string) {
	//hr.ref = append(hr.ref, ref)
	keys := make([]string, 0, len(ref))
	for k := range ref {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		hr.add(k)
		hr.add(ref[k])
	}
}

// AddWholeObject The whole object will be compared. Not recommended, this option is prone to discrepancies due to some fields managed by the Cluster
func (hr *HashableRefHelper) AddWholeObject() {
	hr.add(hr.Object)
}

/*
// AddAllMeta The whole object will be compared. Not recommended, this option is prone to discrepancies due to some fields managed by the Cluster
func (hr *HashableRefHelper) AddAllMeta() {
	hr.add(hr.Object)
}
*/

// AddMetaMainFields Shortcut to add main Meta fields to the reference for hash computation (Labels, Annotations, ownerReferences).
func (hr *HashableRefHelper) AddMetaMainFields() {
	hr.AddMetaLabels()
	hr.AddMetaAnnotations()
	hr.add(hr.GetOwnerReferences())
}

// AddMetaLabels xx
func (hr *HashableRefHelper) AddMetaLabels() {
	//hr.addSortedStringMap(hr.meta.GetLabels())
	hr.add(hr.GetLabels())
}

// AddMetaLabelValues xx
func (hr *HashableRefHelper) AddMetaLabelValues(keys ...string) {
	themap := hr.GetLabels()
	for _, key := range keys {
		hr.add(key)
		hr.add(themap[key])
	}
}

// AddMetaAnnotations Add all annotations (except the hash itself!!!) to the reference for hash computation
func (hr *HashableRefHelper) AddMetaAnnotations() {
	annotations := hr.GetAnnotations()

	// Backup existing has (if any)
	hash, exist := annotations[okthash.OKTHashAnnotationName]

	if exist {
		delete(annotations, okthash.OKTHashAnnotationName)
	}

	// Add
	//hr.addSortedStringMap(annotations)
	hr.add(annotations)

	// Restore hash annotation ?
	if exist {
		annotations[okthash.OKTHashAnnotationName] = hash
	}
}

// AddMetaAnnotationsValues xx
func (hr *HashableRefHelper) AddMetaAnnotationsValues(keys ...string) {
	themap := hr.GetAnnotations()
	for _, key := range keys {
		hr.add(key)
		hr.add(themap[key])
	}
}

// AddUserData xx
func (hr *HashableRefHelper) AddUserData(ref interface{}) {
	hr.add(ref)
}

// AddSpec If the MutationHelper defines what is the Spec of the resource Object, this method adds it to the Hashable Ref.
// If the MutationHelper do not defines a Spec object, this method does nothing but return an error (of type implementation)
func (hr *HashableRefHelper) AddSpec() error {
	if spec := hr.MutationHelper.GetObjectSpec(); spec != nil {
		hr.AddUserData(spec)
		return nil
	}
	return fmt.Errorf("%s", okterr.OperationResultImplementationConcern)
}

// Init Initialize a HashableRef helper
func (hr *HashableRefHelper) Init(mutationHelper oktres.MutationHelper) {
	hr.ref = make([]interface{}, 0)
	hr.Object = mutationHelper.GetObject()
	hr.MutationHelper = mutationHelper
}
