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

package registry

import (
	"testing"

	//	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	//	corev1 "k8s.io/api/core/v1"
	//	meta "k8s.io/apimachinery/pkg/api/meta"
	//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	oktk8s "gitlab.tech.orange/dbmsprivate/operators/okt/resources/k8s"
)

func TestRegistryNew(t *testing.T) {
	registry := New()

	require.NotNil(t, registry)
	require.Equal(t, len(registry.elems), 0)

	registry.addEntry(&oktk8s.ResourceObject{})
	require.Equal(t, len(registry.elems), 1)
	registry.Reset()
	require.Equal(t, len(registry.elems), 0)

	//registry.AddEntry()
}
