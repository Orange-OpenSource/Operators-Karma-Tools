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

package tools

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var clog = logf.Log.WithName("okt_curl")

// HTTPCurlJSON xx
func HTTPCurlJSON(hostname string, port uint16, path string, data interface{}) error {
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	body := bytes.NewReader(payloadBytes)
	url := "http://" + hostname + ":" + strconv.Itoa(int(port)) + path

	//clog.Info("HTTTP Request: ", "host=", url)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		clog.Error(err, "CURL error", "POST", data)
		return err
	}
	defer resp.Body.Close()

	return nil
}
