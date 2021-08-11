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
	oktres "github.com/Orange-OpenSource/Operators-Karma-Tools/resources"
	okterr "github.com/Orange-OpenSource/Operators-Karma-Tools/results"
	"github.com/go-logr/logr"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	sigsreconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Basic xx
type Basic interface {
	sigsreconcile.Reconciler
	//GetLog() logr.Logger
	//GetScheme() *runtime.Scheme
	//GetCR() runtime.Object
	//GetCRMeta() v1.Object
	okterr.Results

	Init(env string, cr client.Object, statusConditions *[]v1.Condition) error
	SetEngine(engine Engine)

	//GetInitialData() map[string]string

	// Return A base name for resource creation (CRName-env)
	GetName() string

	FetchCR(namespacedName types.NamespacedName) error
	Create(resource oktres.Resource, maxCreation uint16) error
	CreateAllResources(maxCreation uint16, stopOnError bool) error
}

// Advanced xx
type Advanced interface {
	Basic
	Mutate(resource oktres.MutableResourceType) error
	Update(resource oktres.MutableResourceType) error
	CreateOrUpdateAllResources(maxCreation uint16, stopOnError bool) error

	/*
		GetAppStatus() (requeueAfterSeconds uint16, err error) // Can return a requeue without error when we are waiting that K8S res are up first
		MutateState() (requeueAfterSeconds uint16, err error)
		//GetApplicationClient() oktclient.Client
		UpdateState() (requeueAfterSeconds uint16, err error)
	*/
}

// Engine reconciler engine
type Engine interface {
	SetLogger(logr.Logger)
	SetResults(okterr.Results)

	Run()
}
