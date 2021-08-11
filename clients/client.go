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

// Client Generic client type
type Client interface {
	Get() error
	Create() error
	Update() error
	Delete() error
}

// AppClient client interface allowing to interact with a service or an application
type AppClient interface {
	Execute(cmd string, params []string) error
	Ping() error
}
