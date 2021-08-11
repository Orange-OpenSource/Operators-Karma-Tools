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
	oktres "github.com/Orange-OpenSource/Operators-Karma-Tools/resources"
	"sigs.k8s.io/controller-runtime/pkg/client"
	//"k8s.io/client-go/kubernetes/scheme"
)

// DefaultMutationHelper provides specific pre and post mutation operations on a Secret object.
// Important, suppose that hash computation is done on Secret's "Data" field.
// However "StringData" can be used as usual to store string information.
type DefaultMutationHelper struct {
	Expected client.Object
}

// blank assignment to verify that ReconcileCockroachDB implements reconcile.Reconciler
var _ oktres.MutationHelper = &DefaultMutationHelper{}

// GetObject return the (K8S) resource Object (i.e. Meta + Runtime part) for used by the MutationHelper
func (r *DefaultMutationHelper) GetObject() client.Object {
	return r.Expected
}

// GetSpec provide a "virtual" Spec for this object that can be used to compute hashable Ref.
// Pre/PostMutate() methods, here, should be built in respect to this choice
func (r *DefaultMutationHelper) GetObjectSpec() interface{} {
	return nil
}

// PreMutate xx
func (r *DefaultMutationHelper) PreMutate() error {
	return nil
}

// PostMutate xx
func (r *DefaultMutationHelper) PostMutate() error {
	return nil
}
