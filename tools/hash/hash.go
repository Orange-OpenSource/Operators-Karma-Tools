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

package hash

import (
	"fmt"
	"hash/fnv"

	"github.com/davecgh/go-spew/spew"
	meta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	// OKTHashAnnotationName is hash key to annotate a Kubernetes resource
	// with the hash of its template.
	OKTHashAnnotationName = "operator.k8s.orange.com/okt-hash"
)

// SetTemplateHashAnnotation adds an annotation containing the hash of the given template into the
// given annotations. This annotation can then be used for template comparisons.
func SetTemplateHashAnnotation(annotations map[string]string, template interface{}) map[string]string {
	return setHashAnnotation(OKTHashAnnotationName, annotations, template)
}

func setHashAnnotation(annotationName string, annotations map[string]string, template interface{}) map[string]string {
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[annotationName] = Compute(template)
	return annotations
}

// GetTemplateHashAnnotation returns the template hash annotation value if set, or an empty string.
func GetTemplateHashAnnotation(annotations map[string]string) string {
	return annotations[OKTHashAnnotationName]
}

// GenerateNew Get object's template hash from its annotation (cur hash), if it exists, and compute the new object's template hash.
// The object (obj) passed as argument, is storing the hash value in its metadata annotations.
// The template passed as argument represents all the fields and data that must be taken under account to compute the Hash value.
// If they are different, then store/replace the old value in annotation by the new one. Then return true (new generated).
// It returns false (not new!) if the object's template hash is equal to the current hash value in annotations.
// It returns true if the object had no hash value.
// The template should be a relevant part (the Spec if it exists or another significant part) on which the Hash
// key can be computed independantly to any un-relevant modifications on the object.
func GenerateNew(obj runtime.Object, template interface{}) (bool, error) {
	metaObj, err := meta.Accessor(obj)
	if err != nil {
		return false, err
	}

	//version := metaObj.GetResourceVersion()
	//metaObj.SetResourceVersion("")

	annotations := metaObj.GetAnnotations()

	var curHash, newHash string

	if annotations != nil {
		curHash = GetTemplateHashAnnotation(annotations)
		delete(annotations, OKTHashAnnotationName)
		//metaObj.SetAnnotations(newAnnotations)
	}

	newAnnotations := SetTemplateHashAnnotation(annotations, template)
	metaObj.SetAnnotations(newAnnotations)
	newHash = GetTemplateHashAnnotation(metaObj.GetAnnotations())

	// Restore version
	//metaObj.SetResourceVersion(version)

	//fmt.Printf("Hashes: %v / %v\n", curHash, newHash)

	return curHash != newHash, nil
}

// Compute writes the specified object to a hash using the spew library
// which follows pointers and prints actual values of the nested objects
// ensuring the hash does not change when a pointer changes.
// The returned hash can be used for object comparisons.
//
// This is inspired by controller revisions in StatefulSets and ElasticSearch Operator:
// https://github.com/kubernetes/kubernetes/blob/8de1569ddae62e8fab559fe6bd210a5d6100a277/pkg/controller/history/controller_history.go#L89-L101
// https://github.com/elastic/cloud-on-k8s/blob/master/pkg/controller/common/hash/hash.go
func Compute(obj interface{}) string {
	hf := fnv.New32()
	printer := spew.ConfigState{
		Indent:         " ",
		SortKeys:       true,
		DisableMethods: true,
		SpewKeys:       true,
	}
	_, _ = printer.Fprintf(hf, "%#v", obj)
	return fmt.Sprint(hf.Sum32())
}
