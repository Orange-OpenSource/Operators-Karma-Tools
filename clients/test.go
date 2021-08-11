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

package client

import (
	"context"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Test client to address tests needs.
type Test struct {
}

// Blank assignement to check type
var _ client.Client = &Test{}

// Get a resource
func (c *Test) Get(ctx context.Context, key types.NamespacedName, obj client.Object) error {
	return nil
}

// Create creates a resource
func (c *Test) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	return nil
}

// Update update a resource
func (c *Test) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return nil
}

// Delete delete a resource
func (c *Test) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	return nil
}

func (c *Test) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	return nil
}

func (c Test) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return nil
}

func (c *Test) Patch(ctx context.Context, obj client.Object, patch client.Patch, ops ...client.PatchOption) error {
	return nil
}

func (c *Test) RESTMapper() meta.RESTMapper {
	return nil
}

func (c *Test) Scheme() *runtime.Scheme {
	return &runtime.Scheme{}
}

// Status return a StatusWriter
func (c *Test) Status() client.StatusWriter {
	return &Status{}
}

type Status struct {
	//client.StatusWriter
}

var _ client.StatusWriter = &Status{}

func (s Status) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return nil
}

func (s Status) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	return nil
}
