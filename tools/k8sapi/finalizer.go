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

/* Package k8suti brings some utilities on K8S resources and API.
2020, Nov

This file is adapted from this files (under Apache 2.0 License):
	 https://github.com/redhat-cop/operator-utils/blob/master/pkg/util/finalizer.go
	 and https://github.com/kubernetes-sigs/controller-runtime/blob/v0.6.3/pkg/controller/controllerutil/controllerutil.go

*/

package k8suti

import (
	//"github.com/redhat-cop/operator-utils/pkg/util/apis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IsBeingDeleted returns whether this object has been requested to be deleted
func IsBeingDeleted(obj metav1.Object) bool {
	return !obj.GetDeletionTimestamp().IsZero()
}

// HasFinalizer returns whether this object has the passed finalizer
func HasFinalizer(obj metav1.Object, finalizer string) bool {
	for _, fin := range obj.GetFinalizers() {
		if fin == finalizer {
			return true
		}
	}
	return false
}

// AddFinalizer adds the passed finalizer this object
func AddFinalizer(obj metav1.Object, finalizer string) {
	f := obj.GetFinalizers()
	for _, e := range f {
		if e == finalizer {
			return
		}
	}
	obj.SetFinalizers(append(f, finalizer))
}

// RemoveFinalizer removes the passed finalizer from object
func RemoveFinalizer(obj metav1.Object, finalizer string) {
	f := obj.GetFinalizers()
	for i := 0; i < len(f); i++ {
		if f[i] == finalizer {
			f = append(f[:i], f[i+1:]...)
			i--
		}
	}
	obj.SetFinalizers(f)
}
