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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestHashObject(t *testing.T) {
	// nil objects hash the same
	require.Equal(t, Compute(nil), Compute(nil))

	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "name",
			Labels: map[string]string{
				"a": "b",
				"c": "d",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "container",
					Env: []corev1.EnvVar{
						{
							Name:  "var1",
							Value: "value1",
						},
					},
				},
			},
		},
	}
	samePod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "name",
			Labels: map[string]string{
				"a": "b",
				"c": "d",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "container",
					Env: []corev1.EnvVar{
						{
							Name:  "var1",
							Value: "value1",
						},
					},
				},
			},
		},
	}

	// hashes are consistent
	hash := Compute(pod)
	// same object
	require.Equal(t, hash, Compute(pod))
	// different object but same content
	require.Equal(t, hash, Compute(samePod))

	// /!\ hashing an object and its pointer lead to different values
	require.NotEqual(t, hash, Compute(&pod))

	// hashes ignore different pointer addresses
	userID := int64(123)
	securityContext1 := corev1.PodSecurityContext{RunAsUser: &userID}
	securityContext2 := corev1.PodSecurityContext{RunAsUser: &userID}
	pod.Spec.SecurityContext = &securityContext1
	hash = Compute(pod)
	pod.Spec.SecurityContext = &securityContext2
	require.Equal(t, hash, Compute(pod))

	// different hash on any object modification
	pod.Labels["c"] = "newvalue"
	require.NotEqual(t, hash, Compute(pod))
}

func TestSetTemplateHashAnnotation(t *testing.T) {
	spec := corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name: "container",
				Env: []corev1.EnvVar{
					{
						Name:  "var1",
						Value: "value1",
					},
				},
			},
		},
	}
	labels := map[string]string{
		"a": "b",
		"c": "d",
	}
	expected := map[string]string{
		"a":                   "b",
		"c":                   "d",
		OKTHashAnnotationName: Compute(spec),
	}
	require.Equal(t, expected, SetTemplateHashAnnotation(labels, spec))
}

func TestGenerateNew(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "name",
			Labels: map[string]string{
				"a": "b",
				"c": "d",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "container",
					Env: []corev1.EnvVar{
						{
							Name:  "var1",
							Value: "value1",
						},
					},
				},
			},
		},
	}

	var err error
	var diff bool
	var metaObj metav1.Object

	rawHash := Compute(&pod.Spec)
	assert.NotNil(t, rawHash)

	diff, err = GenerateNew(&pod, &pod.Spec)
	require.Equal(t, nil, err)
	require.Equal(t, true, diff) // There was no curHash, thus a new hash has been computed

	metaObj, err = meta.Accessor(&pod)
	assert.Nil(t, err)
	assert.NotNil(t, metaObj)

	annotations := metaObj.GetAnnotations()
	assert.NotNil(t, annotations)
	hash := GetTemplateHashAnnotation(annotations)
	assert.NotNil(t, hash)
	require.Equal(t, hash, rawHash)

	diff, err = GenerateNew(&pod, &pod.Spec)
	require.Equal(t, nil, err)
	require.Equal(t, false, diff) // No new hash has been computed

	pod.Spec.NodeName = pod.Spec.NodeName + "X"

	diff, err = GenerateNew(&pod, &pod.Spec)
	require.Equal(t, nil, err)
	require.Equal(t, true, diff) // A new hash has been computed
}
