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
	"context"
	"errors"

	k8sapp "k8s.io/api/apps/v1"
	k8score "k8s.io/api/core/v1"

	//"k8s.io/kubernetes/pkg/apis/apps"

	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type StatefulSetStub interface {
	GetResourceObject() *ResourceObject
	GetExpected() *k8sapp.StatefulSet
}

// StatefulSetHelper an OKT extended Statefulset resource
type StatefulSetHelper struct {
	StatefulSetStub
}

// GetReadyPodsCount Returns ready pods count and remaining count with regard to the total count deployed.
// Note that if the StatefulSet is not yet deployed on Cluster, the remaining count is always 0
func (r *StatefulSetHelper) GetReadyPodsCount() (ready, remaining int32) {
	if r.GetResourceObject().IsCreation() {
		return 0, 0
	}
	ready = r.GetExpected().Status.ReadyReplicas

	return ready, *r.GetExpected().Spec.Replicas - ready
}

// GetRunningPodsCount Returns running pods count and remaining count with regard to the total count deployed.
// Note that if the StatefulSet is not yet deployed on Cluster, the remaining count is always 0
func (r *StatefulSetHelper) GetRunningPodsCount() (running, remaining int32, err error) {
	if r.GetResourceObject().IsCreation() {
		return 0, 0, nil
	}

	listOpts := []k8sclient.ListOption{
		k8sclient.InNamespace(r.GetExpected().Namespace),
		k8sclient.MatchingLabels(r.GetExpected().Spec.Template.GetLabels()),
	}
	podList := &k8score.PodList{}

	kcli := r.GetResourceObject().Kube //.(*oktclients.Kube)

	k8scli := kcli.Client

	if err = k8scli.List(context.TODO(), podList, listOpts...); err != nil {
		return 0, 0, errors.New("Unable to convert OKT client to Kube client")
	}

	for _, pod := range podList.Items {
		podPhase := pod.Status.Phase
		if podPhase == "Running" { // TODO: Beurk, no constant ?
			running++
		}
	}

	return running, *r.GetExpected().Spec.Replicas - running, nil
}
