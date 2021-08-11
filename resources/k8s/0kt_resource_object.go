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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	oktclients "github.com/Orange-OpenSource/Operators-Karma-Tools/clients"
	oktres "github.com/Orange-OpenSource/Operators-Karma-Tools/resources"
	oktngvk "github.com/Orange-OpenSource/Operators-Karma-Tools/tools/ngvk"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	//"k8s.io/kubernetes/pkg/apis/apps

	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ResourceObject An OKT resource for K8S types which implements the OKT Resource (not Mutable) interface.
// Thus it can not be instancied as is but, instead, is dedicated to either be
// implemented as a new Mutable type provinding the Mutations (with Default and with CR) functions,
// or as a not mutable K8S resource (Read-Only for example).
type ResourceObject struct {
	//context.Context
	oktclients.Kube // Kube implements the OKT Client interface for a Kubernetes object

	//Computed key
	key *oktngvk.NGVK

	// True by default, tell if yes or a call to SetOwnerReference modify Metadata with the owner or not
	EnableOwnerReference bool

	createObj bool // The object is not yet created

	params map[string]string
}

// Blank assignement to check type
var _ oktres.Resource = &ResourceObject{}

// Init Initialize this resource with its Client (K8S) and a runtime object for the Namespace and Name provided
func (or *ResourceObject) Init(client k8sclient.Client, objtyp k8sclient.Object, namespace, name string) error {
	//var err error

	or.EnableOwnerReference = true

	/*
		if or.metav1, err = meta.Accessor(objtyp); err != nil {
			return err
		}
	*/

	or.Client = client
	or.Object = objtyp
	or.Object.SetNamespace(namespace)
	or.Object.SetName(name)

	return or.setIndex()
}

// SetData Set parameters data
func (or *ResourceObject) SetData(params map[string]string) {
	or.params = params
}

// GetData Set parameters data
func (or *ResourceObject) GetData() map[string]string {
	return or.params
}

// GetObject xx
func (or *ResourceObject) GetObject() runtime.Object {
	return or.Object
}

// CopyTpl Apply a yaml manifest to the current Object which can be a template that is executed with the tplValues (if this last is not nil).
// The yaml string can also be used to pass a JSON data string, howver in this case, the tplValues are totaly useless and must be nil.
// Meta are initialised in respect to the OKT principles. See InitResource for details.
func (or *ResourceObject) CopyTpl(yaml string, tplValues interface{}) error {
	var err error
	if err = oktres.DecodeYaml(yaml, tplValues, or.Object); err != nil {
		return err
	}

	// Ensure that default object's NamespacedName is not badly overriden by the template
	sKey := or.NamespacedName().String()

	if err = checkKeys(sKey, or.Object); err != nil {
		return err
	}

	return nil
}

// ApplyGOStruct is a way to copy a source object into a destination object different than the src.DeepCopyInto(dst).
// It Apply the src GO data structure to a resource object preserving all other existing fields in the destination not present in the source struct.
//  Here dst object is the resource to update. and the src is the GO struct provided for the update.
func (or *ResourceObject) CopyGOStruct(src interface{}) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "    ")
	enc.Encode(src)

	jsonStr := buf.String()
	return or.CopyTpl(jsonStr, nil)
}

// SetOwnerReference Set a the owner reference on a controlled object (in metadata.OwnerReferences field).
// Warning: This method do nothing if this object has the EnableOwnerReference value set to False!
// By this way the deletion of the owner (the Custom Resource) will automatically delete the controlled object.
// Only one controller can be indicated as owner to a controlled object.
func (or *ResourceObject) SetOwnerReference(owner runtime.Object, scheme *runtime.Scheme) error {
	var err error
	var metaOwner metav1.Object

	if !or.EnableOwnerReference {
		return nil
	}

	if metaOwner, err = meta.Accessor(owner); err != nil {
		return err
	}
	if err := controllerutil.SetControllerReference(metaOwner, or.Object, scheme); err != nil {
		return err
	}

	return nil
}

/*
// SetPeerClient The client for Get and CRUD operations on Peer object (if any)
func (or *ResourceObject) SetPeerClient(client oktclients.Client) {
	or.Client = client
}
*/

// setIndex checks fields by composing the NGVK (NamespacedName Group Version Kind) key. Abort in case of error.
func (or *ResourceObject) setIndex() error {
	var err error

	// Check that all key fields (NamespacedName Group Version Kind) can be obtained
	or.key, err = oktngvk.New(or.Object)
	if err != nil {
		return err
	}
	return err
}

// Index Return entry a key index (NGVK, aka GroupVersion Kind and Name) string
func (or *ResourceObject) Index() string {
	return or.key.String()
}

// KindName Return Kind/Name string for this resource
func (or *ResourceObject) KindName() string {
	return or.key.KN()
}

// SyncFromPeer Try to get peer object which determines if it is a creation or not
// The caller (typically the Reconciler) is in 3 possibles states regarding the resource:
//     - It dont know if the resource exists on the cluster
//     - It has already got an existing resource on the cluster and need a refresh
//     - It is already informed that the resource doest not exists but ask again => NOTHING WILL BE DONE HERE
func (or *ResourceObject) SyncFromPeer() error {
	// Already done and a creation is required first ?
	if or.createObj {
		return nil
	}

	// Get peer
	if err := or.Get(); err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
		or.createObj = true // Not Found! It's a creation.
	}
	return nil
}

// NamespacedName return NamespacedName type of the registered object
func (or *ResourceObject) NamespacedName() types.NamespacedName {
	return or.key.NamespacedName()
}

// IsCreation Tell if yes or no this entry designate an object to create on the Cluster
func (or *ResourceObject) IsCreation() bool {
	return or.createObj
}

// CreatePeer Creates a resource peer on the Cluster site (at the image of the resource in memory)
func (or *ResourceObject) CreatePeer() error {
	if !or.createObj {
		return errors.New("Peer is presumed yet existing on its end:" + or.Index())
	}

	if err := or.Create(); err != nil {
		return err
	}

	or.createObj = false

	return nil
}

// TODO: Why not comparing the whole key NGVK instead of the only N (NamespacedName) ?
func checkKeys(sKey string, newObj k8sclient.Object) error {
	newKey, err := oktngvk.New(newObj)
	if err != nil {
		return err
	}

	if sKey != newKey.NamespacedName().String() {
		return fmt.Errorf("Discrepancy in object name and/or object namespace: " + sKey + "<-VS->" + newKey.String())
	}

	return nil
}

// CheckExpectedKey Compare NamespacedName before and after mutation. They Must be equals.
// TODO: Take note that the resource type is not compared (but should it be ?)
func (or *ResourceObject) CheckExpectedKey() error {
	sKey := or.NamespacedName().String()

	return checkKeys(sKey, or.Object)
}
