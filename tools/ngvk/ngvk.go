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

package ngvk

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Compose Build a string which can serves as key to identify aan object
func compose(gvk schema.GroupVersionKind, key types.NamespacedName) string {
	return key.Namespace + " " + gvk.Kind + "/" + key.Name + " " + gvk.Group + "/" + gvk.Version
}

// NGVK is a Key object type that represent a uniq object key based on GroupVersion Kind and Name
type NGVK struct {
	key types.NamespacedName
	gvk schema.GroupVersionKind
}

// String builds an key string based on GroupVersion Kind and Name (NGVK)
func (ngvk NGVK) String() string {
	return compose(ngvk.gvk, ngvk.key)
}

// KN return the string composition of Kind/Name for the resource
func (ngvk NGVK) KN() string {
	return ngvk.gvk.Kind + "/" + ngvk.key.Name
}

// NamespacedName Returns the Name part ("N") of the NGVK type
func (ngvk NGVK) NamespacedName() types.NamespacedName {
	return ngvk.key
}

func (ngvk *NGVK) set(obj client.Object) error {
	//var err error

	ngvk.gvk = obj.GetObjectKind().GroupVersionKind() // GVK
	ngvk.key = client.ObjectKeyFromObject(obj)        // NamespacedName

	return nil
}

// Equal compares to NGVK types
func (ngvk NGVK) Equal(cmp NGVK) bool {
	if cmp.String() == ngvk.String() {
		return true
	}
	return false
}

// IsFor Tells if the object has the same NGVK
func (ngvk NGVK) IsFor(obj client.Object) bool {
	cmp := NGVK{}

	if err := cmp.set(obj); err != nil {
		return false
	}

	return ngvk.Equal(cmp)
}

// New Creates a new NGVK type and return its pointer
func New(obj client.Object) (*NGVK, error) {
	ngvk := &NGVK{}

	err := ngvk.set(obj)

	return ngvk, err
}
