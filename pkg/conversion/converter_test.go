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

package conversion

import (
	"fmt"
	"reflect"
	"testing"
)

func TestConverter_byteSlice(t *testing.T) {
	c := NewConverter(nil)
	src := []byte{1, 2, 3}
	dest := []byte{}
	err := c.Convert(&src, &dest, nil)
	if err != nil {
		t.Fatalf("expected no error")
	}
	if e, a := src, dest; !reflect.DeepEqual(e, a) {
		t.Errorf("expected %#v, got %#v", e, a)
	}
}

func TestConverter_CallsRegisteredFunctions(t *testing.T) {
	type A struct {
		Foo string
		Baz int
	}
	type B struct {
		Bar string
		Baz int
	}
	type C struct{}
	c := NewConverter(nil)
	convertFn1 := func(in *A, out *B, s Scope) error {
		out.Bar = in.Foo
		out.Baz = in.Baz
		return nil
	}
	if err := c.RegisterUntypedConversionFunc(
		(*A)(nil), (*B)(nil),
		func(a, b interface{}, s Scope) error {
			return convertFn1(a.(*A), b.(*B), s)
		},
	); err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	convertFn2 := func(in *B, out *A, s Scope) error {
		out.Foo = in.Bar
		out.Baz = in.Baz
		return nil
	}
	if err := c.RegisterUntypedConversionFunc(
		(*B)(nil), (*A)(nil),
		func(a, b interface{}, s Scope) error {
			return convertFn2(a.(*B), b.(*A), s)
		},
	); err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	x := A{"hello, intrepid test reader!", 3}
	y := B{}

	if err := c.Convert(&x, &y, nil); err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if e, a := x.Foo, y.Bar; e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
	if e, a := x.Baz, y.Baz; e != a {
		t.Errorf("expected %v, got %v", e, a)
	}

	z := B{"all your test are belong to us", 42}
	w := A{}

	if err := c.Convert(&z, &w, nil); err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if e, a := z.Bar, w.Foo; e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
	if e, a := z.Baz, w.Baz; e != a {
		t.Errorf("expected %v, got %v", e, a)
	}

	convertFn3 := func(in *A, out *C, s Scope) error {
		return fmt.Errorf("C can't store an A, silly")
	}
	if err := c.RegisterUntypedConversionFunc(
		(*A)(nil), (*C)(nil),
		func(a, b interface{}, s Scope) error {
			return convertFn3(a.(*A), b.(*C), s)
		},
	); err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if err := c.Convert(&A{}, &C{}, nil); err == nil {
		t.Errorf("unexpected non-error")
	}
}

func TestConverter_GeneratedConversionOverridden(t *testing.T) {
	type A struct{}
	type B struct{}
	c := NewConverter(nil)
	convertFn1 := func(in *A, out *B, s Scope) error {
		return nil
	}
	if err := c.RegisterUntypedConversionFunc(
		(*A)(nil), (*B)(nil),
		func(a, b interface{}, s Scope) error {
			return convertFn1(a.(*A), b.(*B), s)
		},
	); err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	convertFn2 := func(in *A, out *B, s Scope) error {
		return fmt.Errorf("generated function should be overridden")
	}
	if err := c.RegisterGeneratedUntypedConversionFunc(
		(*A)(nil), (*B)(nil),
		func(a, b interface{}, s Scope) error {
			return convertFn2(a.(*A), b.(*B), s)
		},
	); err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	a := A{}
	b := B{}
	if err := c.Convert(&a, &b, nil); err != nil {
		t.Errorf("%v", err)
	}
}
