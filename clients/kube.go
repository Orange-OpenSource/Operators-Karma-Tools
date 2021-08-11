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

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Kube client to address Kube resources
type Kube struct {
	//	context.Context
	client.Client
	client.Object
}

// Blank assignement to check type
var _ Client = &Kube{}

// Get creates a resource
func (c *Kube) Get() error {
	return c.Client.Get(context.TODO(), types.NamespacedName{Namespace: c.Object.GetNamespace(), Name: c.Object.GetName()}, c.Object)
}

// Create creates a resource
func (c *Kube) Create() error {
	return c.Client.Create(context.TODO(), c.Object)
}

// Update update a resource
func (c *Kube) Update() error {
	return c.Client.Update(context.TODO(), c.Object)
}

// Delete delete a resource
func (c *Kube) Delete() error {
	return c.Client.Delete(context.TODO(), c.Object)
}

// NewKube New client mapper
// /!\ The runtime object passed as argument is already IDENTIFIABLE so it must have its Namespace and name defined!!
func NewKube(client client.Client, obj client.Object) *Kube {
	c := &Kube{}
	//	c.Context = ctx
	c.Client = client
	c.Object = obj

	return c
}
