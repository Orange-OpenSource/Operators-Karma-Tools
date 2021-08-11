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

package registry

import (
	oktres "gitlab.tech.orange/dbmsprivate/operators/okt/resources"
)

// Reader provide read only access to a registry
type Reader interface {
	GetEntry(index string) oktres.Resource
	Entries() []oktres.Resource
}

// Registry manages a list of objects to reconcile
type Registry struct {
	elems []oktres.Resource
}

// blank assignment to check at compilation time this type implementation
var _ Reader = &Registry{}

func getEntry(elems []oktres.Resource, index string) oktres.Resource {
	for _, entry := range elems {
		if entry.Index() == index {
			return entry
		}
	}
	return nil
}

// GetEntry Retrieve entry from its index key
func (reg *Registry) GetEntry(index string) oktres.Resource {
	return getEntry(reg.elems, index)
}

// Entries is the slice of all registry elements
func (reg *Registry) Entries() []oktres.Resource {
	return reg.elems
}

// MutableEntries is the slice of all registry elements

// Reset Delete all registry elements and re-init the registry to 0 element
func (reg *Registry) Reset() {
	reg.elems = nil
	reg.elems = make([]oktres.Resource, 0)
}

func (reg *Registry) addEntry(entry oktres.Resource) error {
	reg.elems = append(reg.elems, entry)

	return nil
}

// AddEntry Register a resource (Mutable or Not) in a registry if it doesn't already exsist.
// Else return an error
func (reg *Registry) AddEntry(entry oktres.Resource) error {
	return reg.addEntry(entry)
}

// New Allocates a new registry
func New() *Registry {
	elems := make([]oktres.Resource, 0)

	return &Registry{elems: elems}
}
