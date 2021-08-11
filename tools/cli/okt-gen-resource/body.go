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

import oktres "gitlab.tech.orange/dbmsprivate/operators/okt/resources"

const templateBody = `// {{ .ResourceName }} xx
type {{ .ResourceName }} struct {
	{{ .ResourceType }}ResourceStub
	cr *appapi.{{ .MyCR }}
}

// blank assignment to verify this resource implements an OKT Resource
var _ oktres.Mutator = &{{ .ResourceName }}{}

// New{{ .ResourceName }} xx
func New{{ .ResourceName }}(cr *appapi.{{ .MyCR }}, client k8sclient.Client, namespace, name string) (*{{ .ResourceName }}, error) {
	res := &{{ .ResourceName }}{cr: cr}

	if err := res.Init(client, namespace, name); err != nil {
		return nil, err
	}

	return res, nil
}

//--
// TODO: CUSTOMIZE HERE YOUR OWN MUTATIONS WITH DEFAULTS AND THE CUSTOM RESOURCE
//--

// GetHashableRef xx
// Note that a Spec reference can always be added by the helper. It is either the K8S Object's Spec or 
// data as defined by OKT dictionary used by the resource code generator
func (r *{{ .ResourceName }}) GetHashableRef() okthash.HashableRef {
	helper := r.GetHashableRefHelper()
	helper.AddMetaLabels()
	//helper.AddUserData(&r.Expected.Spec)  

	return helper
}

// MutateWithInitialData Initialize the Expected object with intial deployment data
func (r *{{ .ResourceName }}) MutateWithInitialData() error {
	yaml := r.{{ .InitialDataInYAMLFuncName }}()
	// Initialize the Expected object with YAML
	if err := r.CopyTpl(yaml, r.GetData()); err != nil {
			return err
	}
	/* Alternatively, set directly the Expected Object GO structure with initial data
	r.Expected.xxx = xxx

	// Or if a GO struct already exists, you can also copy it in Expected object
	// Do not use myStruct.DeepCopyInto(&r.Expected) as this would override all fields and produce unexpected changes
	if err := r.CopyGOStruct(myStruct); err != nil {
		return err
	}
	*/

	return nil
}

// MutateWithCR xx
func (r *{{ .ResourceName }}) MutateWithCR() (requeueAfterSeconds uint16, err error) {
	// Apply CR values
	//r.Expected.Spec.xxx = r.cr.Spec.xxx

	return 0, nil
}
`

func getResourceFunc(resourceName, yaml string) (funcName, funcCode string) {
	funcName = "getTpl"
	funcCode = "func (r " + resourceName + ") " + funcName + "() string {\n  yaml := `\n" + yaml + "\n`\n  return yaml\n}"
	return funcName, funcCode
}

// getBodyData xx
func getBodyData(resourceYaml, opKind, resourceType, resourceName string) (string, error) {
	resourceName = "Resource" + resourceName + "Mutator"
	funcName, yamlFuncDeclaration := getResourceFunc(resourceName, resourceYaml)

	values := map[string]string{
		"ResourceType":              resourceType,
		"ResourceName":              resourceName,
		"InitialDataInYAMLFuncName": funcName,
		"MyCR":                      opKind,
	}

	bTplBytes, err := oktres.TplToBytes(templateBody, values)
	if err != nil {
		return "", err
	}

	return yamlFuncDeclaration + "\n\n" + string(bTplBytes), err
}
