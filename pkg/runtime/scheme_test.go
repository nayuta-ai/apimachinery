/*
Copyright 2014 The Kubernetes Authors.

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

package runtime_test

import (
	"testing"

	"apimachinery/pkg/runtime"
	"apimachinery/pkg/runtime/schema"
	runtimetesting "apimachinery/pkg/runtime/testing"
)

func GetTestScheme() *runtime.Scheme {
	internalGV := schema.GroupVersion{Version: runtime.APIVersionInternal}
	externalGV := schema.GroupVersion{Version: "v1"}

	s := runtime.NewScheme()
	// (WIP)
	s.AddKnownTypes(internalGV, &runtimetesting.TestType{})
	s.AddKnownTypes(externalGV, &runtimetesting.ExternalTestType{})
	s.AddKnownTypeWithName(externalGV.WithKind("TestType1"), &runtimetesting.ExternalTestType{})
	s.AddKnownTypeWithName(externalGV.WithKind("TestType2"), &runtimetesting.ExternalTestType{})

	return s
}

func TestKnownTypes(t *testing.T) {
	s := GetTestScheme()
	if len(s.KnownTypes(schema.GroupVersion{Group: "group", Version: "v2"})) != 0 {
		t.Errorf("should have no known types for v2")
	}

	types := s.KnownTypes(schema.GroupVersion{Version: "v1"})
	for _, s := range []string{"TestType1", "TestType2"} {
		if _, ok := types[s]; !ok {
			t.Errorf("missing type %q", s)
		}
	}
}
