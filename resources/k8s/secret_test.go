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
	b64 "encoding/base64"
	"testing"

	oktres "github.com/Orange-OpenSource/Operators-Karma-Tools/resources"
	okthash "github.com/Orange-OpenSource/Operators-Karma-Tools/tools/hash"
	"github.com/stretchr/testify/require"
	k8sres "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ==== BEGINNING OF STUB (TO GENERATE WITH CLI COMMAND)
// SecretResourceStub an OKT extended Secret resource
type SecretResourceStub struct {
	Expected              k8sres.Secret
	MutableResourceObject // OKT K8S resource
	oktres.MutationHelper
}

// blank assignment to verify that ReconcileCockroachDB implements reconcile.Reconciler
var _ oktres.MutableResource = &SecretResourceStub{}

// GetResourceObject Implement a Stub interface function to get the Mutable Object
func (r *SecretResourceStub) GetResourceObject() *ResourceObject {
	return &r.ResourceObject
}

// GetExpected Implements a Stub interface function to get the Expected object
func (r *SecretResourceStub) GetExpected() *k8sres.Secret {
	return &r.Expected
}

// Init Initialize OKT resource with K8S runtime object in the same Namespace of the Custom Resource
func (r *SecretResourceStub) Init(client k8sclient.Client, namespace, name string) error {
	r.Expected.APIVersion = "core/v1"
	r.Expected.Kind = "Secret"

	r.MutationHelper = &SecretMutationHelper{Expected: &r.Expected}

	return r.MutableResourceObject.Init(client, &r.Expected, namespace, name)
}

// PreMutate xx
func (r *SecretResourceStub) PreMutate(scheme *runtime.Scheme) error {

	if err := r.MutationHelper.PreMutate(); err != nil {
		return err
	}

	return nil
}

// PostMutate xx
func (r *SecretResourceStub) PostMutate(cr k8sclient.Object, scheme *runtime.Scheme) error {
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

// HashableRefSecretHelper xx
type HashableRefSecretHelper struct {
	HashableRefHelper
}

// GetHashableRefHelper provide an helper for the HashableRef interface
// This will help to defines which object(s) data has to be used to detect modifications thanks to the Hash computation
func (r *SecretResourceStub) GetHashableRefHelper() *HashableRefSecretHelper {
	hr := &HashableRefSecretHelper{}
	hr.Init(r.MutationHelper)

	return hr
}

// ==== END OF STUB (TO GENERATE WITH CLI COMMAND)

type myMutableSecret struct {
	SecretResourceStub
}

// Blank assignement to check type
var _ oktres.Mutator = &myMutableSecret{}

func (r *myMutableSecret) GetHashableRef() okthash.HashableRef {
	helper := r.GetHashableRefHelper()
	helper.AddMetaLabels()
	helper.AddSpec()

	return helper
}

// MutateWithInitialData xx
func (r *myMutableSecret) MutateWithInitialData() error {
	return nil
}

// MutateWithCR xx
func (r *myMutableSecret) MutateWithCR() (requeueAfterSeconds uint16, err error) {
	return 0, nil
}

// blank assignment to verify that the resource implements the Mutable interface
//var _ oktres.MutableResource = &myMutableSecret{}

func initTestSecret(t *testing.T, tObj oktres.MutableResourceType) {
	require.False(t, tObj.IsCreation(), "Obj creation flag must be False while no verification to the existence on Cluster has been done for this test")
	require.False(t, tObj.NeedResync(), "Must be false by default")
	require.False(t, tObj.LastSyncState(), "Must be false by default")

	tObj.UpdateSyncStatus(tObj.GetHashableRef())
	require.True(t, tObj.LastSyncState(), "Creation time. Must be True, rigth after the first update of the hash key in annotations")
	require.True(t, tObj.NeedResync(), "Creation time. Must be True, rigth after the first update of the hash key in annotations")

	tObj.UpdateSyncStatus(tObj.GetHashableRef()) // Called twice without new modification!
	require.False(t, tObj.LastSyncState(), "A second call without any modification must show that the object has its hash key up to date")
	require.True(t, tObj.NeedResync(), "The object remains unsynched, even after a second call of UpdateSynchStatus()")
}

func TestMutableSecretObject(t *testing.T) {
	labels := map[string]string{
		"a": "b",
		"c": "d",
	}
	sData := map[string]string{
		"password": "zoz",
	}

	// Create and Init resource and check Sync status
	tObj := myMutableSecret{}
	tObj.Init(nil, "myns", "myname")
	expected := &tObj.Expected
	expected.SetLabels(labels)
	expected.StringData = sData
	initTestSecret(t, &tObj)

	// Modify just a Label and check the detection
	labels = map[string]string{
		"a": "B",
		"c": "D",
	}
	expected.SetLabels(labels)
	tObj.UpdateSyncStatus(tObj.GetHashableRef())
	require.True(t, tObj.LastSyncState(), "Modify just a Label and check the detection")

	// Update with same label values
	labels = map[string]string{
		"a": "B",
		"c": "D",
	}
	expected.SetLabels(labels)
	tObj.UpdateSyncStatus(tObj.GetHashableRef())
	require.False(t, tObj.LastSyncState(), "Update with same label values") // No change

	// Modify data and check status
	b64pass, _ := b64.StdEncoding.DecodeString("ZAZ")

	expected.Data = map[string][]byte{
		"password": b64pass,
	}
	tObj.UpdateSyncStatus(tObj.GetHashableRef())
	require.True(t, tObj.LastSyncState())
}

type myMutableSecret2 struct {
	SecretResourceStub
}

// Blank assignement to check type
var _ oktres.Mutator = &myMutableSecret2{}

func (r *myMutableSecret2) GetHashableRef() okthash.HashableRef {
	helper := r.GetHashableRefHelper()
	helper.AddMetaLabelValues("b", "d")
	helper.AddSpec()

	return helper
}

// MutateWithInitialData xx
func (r *myMutableSecret2) MutateWithInitialData() error {
	return nil
}

// MutateWithCR xx
func (r *myMutableSecret2) MutateWithCR() (requeueAfterSeconds uint16, err error) {
	return 0, nil
}

func TestMutableSecret2Object(t *testing.T) {
	labels := map[string]string{
		"a": "A", // Not in scope for hash computation
		"c": "C", // Not in scope for hash computation
	}
	sData := map[string]string{
		"password": "zoz",
	}

	// Create and Init resource and check Sync status
	tObj := myMutableSecret2{}
	tObj.Init(nil, "myns", "myname")
	expected := &tObj.Expected
	expected.SetLabels(labels)
	expected.StringData = sData
	initTestSecret(t, &tObj)

	// Modify just a Label and check the detection
	labels = map[string]string{
		"b": "B", // In scope for hash computation
	}
	metaObj, _ := meta.Accessor(expected)
	metaObj.SetLabels(labels)
	tObj.UpdateSyncStatus(tObj.GetHashableRef())
	require.True(t, tObj.LastSyncState(), "b is in scope of hash computation and added here")

	labels = map[string]string{
		"a": "A",     // Unchanged but Not in scope for hash computation
		"b": "B",     // Unchanged
		"c": "CDERR", // Changed but Not in scope for hash computation
	}
	metaObj.SetLabels(labels)
	tObj.UpdateSyncStatus(tObj.GetHashableRef())
	require.False(t, tObj.LastSyncState(), "a and c are label keys not in scope for the hash computation, and b key is unchanged")
}

type myMutableSecret3Annotations struct {
	SecretResourceStub
}

// Blank assignement to check type
var _ oktres.Mutator = &myMutableSecret3Annotations{}

func (r *myMutableSecret3Annotations) GetHashableRef() okthash.HashableRef {
	helper := r.GetHashableRefHelper()
	helper.AddMetaAnnotations()

	return helper
}

// MutateWithInitialData xx
func (r *myMutableSecret3Annotations) MutateWithInitialData() error {
	return nil
}

// MutateWithCR xx
func (r *myMutableSecret3Annotations) MutateWithCR() (requeueAfterSeconds uint16, err error) {
	return 0, nil
}
func TestMutableSecret3AnnotationsObject(t *testing.T) {
	labels := map[string]string{
		"a": "A", // Not in scope for hash computation
		"c": "C", // Not in scope for hash computation
	}
	annotations := map[string]string{
		"ann": "Ann", // In scope for hash computation
		"cnn": "Cnn", // In scope for hash computation
	}
	sData := map[string]string{
		"password": "zoz",
	}

	tObj := myMutableSecret3Annotations{}
	tObj.Init(nil, "myns", "myname")
	expected := &tObj.Expected
	expected.SetLabels(labels)
	expected.SetAnnotations(annotations)
	expected.StringData = sData
	initTestSecret(t, &tObj)

	// Modify just an annotation with THE SAME VALUE and check the detection
	metaObj, _ := meta.Accessor(expected)
	annotations = metaObj.GetAnnotations()
	annotations["ann"] = "Ann"
	metaObj.SetAnnotations(annotations)
	tObj.UpdateSyncStatus(tObj.GetHashableRef())
	require.False(t, tObj.LastSyncState(), "annotations are in scope of hash computation and are NOT modified here")

	// Modify just an annotation and check the detection
	annotations = metaObj.GetAnnotations()
	annotations["cnn"] = "NEW VALUE"
	metaObj.SetAnnotations(annotations)
	tObj.UpdateSyncStatus(tObj.GetHashableRef())
	require.True(t, tObj.LastSyncState(), "annotations are in scope of hash computation and modified here")
}

type myMutableSecret4MetaMainFields struct {
	SecretResourceStub
}

// Blank assignement to check type
var _ oktres.Mutator = &myMutableSecret4MetaMainFields{}

func (r *myMutableSecret4MetaMainFields) GetHashableRef() okthash.HashableRef {
	helper := r.GetHashableRefHelper()
	helper.AddMetaMainFields()

	return helper
}

// MutateWithInitialData xx
func (r *myMutableSecret4MetaMainFields) MutateWithInitialData() error {
	return nil
}

// MutateWithCR xx
func (r *myMutableSecret4MetaMainFields) MutateWithCR() (requeueAfterSeconds uint16, err error) {
	return 0, nil
}
func TestMutableSecret4MetaMainFieldsObject(t *testing.T) {
	labels := map[string]string{
		"a": "A", // In scope for hash computation
		"c": "C", // In scope for hash computation
	}
	annotations := map[string]string{
		"ann": "Ann", // In scope for hash computation
		"cnn": "Cnn", // In scope for hash computation
	}
	sData := map[string]string{
		"password": "zoz",
	}

	tObj := myMutableSecret4MetaMainFields{}
	tObj.Init(nil, "myns", "myname")
	expected := &tObj.Expected
	expected.SetLabels(labels)
	expected.SetAnnotations(annotations)
	expected.StringData = sData
	initTestSecret(t, &tObj)

	// Modify just an annotation with THE SAME STRING VALUE and check the (NO) detection => No change here
	annotations = expected.Annotations
	annotations["ann"] = "Ann"
	expected.ObjectMeta.SetAnnotations(annotations)
	tObj.UpdateSyncStatus(tObj.GetHashableRef())
	require.False(t, tObj.LastSyncState(), "annotations are in scope of hash computation and are NOT modified here")

	/* THIS IS NOT SUITABLE IN PRACTICE, AS IS, BECAUSE SetAnnotations remove the current HASH value !! */
	// Add equal annotations in different order and in a different map structure => it's a same resource,so NO difference
	annotationsEqual := map[string]string{
		"cnn":                         "Cnn", // In scope for hash computation
		"ann":                         "Ann", // In scope for hash computation
		okthash.OKTHashAnnotationName: annotations[okthash.OKTHashAnnotationName],
	}
	expected.ObjectMeta.SetAnnotations(annotationsEqual)
	tObj.UpdateSyncStatus(tObj.GetHashableRef())
	require.False(t, tObj.LastSyncState(), "annotations are in scope of hash computation and the content is still the same")

	// Modify just an annotation and check the detection
	annotations = expected.Annotations
	annotations["cnn"] = "NEW VALUE"
	expected.ObjectMeta.SetAnnotations(annotations)
	tObj.UpdateSyncStatus(tObj.GetHashableRef())
	require.True(t, tObj.LastSyncState(), "annotations are in scope of hash computation and modified here")
}

type myMutableSecret5MetaHiddenFieldsObject struct {
	SecretResourceStub
}

// Blank assignement to check type
var _ oktres.Mutator = &myMutableSecret5MetaHiddenFieldsObject{}

func (r *myMutableSecret5MetaHiddenFieldsObject) GetHashableRef() okthash.HashableRef {
	helper := r.GetHashableRefHelper()
	//helper.AddMetaMainFields()
	helper.AddMetaLabels()
	helper.AddMetaAnnotations()

	return helper
}

// MutateWithInitialData xx
func (r *myMutableSecret5MetaHiddenFieldsObject) MutateWithInitialData() error {
	return nil
}

// MutateWithCR xx
func (r *myMutableSecret5MetaHiddenFieldsObject) MutateWithCR() (requeueAfterSeconds uint16, err error) {
	return 0, nil
}

func TestMutableSecret5MetaHiddenFieldsObject(t *testing.T) {
	labels := map[string]string{
		"a": "A", // In scope for hash computation
		"c": "C", // In scope for hash computation
	}
	annotations := map[string]string{
		"ann": "Ann", // In scope for hash computation
		"cnn": "Cnn", // In scope for hash computation
	}
	sData := map[string]string{
		"password": "zoz",
	}

	tObj := myMutableSecret5MetaHiddenFieldsObject{}
	_ = tObj.Init(nil, "myns", "myname")
	expected := &tObj.Expected
	expected.SetLabels(labels)
	expected.SetAnnotations(annotations)
	expected.StringData = sData
	initTestSecret(t, &tObj)

	expected.ObjectMeta.SetLabels(labels)
	tObj.UpdateSyncStatus(tObj.GetHashableRef())
	require.False(t, tObj.LastSyncState(), "Must be false, defaults remains the same and only defaults LABELS are copied into expected")

	var duration int64
	duration = 9185997998585
	expected.ObjectMeta.DeletionGracePeriodSeconds = &duration
	tObj.UpdateSyncStatus(tObj.GetHashableRef())
	require.False(t, tObj.LastSyncState(), "Must be false after a modification on a field out of the HashableRef definition")
}
