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
	oktres "gitlab.tech.orange/dbmsprivate/operators/okt/resources"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	//"k8s.io/client-go/kubernetes/scheme"
)

// SecretMutationHelper provides specific pre and post mutation operations on a Secret object.
// Important, suppose that hash computation is done on Secret's "Data" field.
// However "StringData" can be used as usual to store string information.
type SecretMutationHelper struct {
	Expected *v1.Secret
}

// blank assignment to verify that ReconcileCockroachDB implements reconcile.Reconciler
var _ oktres.MutationHelper = &SecretMutationHelper{}

// GetObject return the (K8S) resource Object (i.e. Meta + Runtime part) for used by the MutationHelper
func (r *SecretMutationHelper) GetObject() client.Object {
	return r.Expected
}

// GetSpec provide a "virtual" Spec for this object that can be used to compute hashable Ref.
// Pre/PostMutate() methods, here, are built in respect to this choice
func (r *SecretMutationHelper) GetObjectSpec() interface{} {
	return r.Expected.Data
}

// PreMutate xx
func (r *SecretMutationHelper) PreMutate() error {
	//scheme.Default(&r.Expected)
	if r.Expected.StringData == nil {
		r.Expected.StringData = make(map[string]string, 0)
	}
	return nil
}

// PostMutate xx
func (r *SecretMutationHelper) PostMutate() error {
	// If StringData is not used, keep as is Data map
	// If StringData is defined with some values, add/replace (merge) the StringData keys/values in Data
	// The idea is to add only byte encoded Data in Secret HashReference
	if r.Expected.Data == nil && len(r.Expected.StringData) > 0 {
		r.Expected.Data = make(map[string][]byte)
	}
	for key, val := range r.Expected.StringData {
		r.Expected.Data[key] = []byte(val)
	}
	return nil
}
