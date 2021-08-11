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

package engines

import (
	"github.com/go-logr/logr"
	oktreconciler "gitlab.tech.orange/dbmsprivate/operators/okt/reconciler"
	okterr "gitlab.tech.orange/dbmsprivate/operators/okt/results"
)

// FreeStyleHook Defines the hook called by the engine at each reconciliation loop (after CR fetched on cluster)
type FreeStyleHook interface {
	ReconcileWithCR()
}

// FreeStyle reconciler engine type. Let you manage your own reconciliation process as you want.
// Only the CR is fetched from Cluster before calling the Start() method
type FreeStyle struct {
	logr.Logger
	okterr.Results
	hook FreeStyleHook
}

// Blank assignement to check type
var _ oktreconciler.Engine = &FreeStyle{}

// SetLogger provides the logger to use
func (e *FreeStyle) SetLogger(logr logr.Logger) {
	e.Logger = logr
}

// SetResults provides the operation results (success, errors, ...) generated during the reconciliation
func (e *FreeStyle) SetResults(results okterr.Results) {
	e.Results = results
}

// Run starts a reconciliation loop
// All results (errors, operations, ...) have to be collected in Reconciler's results (Results)data
func (e *FreeStyle) Run() {
	if e.hook != nil {
		e.hook.ReconcileWithCR()
	}
}

// NewFreeStyle creates a reconciler allowing a free reconciliation. However, the hook passed as argument,
// is called after CR is fetched (or not) from the cluster
func NewFreeStyle(hook FreeStyleHook) *FreeStyle {
	return &FreeStyle{hook: hook}
}
