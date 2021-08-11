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

package resources

import (
	"bytes"
	"html/template"

	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
)

// TplToByteBuffer xx
func TplToByteBuffer(doc string, tplValues interface{}) (bytes.Buffer, error) {
	template := template.New("okt")

	var buf bytes.Buffer

	template, err := template.Parse(doc)
	if err != nil {
		return buf, err
	}

	if err := template.Execute(&buf, tplValues); err != nil {
		return buf, err
	}

	return buf, nil
}

// TplToBytes xx
func TplToBytes(doc string, tplValues interface{}) ([]byte, error) {
	bDoc := []byte(doc)

	if tplValues != nil {
		var buf bytes.Buffer

		var err error
		if buf, err = TplToByteBuffer(doc, tplValues); err != nil {
			return buf.Bytes(), err
		}
		bDoc = buf.Bytes()
	}
	return bDoc, nil
}

// DecodeYaml Initialise the resource object passes as argument with its yaml definition.
// If tplValues is provided (key/values or structs as documented by the html/template GO module), the yaml string
// is considered as a template that will be interpreted.
// The yaml string can also be used to pass a JSON data string, howver in this case, the tplValues are totaly useless and must be nil.
// Note that any data existing in the resource object not described in the yaml file, are preserved by this function.
// Thus the DecodeYaml function is more a merge function that preserve existing values in the resource and override only values provided by the YAML data.
func DecodeYaml(yaml string, tplValues interface{}, resource interface{}) error {
	var err error
	var bDoc []byte

	if bDoc, err = TplToBytes(yaml, tplValues); err != nil {
		return err
	}

	//TODO: Later try resource.Marshall(bytes.NewReader([]byte(yaml))
	dec := k8syaml.NewYAMLOrJSONDecoder(bytes.NewReader(bDoc), 1000)
	if err := dec.Decode(resource); err != nil {
		return err
	}

	return nil
}
