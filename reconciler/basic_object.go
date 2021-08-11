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
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8scond "k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	oktregistry "github.com/Orange-OpenSource/Operators-Karma-Tools/registry"
	oktres "github.com/Orange-OpenSource/Operators-Karma-Tools/resources"
	okterr "github.com/Orange-OpenSource/Operators-Karma-Tools/results"
	okttools "github.com/Orange-OpenSource/Operators-Karma-Tools/tools/k8sapi"
)

// BasicObject Elemantary implementation of a Reconciler only able to Create resource.
// Do not mutate them nor update them. Right for an initial deployement only.
// For a complete set of features, use the Advanced Reconciler object.
/* Basic Example:

type MyReconciler struct {
	oktreconciler.BasicObject
	CR myopv1alpha1.MyApp
}
// Blank assignement to check type
var _ oktreconciler.Engine = &MyReconciler{}

func (r *MyReconciler) ReconcileWithCR() {
	// Your reconciliation logic, here!
	// Note that the CR is already fetched from the Cluster (if it exists)
}

func (r *MyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	r.Log = ctrl.Log.WithName("Memcached")
	r.Scheme = mgr.GetScheme()

	r.Init("dev", &r.CR, r.CR.Status)
	engine := oktengines.NewFreeStyle(r)
	r.SetEngine(engine)
	r.Params = parameters  // Not mandatory

	// Same as the standard way
	return ctrl.NewControllerManagedBy(mgr).
			For(&r.CR).
			Owns(&v1.Secret{}).
			...
			Complete(r)
}

func (r *MyReconciler) Run() {
	// Manage your reconciliation code here and take benefit of OKT Reconciler facilities
	// Store all your error(s) and success operation(s) by using the OKT Results interface

	// CR is already picked from the K8S Cluster
	// Check CR
	if r.CR.xxx .... { do_something }
	if r.CR.yyy .... {
		raisedError = EEE  // Something is wrong in CE We have to giveup here
		err := okterr.ErrGiveUpReconciliation.Reason(raisedError)
		r.Results.AddOp(r.CR, okterr.OperationResultCRSemanticError, err, 0)
		return // Giveup the reconciliation
	}
	// Create your resources
	mysecret := &oktk8s.IngressResource{}
	mysecret.Init(r.Client, r.CR.Namespace(), r.CR.Name()+"-secret1", nil)

	r.Results.AddSuccess(mysecret, oktresults.OperationResultRegistrationSuccess, nil)
}
*/
type BasicObject struct {
	client.Client                 // Should be set by the controller-runtime Manager
	Log           logr.Logger     // Should be set by the controller-runtime Manager
	Scheme        *runtime.Scheme // Should be set by the controller-runtime Manager

	// Results is a buffer of errors and/or misc. operations that happened during reconciliation
	okterr.Results
	cr                      client.Object
	managedStatusConditions *[]v1.Condition
	env                     string
	controllerName          string

	registry *oktregistry.Registry

	engine      Engine
	crIsFetched bool

	// Indicates wether or not te CR has to be finalized
	CRHasToBeFinalized bool

	Params map[string]string
}

// blank assignment to correct implementation
//var _ context.Context = &BasicObject{}
var _ Basic = &BasicObject{}

// Reconcile is the native Reconcile method  (sigs.k8s.io) called by the Operator manager
// This is the interface between OKT Reconciler and the OperatorSDK
func (r *BasicObject) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	r.ResetAllResults() // Reset results and stats to empty
	r.registry = oktregistry.New()

	r.Log.V(1).Info("Reconcile: " + request.NamespacedName.String())

	// Fetch, from Cluster, the Operator's Custom Resource instance
	if err := r.FetchCR(request.NamespacedName); err != nil {
		return r.ConsolidatedSigsK8S()
	}

	// Now, launch the Reconcile process !
	r.engine.Run()

	if r.CRHasToBeFinalized {
		r.RemoveCRFinalizer()
	}

	r.DisplayOpList(r.Log)
	r.DisplayCounters(r.Log)

	return r.ConsolidatedSigsK8S()
}

// Durations expressed in seconds
const (
	requeueDurationOnCRUDError          uint16 = 3
	requeueDurationOnResourceUnreadable uint16 = 5
	requeueDurationOnResultNone         uint16 = 3
	requeueDurationOnCreateDelayed      uint16 = 4
	requeueDurationOnStatusUpdateError  uint16 = 5
)

type crInfo struct {
	index    string
	kindName string
	cr       client.Object
	//key      ngvk.NGVK
}

func (cri *crInfo) Index() string {
	return cri.cr.GetNamespace() + "/" + cri.cr.GetObjectKind().GroupVersionKind().Kind + "/" + cri.cr.GetResourceVersion() + "/" + cri.cr.GetName()
}

func (cri *crInfo) KindName() string {
	return cri.cr.GetObjectKind().GroupVersionKind().Kind + "/" + cri.cr.GetName()
}

// Init initialize this reconciler for the current environement (env name) and the CR type provided
// engine argument is a reconciliation engine to link with.
// The Status type must fulfill the interface OKT Status (okt/results/Status interface) to be taken into account. Else, it is ignored.
func (r *BasicObject) Init(env string, cr client.Object, statusConditions *[]v1.Condition) error {
	//r.k8sClient = manager.GetClient()
	//r.Scheme = manager.GetScheme()
	//r.Log = r.Log.WithValues("ENV", env)
	r.Log = r.Log.WithName("ENV=" + env)
	r.env = env
	r.Results = okterr.NewResultList()
	r.cr = cr

	if statusConditions != nil {
		r.managedStatusConditions = statusConditions
		r.Results.AddOpSuccess(&crInfo{cr: r.cr}, okterr.OperationResultStatusEnabled)
	}

	// Initialize params list for convenience at resource level
	if r.Params == nil {
		r.Params = make(map[string]string, 0)
	}

	return nil
}

const (
	// statusConditionType stands for a Reconciliation status
	statusConditionType = "ReconciliationSuccess"
)

// GetManagedStatusConditionType Return the status condition type managed by the Reconciler
// If no condition list has been provided during the Init() call for this reconciler, return an empty string ("").
func (r *BasicObject) GetManagedStatusConditionType() string {
	if r.managedStatusConditions != nil {
		return statusConditionType
	}
	return ""
}

// SetEngine Set the OKT engine to use with this reconciler
func (r *BasicObject) SetEngine(engine Engine) {
	r.engine = engine
	engine.SetLogger(r.Log)
	engine.SetResults(r.Results)
}

// GetCR return Custom Resource
func (r *BasicObject) GetCR() client.Object {
	return r.cr
}

// GetCRMeta return CR's metadata
// DEPRECATED: kept to not break interface with older implementations (before sigs.k8s/.../client.Object introduction). Use GetCR instead.
func (r *BasicObject) GetCRMeta() v1.Object {
	return r.cr
}

// GetScheme return Custom Resource
func (r *BasicObject) GetScheme() *runtime.Scheme {
	return r.Scheme
}

// GetLog return logger
func (r *BasicObject) GetLog() logr.Logger {
	return r.Log
}

// GetName return the application name based on the CR name + the environement specified at the application Init
// Return "" if the CR is not yet fetched from the Cluster
func (r *BasicObject) GetName() string {
	if r.crIsFetched {
		return r.controllerName
	}
	return ""
}

// FetchCR Fetch Custom (primary) Resource from the K8S Cluster using the Client
// provided by the Operator's Manager for this Reconciler
func (r *BasicObject) FetchCR(namespacedName types.NamespacedName) error {
	err := r.Client.Get(context.TODO(), namespacedName, r.cr)

	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return r.Results.AddGiveupError(nil, okterr.OperationResultCRNotExist, nil) // Not an alarming error
		}
		// Error reading the object - requeue the request.
		return r.Results.AddOp(nil, okterr.OperationResultCRUnreadable, err, requeueDurationOnResourceUnreadable)
	}
	r.crIsFetched = true
	r.controllerName = r.cr.GetName() //TODO? + "-" + r.env
	r.CRHasToBeFinalized = okttools.IsBeingDeleted(r.cr) && okttools.HasFinalizer(r.cr, r.controllerName)

	if r.CRHasToBeFinalized {
		r.Results.AddOpSuccess(&crInfo{cr: r.cr}, okterr.OperationResultCRIsFinalizing)
	}

	r.Results.AddOpSuccess(&crInfo{cr: r.cr}, okterr.OperationResultFetchCRSuccess)
	return nil
}

// RemoveCRFinalizer Remove CR finalizer (based on Controller Name) and update CR on Cluster.
// You must Ensure first that the CR has a finalizer to remove!
func (r *BasicObject) RemoveCRFinalizer() error {
	okttools.RemoveFinalizer(r.cr, r.controllerName)

	if err := r.Client.Update(context.TODO(), r.cr); err != nil {
		r.Results.AddOp(&crInfo{cr: r.cr}, okterr.OperationResultCRUDError, err, 0)
		return err
	}

	return nil
}

// RegisterResource Register OKT Resource in Reconciler registry.
// Sync OKT Resource with Peer and check its modification status.
// The modification status is obtained thanks to a hash key computed on the objects' spec.
// Any modification is thus detected because a new computed key will produce a new hash
// different than the one stored in the annotations.
// TODO: Later let the possibility to specify the client to use for each resource (as parameter ?) and
// not use systematicaly the client provided by the Manager ?
func (r *BasicObject) RegisterResource(resource oktres.Resource) error {
	var err error

	// It is a K8S resource, the peer client is the the configured client for K8S Cluster provided by the manager
	//resource.SetPeerClient(r.Client)

	// Put in reconciler registry
	if err = r.registry.AddEntry(resource); err != nil {
		return r.Results.AddGiveupError(resource, okterr.OperationResultImplementationConcern, err)
	}

	// Get Peer object if it exists and then concludes it is a creation of a new resource or not
	if err := resource.SyncFromPeer(); err != nil {
		return r.Results.AddOp(resource, okterr.OperationResultResourceUnreadable, err, requeueDurationOnResourceUnreadable)
	}

	// Propagate params, if any, to this resource
	resource.SetData(r.Params)

	r.Results.AddOpSuccess(resource, okterr.OperationResultRegistrationSuccess)
	return nil
}

// GetResource Return the OKT resource from its index in the registry
func (r *BasicObject) GetResource(index string) oktres.Resource {
	return r.registry.GetEntry(index)
}

// GetRegisteredResources Return a slice on all entry pointers in the registry
func (r *BasicObject) GetRegisteredResources() []oktres.Resource {
	return r.registry.Entries()
}

// Create creates the given K8S resource on Kubernetes Cluster
// Warning!! Before calling this function you must ensure (with res.IsCreation() test) that the resource need to be created and does not yet exist (else it will raise an error)
// MaxCreation is a limit you want to set to avoid to create too much resources during one reconciliation phase. If set to 0, there's NO limit.
// In case where the max creation count were reached, the create request is not called and a new reconciliation request dealyed at later time (set by requeueDurationOnCreateDelayed seconds)
// This function adds operation's result in reconciler's Results list
// Returns the error if any.
func (r *BasicObject) Create(resource oktres.Resource, maxCreation uint16) error {
	if maxCreation > 0 {
		if count := r.OpsCount(okterr.OperationResultCreated); count > maxCreation {
			r.AddOp(resource, okterr.OperationResultCreateDelayed, nil, requeueDurationOnCreateDelayed)
			return nil
		}
	}

	if err := resource.CreatePeer(); err != nil {
		return r.AddOp(resource, okterr.OperationResultCRUDError, err, requeueDurationOnCRUDError)
	}
	r.AddOpSuccess(resource, okterr.OperationResultCreated)
	return nil
}

// CreateAllResources is a convenient method to Create all OKT resources (taking care of their created status)
// The parameter maxCreation specified the maximum count of resources to create in one shot
// Return immediatley if a raised or current consolidated error is GiveUpReconciliation
// If stopOnError is true, stop as soon as an error is raised
// Return the last raised error during the updates
func (r *BasicObject) CreateAllResources(maxCreation uint16, stopOnError bool) error {
	if giveup, err := r.ConsolidatedError(); giveup {
		return err
	}

	var err error

	for _, res := range r.GetRegisteredResources() {
		if res.IsCreation() {
			if err = r.Create(res, maxCreation); err != nil {
				if stopOnError {
					return err
				}
			}
		}
	}

	return err
}

// ManageSuccess Take care of the Status data of the CR for this reconciler and update it if possible
// The Status type must fulfill the interface OKT Status (okt/results/Status)
func (r *BasicObject) ManageSuccess() {
	if r.managedStatusConditions == nil {
		return
	}

	opsCountType, opsCount := r.Results.TotalOpsCount()
	msg := fmt.Sprint(opsCount, " successful operation(s) of ", opsCountType, "  different type(s)")
	s := v1.Condition{}
	s.Type = statusConditionType
	s.Status = v1.ConditionTrue
	s.Reason = "Success"
	s.Message = msg
	//s.LastTransitionTime = v1.Now()

	k8scond.SetStatusCondition(r.managedStatusConditions, s)

	if err := r.Client.Status().Update(context.Background(), r.cr); err != nil {
		// Do not track this error but requeue
		r.Results.AddOp(&crInfo{cr: r.cr}, okterr.OperationResultStatusUpdateError, nil, requeueDurationOnStatusUpdateError)
		return
	}
	r.Results.AddOpSuccess(&crInfo{cr: r.cr}, okterr.OperationResultStatusUpdated)
}

// ManageError Take care of the Status data conditions of the CR for this reconciler and update it if possible
// The condition Type managed here is "Reconciliation" with a Status set to True in case of Success or False in case of Error
// In case of recurrent error, the timer interval growth up exponentialy at each reconciliation cycle (except for an Update Status problem).
func (r *BasicObject) ManageError() {
	if r.managedStatusConditions == nil {
		return
	}

	giveup, err := r.Results.ConsolidatedError()

	s := v1.Condition{}
	s.Type = statusConditionType
	s.Status = v1.ConditionFalse
	s.Reason = "Error"
	if giveup {
		s.Reason += "AndGiveup"
	}
	s.Message = err.Error()

	var timeInterval uint16
	timeInterval = 0
	sameError := false

	// Is the last Error Condition Equal i.e same Type, same Status AND same Reason than the current error ?
	if k8scond.IsStatusConditionPresentAndEqual(*r.managedStatusConditions, s.Type, s.Status) {
		prevS := k8scond.FindStatusCondition(*r.managedStatusConditions, s.Type)
		if prevS.Reason == s.Reason && prevS.Message == s.Message {
			sameError = true
			timeInterval = uint16(s.LastTransitionTime.Sub(v1.Now().Time).Round(time.Second).Seconds())
			s.LastTransitionTime = prevS.LastTransitionTime
		}

	}
	k8scond.SetStatusCondition(r.managedStatusConditions, s)

	if sameError {
		r.Results.AddOp(&crInfo{cr: r.cr}, okterr.OperationResultSameStatusError, err, timeInterval)
	}

	if err := r.Client.Status().Update(context.Background(), r.cr); err != nil {
		// Do not track this error but requeue
		r.Results.AddOp(&crInfo{cr: r.cr}, okterr.OperationResultStatusUpdateError, nil, requeueDurationOnStatusUpdateError)
		return
	}

	r.Results.AddOpSuccess(&crInfo{cr: r.cr}, okterr.OperationResultStatusUpdated)
}

/*
// Implementation of the context.Context interface,  as a context.Background()
//TODO: Right now an EmptyCtx, but could evolve to deal with deadline and values....

// Deadline No deadline is set
func (*BasicObject) Deadline() (deadline time.Time, ok bool) {
	return
}

// Done This context can never be cancelled
func (*BasicObject) Done() <-chan struct{} {
	return nil
}

// Err xx
func (*BasicObject) Err() error {
	return nil
}

// Value This context has no value
func (*BasicObject) Value(key interface{}) interface{} {
	return nil
}
*/
