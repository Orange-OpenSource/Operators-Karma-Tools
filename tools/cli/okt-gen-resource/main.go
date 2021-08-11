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
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

var resourceName string
var resourceType string
var opKind string
var opGroup string
var opVersion string
var opPath string
var showList bool

type resourceEntry struct {
	kind           string
	apiVersion     string
	helper         string // Resource helper for added functionalities
	mutationHelper string // Helper to make hash computation easier on some resources
	comment        string
}

var resourcesDico map[string]resourceEntry

// Dictionary defines which resources to use and some properties for the resource code generator
// All resources are indexed by the "type" that is just a uniq/key name used to identify a Resource and its API Version.
func loadDico( /* LATER: pass an alternate dico as parameter, to complete/replace with missing ref */ ) {
	resourcesDico = map[string]resourceEntry{
		"ConfigMap":           {kind: "ConfigMap", apiVersion: "core/v1"},
		"Deployment":          {kind: "Deployment", apiVersion: "apps/v1"},
		"Ingress":             {kind: "Ingress", apiVersion: "networking/v1beta1"},
		"Pod":                 {kind: "Pod", apiVersion: "core/v1"},
		"PodDisruptionBudget": {kind: "PodDisruptionBudget", apiVersion: "policy/v1beta1"},
		"Role":                {kind: "Role", apiVersion: "rbac/v1"},
		"Secret":              {kind: "Secret", apiVersion: "core/v1", mutationHelper: "SecretMutationHelper", comment: "Data map is the data added to the HashReference. StringData is omitted as it is copied (by OKT) into Data before Hash computation."},
		"Service":             {kind: "Service", apiVersion: "core/v1"},
		"ServiceAccount":      {kind: "ServiceAccount", apiVersion: "core/v1"},
		"StatefulSet":         {kind: "StatefulSet", apiVersion: "apps/v1", helper: "StatefulSetHelper"},
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func getNewBackupName(filename string) (string, error) {
	ext := ".bak"
	newName := filename + ext

	if exist := fileExists(newName); exist {
		for i := 0; i < 9 && exist; i++ {
			newName += ext
			exist = fileExists(newName)
		}
		if exist {
			return "none", errors.New("Too much backup files generated, please remove them.")
		}
	}
	return newName, nil
}

// writeToFile will print any string of text to a file safely by
// checking for errors and syncing at the end.
// If the file to create already exists, it is renamed with a ".bak" extension added to the last back file (up to 9 times)
func writeToFile(filename string, data string) error {
	// Creates a back file (.bak) to not overwrite an existing generation
	if fileExists(filename) {
		var err error
		var filenameBak string
		if filenameBak, err = getNewBackupName(filename); err != nil {
			return err
		}
		if err = os.Rename(filename, filenameBak); err != nil {
			return err
		}
	}

	// Generates the file
	fmt.Println("Write file: " + filename)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.WriteString(file, data)
	if err != nil {
		return err
	}
	return file.Sync()
}

func getStubFileName(resourceKind string) string {
	return resourceKind + "ResourceStub.go"
}

func generateStubFile(resourceType string) error {
	resInfo := resourcesDico[resourceType]
	if resInfo.kind == "" {
		return errors.New("Invalid resource kind, not in dictionary: " + resourceType)
	}
	data, err := getResourceStubData(resourceType, &resInfo)
	if err != nil {
		return err
	}

	filename := getStubFileName(resourceType)
	return writeToFile(filename, data)
}

func generateFile(filename, head, body string) error {
	data := head + "\n" + body + "\n"
	return writeToFile(filename, data)
}

func getResourceName(inputFileName string) string {
	filename := path.Base(inputFileName)
	rSlices := strings.Split(filename, ".")
	index := 0
	length := len(rSlices)
	if length >= 2 {
		index = length - 2
	}

	//str := strings.ReplaceAll(rSlices[index], "/", "")
	str := strings.ReplaceAll(rSlices[index], "-", "")

	return str
}

func getOutputFileName(inputFileName string) string {
	return "./" + getResourceName(inputFileName) + ".go"
}

func readManifest(file string) (filename, content string, err error) {
	var bContent []byte

	filename = getResourceName(file)

	if !fileExists(file) {
		return filename, "#No template provided", nil
	}

	bContent, err = ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	content = string(bContent)

	return filename, content, nil
}

func init() {
	flag.BoolVar(&showList, "list", false, "List all resources types available")
	flag.StringVar(&resourceType, "type", "undeftype", "Resource type")
	flag.StringVar(&opKind, "kind", "MyKind", "Operator Kind")
	flag.StringVar(&opGroup, "group", "mygroup", "Operator API Group")
	flag.StringVar(&opVersion, "version", "v1alpha1", "Operator API version")
	flag.StringVar(&opPath, "path", "gitlab.tech.orange/dbmsprivate/operators/myapp-operator", "Operator path")
}

func main() {
	var err error

	flag.Parse()
	loadDico()

	if showList {
		fmt.Println("Resources entries as follow: EntryKey { kind:, apiVersion:, [extension:]}")
		for key, resInfo := range resourcesDico {
			//os.Stdout.WriteString("Type: " + key + "  kind: " + resInfo.kind + "  API Version: " + resInfo.apiVersion + "   Extension: " + resInfo.extension + "\n")
			//fmt.Println("Type:", key, "Resource:", resInfo)
			fmt.Printf("EntryKey %s: %v\n", key, resInfo)
		}
		os.Exit(0)
	}

	// Get tail args
	if len(flag.Args()) < 1 {
		err := errors.New("Error, Manifest files is expected")
		log.Fatal(err)
	}
	if err := generateStubFile(resourceType); err != nil {
		log.Fatal(err)
	}
	resourceName, yaml, err := readManifest(flag.Args()[0])
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("OP Path: " + opPath)
	headData, err := getHeadData(opPath, opGroup, opVersion)
	if err != nil {
		log.Fatal(err)
	}

	bodyData, err := getBodyData(yaml, opKind, resourceType, resourceName)
	if err != nil {
		log.Fatal(err)
	}

	outputFileName := getOutputFileName(resourceName)
	if err := generateFile(outputFileName, headData, bodyData); err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}
