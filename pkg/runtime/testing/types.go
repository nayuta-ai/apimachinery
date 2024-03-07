/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package testing

import "apimachinery/pkg/runtime/schema"

// +k8s:deepcopy-gen:interfaces=apimachinery/pkg/runtime.Object
type TestType struct {
	A string `json:"A,omitempty"`
	B int    `json:"B,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=apimachinery/pkg/runtime.Object
type ExternalTestType struct {
	A string `json:"A,omitempty"`
	B int    `json:"B,omitempty"`
}

func (obj *TestType) GetObjectKind() schema.ObjectKind         { return schema.EmptyObjectKind }
func (obj *ExternalTestType) GetObjectKind() schema.ObjectKind { return schema.EmptyObjectKind }
