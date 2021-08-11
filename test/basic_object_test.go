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

package reconciler

import (
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"

	"github.com/stretchr/testify/require"
	k8sres "k8s.io/api/core/v1"
	k8scond "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"

	oktclient "gitlab.tech.orange/dbmsprivate/operators/okt/clients"
	oktreconciler "gitlab.tech.orange/dbmsprivate/operators/okt/reconciler"
	oktengines "gitlab.tech.orange/dbmsprivate/operators/okt/reconciler/engines"
	oktres "gitlab.tech.orange/dbmsprivate/operators/okt/resources"
	okthelpers "gitlab.tech.orange/dbmsprivate/operators/okt/resources/k8s"
	okterr "gitlab.tech.orange/dbmsprivate/operators/okt/results"
)

// ==== BEGINNING OF STUB (TO GENERATE WITH CLI COMMAND)
// Resource type ConfigMap: &amp;{ConfigMap core/v1   }

// ConfigMapResourceStub an OKT extended ConfigMap resource
type ConfigMapResourceStub struct {
	Expected                         k8sres.ConfigMap
	okthelpers.MutableResourceObject // OKT K8S resource
	oktres.MutationHelper
}

// blank assignment to verify that ReconcileCockroachDB implements reconcile.Reconciler
var _ oktres.MutableResource = &ConfigMapResourceStub{}

// GetResourceObject Implement a Stub interface function to get the Mutable Object
func (r *ConfigMapResourceStub) GetResourceObject() *okthelpers.ResourceObject {
	return &r.ResourceObject
}

// GetExpected Implements a Stub interface function to get the Expected object
func (r *ConfigMapResourceStub) GetExpected() *k8sres.ConfigMap {
	return &r.Expected
}

// Init Initialize OKT resource with K8S runtime object in the same Namespace of the Custom Resource
func (r *ConfigMapResourceStub) Init(client k8sclient.Client, namespace, name string) error {
	r.Expected.APIVersion = "core/v1"
	r.Expected.Kind = "ConfigMap"
	r.MutationHelper = &okthelpers.DefaultMutationHelper{Expected: &r.Expected}

	return r.MutableResourceObject.Init(client, &r.Expected, namespace, name)
}

// PreMutate xx
func (r *ConfigMapResourceStub) PreMutate(scheme *runtime.Scheme) error {
	if err := r.MutationHelper.PreMutate(); err != nil {
		return err
	}
	return nil
}

// PostMutate xx
func (r *ConfigMapResourceStub) PostMutate(cr k8sclient.Object, scheme *runtime.Scheme) error {
	if scheme != nil {
		if err := r.SetOwnerReference(cr, scheme); err != nil {
			return err
		}
	}
	if err := r.MutationHelper.PostMutate(); err != nil {
		return err
	}
	return nil
}

// GetHashableRefHelper provide an helper for the HashableRef interface
// This will help to defines which object(s) data has to be used to detect modifications thanks to the Hash computation
// The AddSpec() method of this helper adds the whole "Spec" object in the hashable Ref.
// The Spec can be something else than the typical K8S Spec resource part when it does not exist in the K8S API definition
// It can be defined by the MutationHelper for this resource, eventualy in accordance with Pre/PostMutate() methods provided with the helper.
func (r *ConfigMapResourceStub) GetHashableRefHelper() *okthelpers.HashableRefHelper {
	hr := &okthelpers.HashableRefHelper{}
	hr.Init(r.MutationHelper)

	return hr
}

// ==== END OF STUB (TO GENERATE WITH CLI COMMAND)

type crtestNoOKTStatus struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status statustest `json:"status,omitempty"`
}

func (c *crtestNoOKTStatus) DeepCopyObject() runtime.Object { return nil }

type crtestOKTStatus struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status statusOKTtest `json:"status,omitempty"`
}

func (c *crtestOKTStatus) DeepCopyObject() runtime.Object { return nil }

type statustest struct {
	Nodes []string `json:"nodes"`
}
type statusOKTtest struct {
	Nodes      []string           `json:"nodes"`
	Conditions []metav1.Condition `json:"conditions"`
}

type myReconciler struct {
	oktreconciler.BasicObject

	//engine    *oktengines.FreeStyle
	t *testing.T
	//oktclient oktclient.Client
}

func (r *myReconciler) ReconcileWithCR() {
	res := &ConfigMapResourceStub{}
	err := r.RegisterResource(res)
	require.NoError(r.t, err, "Register without error")
}

func basicobjtestGetObjs() (logr.Logger, *oktclient.Test) {
	zapLog, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("who watches the watchmen (%v)?", err))
	}
	logger := zapr.NewLogger(zapLog)

	client := &oktclient.Test{}

	return logger, client
}

type testError struct {
	reason error
}

func (e testError) Error() string {
	return "NA"
}

var myTestError error = &testError{}

func TestBasicReconcilerNoStatus(t *testing.T) {
	rec := &myReconciler{}
	rec.Log, rec.Client = basicobjtestGetObjs()

	cr := &crtestNoOKTStatus{}
	rec.Init("test1", cr, nil)
	engine := oktengines.NewFreeStyle(rec)
	rec.SetEngine(engine)
	rec.Log.Info("Init done")
	rec.AddOp(nil, okterr.OperationResultMutationSuccess, myTestError, 1)
	rec.ManageError()
	condType := rec.GetManagedStatusConditionType()
	require.EqualValues(t, "", condType, "The condition type should not be defined")
}

func TestBasicReconcilerStatus(t *testing.T) {
	rec := &myReconciler{}
	rec.Log, rec.Client = basicobjtestGetObjs()

	cr := &crtestOKTStatus{}
	//cr.Status.Conditions = make([]metav1.Condition, 0, 3)

	rec.Init("test2", cr, &cr.Status.Conditions)
	engine := oktengines.NewFreeStyle(rec)
	rec.SetEngine(engine)
	rec.Log.Info("Init done")
	rec.ManageSuccess()
	condType := rec.GetManagedStatusConditionType()
	require.NotEmpty(t, condType, "The condition type should be defined")
	condition := k8scond.FindStatusCondition(cr.Status.Conditions, condType)
	require.NotNil(t, condition, "The Status condition must exists")
}
