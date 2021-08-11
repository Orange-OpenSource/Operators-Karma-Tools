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

package reflect

import (
	"errors"
	"reflect"
)

// IsMethodPureInterface Checks that the method to call is not a pure interface (vs. an interface implemented in a struct)
// Else, display a generic message and continue without error
func IsMethodPureInterface(ptrOnStruct interface{}, methodName string) bool {
	ptr := reflect.ValueOf(ptrOnStruct)
	value := ptr.Elem()

	interfaceMethod := value.MethodByName(methodName)

	if interfaceMethod.IsValid() {
		return true
	}
	return false
}

// GetMethodValue Get the reflect.Value of a method if found in the struct pointed
// by the pointer passed as argument
func GetMethodValue(ptrOnStruct interface{}, methodName string) (reflect.Value, error) {
	ptr := reflect.ValueOf(ptrOnStruct)
	method := ptr.MethodByName(methodName)
	if method.IsValid() {
		return method, nil
	}
	return method, errors.New("Invalid method: " + methodName)
}

// CallHook  Call function: func hooks.hookName(ctx) {}
// Panic error at runtime, surely due to the removal of returned error (Works in CallErrorHook)
func CallHook(hooks interface{}, hookName string, ctx interface{}) {
	finalMethod, err := GetMethodValue(hooks, hookName)
	if err != nil {
		panic(err)
	}

	// Call hook method with parameters and return error if any
	param := reflect.ValueOf(ctx)
	paramValue := []reflect.Value{param}
	_ = finalMethod.Call(paramValue)[0]
}

// CallErrorHook  Call function returning an error: func hooks.hookName(ctx) error {}
// Tested and working
func CallErrorHook(hooks interface{}, hookName string, ctx interface{}) error {
	if IsMethodPureInterface(hooks, hookName) {
		return errors.New("Invalid method: " + hookName)
	}

	finalMethod, err := GetMethodValue(hooks, hookName)
	if err != nil {
		panic(err)
	}

	// Call hook method with parameters and return error if any
	param := reflect.ValueOf(ctx)
	paramValue := []reflect.Value{param}
	returnValue := finalMethod.Call(paramValue)[0]

	if returnValue.IsNil() {
		return nil
	}

	return returnValue.Interface().(error)
}
