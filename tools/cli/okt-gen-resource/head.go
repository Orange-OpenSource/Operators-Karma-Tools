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

package main

import (
	"strings"

	oktres "github.com/Orange-OpenSource/Operators-Karma-Tools/resources"
)

// GetHeadData xx
func getHeadData(opPath, opGroup, opVersion string) (string, error) {
	const templateHead = `
package controllers

import (
	appapi "{{ .opPath }}/api/{{ .opVersion }}"

	oktres "github.com/Orange-OpenSource/Operators-Karma-Tools/resources"
	okthash "github.com/Orange-OpenSource/Operators-Karma-Tools/tools/hash"

	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)
`

	values := map[string]string{
		"opPath":    opPath,
		"opVersion": opVersion,
		"opGroup":   strings.ToLower(opGroup),
	}

	var bHead []byte
	var err error

	if bHead, err = oktres.TplToBytes(templateHead, values); err != nil {
		return "", err
	}

	return string(bHead), nil
}
