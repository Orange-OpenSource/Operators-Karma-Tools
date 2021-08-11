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

package results

// ErrGiveUp defines an error which cause the end of the reconciliation to the current state
type ErrGiveUp struct {
	reason error
}

// blank assignment to verify that ErrGiveUp implements error interface
var _ error = &ErrGiveUp{}

// Reason Set the reason (actualy an alarming error) why we give up the current reconciliation
// In conformance with the error interface requirement the returned error is the error (ErrGiveUp) itself
// Providing no reason (nil) is about not alarming issue, just want to stop the reconciliation until a new event occur
func (e *ErrGiveUp) Reason(alarmingError error) error {
	e.reason = alarmingError
	return e
}

// AlarmingReasonToGiveup Get the reason (actualy the error) why we give up the current reconciliation
// A nil error is about not alarming issue, just want to stop the reconciliation until a new event occur
func (e *ErrGiveUp) AlarmingReasonToGiveup() error {
	return e.reason
}

func (e ErrGiveUp) Error() string {
	if e.reason != nil {
		return "Reason: " + e.reason.Error()
	}
	return "Given up on not alarming issue..."
}

// ErrGiveUpReconciliation the pointer on the exported instance of an error of type ErrGiveUp
var ErrGiveUpReconciliation *ErrGiveUp = &ErrGiveUp{}
