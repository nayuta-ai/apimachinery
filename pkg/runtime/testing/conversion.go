/*
Copyright 2020 The Kubernetes Authors.

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

import (
	"apimachinery/pkg/conversion"
	runtime "apimachinery/pkg/runtime"
)

func convertTestType2ToExternalTestType2(in *TestType, out *ExternalTestType, s conversion.Scope) error {
	out.A = in.A
	out.B = in.B
	return nil
}

func convertExternalTestType2ToTestType2(in *ExternalTestType, out *TestType, s conversion.Scope) error {
	out.A = in.A
	out.B = in.B
	return nil
}

func RegisterConversions(s *runtime.Scheme) error {
	if err := s.AddConversionFunc((*TestType)(nil), (*ExternalTestType)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return convertTestType2ToExternalTestType2(a.(*TestType), b.(*ExternalTestType), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*ExternalTestType)(nil), (*TestType)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return convertExternalTestType2ToTestType2(a.(*ExternalTestType), b.(*TestType), scope)
	}); err != nil {
		return err
	}
	return nil
}
