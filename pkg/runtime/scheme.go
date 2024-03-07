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

package runtime

import (
	"fmt"
	"reflect"

	"apimachinery/pkg/conversion"
	"apimachinery/pkg/runtime/schema"
)

// Scheme defines methods for serializing and deserializing API objects, a type
// registry for converting group, version, and kind information to and from Go
// schemas, and mappings between Go schemas of different versions. A scheme is the
// foundation for a versioned API and versioned configuration over time.
//
// In a Scheme, a Type is a particular Go struct, a Version is a point-in-time
// identifier for a particular representation of that Type (typically backwards
// compatible), a Kind is the unique name for that Type within the Version, and a
// Group identifies a set of Versions, Kinds, and Types that evolve over time. An
// Unversioned Type is one that is not yet formally bound to a type and is promised
// to be backwards compatible (effectively a "v1" of a Type that does not expect
// to break in the future).
//
// Schemes are not expected to change at runtime and are only threadsafe after
// registration is complete.
type Scheme struct {
	// gvkToType allows one to figure out the go type of an object with
	// the given version and name.
	gvkToType map[schema.GroupVersionKind]reflect.Type

	// typeToGVK allows one to find metadata for a given go object.
	// The reflect.Type we index by should *not* be a pointer.
	typeToGVK map[reflect.Type][]schema.GroupVersionKind

	// converter stores all registered conversion functions. It also has
	// default converting behavior.
	converter *conversion.Converter

	// observedVersions keeps track of the order we've seen versions during type registration
	observedVersions []schema.GroupVersion

	// schemeName is the name of this scheme.  If you don't specify a name, the stack of the NewScheme caller will be used.
	// This is useful for error reporting to indicate the origin of the scheme.
	schemeName string
}

// NewScheme creates a new Scheme. This scheme is pluggable by default.
func NewScheme() *Scheme {
	s := &Scheme{
		gvkToType:  map[schema.GroupVersionKind]reflect.Type{},
		typeToGVK:  map[reflect.Type][]schema.GroupVersionKind{},
		schemeName: internalPackages[0],
	}
	s.converter = conversion.NewConverter(nil)
	return s
}

// KnownTypes returns the types known for the given version.
func (s *Scheme) KnownTypes(gv schema.GroupVersion) map[string]reflect.Type {
	types := make(map[string]reflect.Type)
	for gvk, t := range s.gvkToType {
		if gv != gvk.GroupVersion() {
			continue
		}

		types[gvk.Kind] = t
	}
	return types
}

// AddKnownTypes registers all types passed in 'types' as being members of version 'version'.
// All objects passed to types should be pointers to structs. The name that go reports for
// the struct becomes the "kind" field when encoding. Version may not be empty - use the
// APIVersionInternal constant if you have a type that does not have a formal version.
func (s *Scheme) AddKnownTypes(gv schema.GroupVersion, types ...Object) {
	s.addObservedVersion(gv)
	for _, obj := range types {
		t := reflect.TypeOf(obj)
		if t.Kind() != reflect.Pointer {
			panic("All types must be pointers to structs.")
		}
		t = t.Elem()
		s.AddKnownTypeWithName(gv.WithKind(t.Name()), obj)
	}
}

// AddKnownTypeWithName is like AddKnownTypes, but it lets you specify what this type should
// be encoded as. Useful for testing when you don't want to make multiple packages to define
// your structs. Version may not be empty - use the APIVersionInternal constant if you have a
// type that does not have a formal version.
func (s *Scheme) AddKnownTypeWithName(gvk schema.GroupVersionKind, obj Object) {
	s.addObservedVersion(gvk.GroupVersion())
	t := reflect.TypeOf(obj)
	if len(gvk.Version) == 0 {
		panic(fmt.Sprintf("version is required on all types: %s %v", gvk, t))
	}
	if t.Kind() != reflect.Pointer {
		panic("All types must be pointers to structs.")
	}
	t = t.Elem()
	if t.Kind() != reflect.Struct {
		panic("All types must be pointers to structs.")
	}

	if oldT, found := s.gvkToType[gvk]; found && oldT != t {
		panic(fmt.Sprintf("Double registration of different types for %v: old=%v.%v, new=%v.%v in scheme %q", gvk, oldT.PkgPath(), oldT.Name(), t.PkgPath(), t.Name(), s.schemeName))
	}

	s.gvkToType[gvk] = t

	for _, existingGvk := range s.typeToGVK[t] {
		if existingGvk == gvk {
			return
		}
	}
	s.typeToGVK[t] = append(s.typeToGVK[t], gvk)

	// if the type implements DeepCopyInto(<obj>), register a self-conversion
	if m := reflect.ValueOf(obj).MethodByName("DeepCopyInto"); m.IsValid() && m.Type().NumIn() == 1 && m.Type().NumOut() == 0 && m.Type().In(0) == reflect.TypeOf(obj) {
		if err := s.AddGeneratedConversionFunc(obj, obj, func(a, b interface{}, scope conversion.Scope) error {
			// copy a to b
			reflect.ValueOf(a).MethodByName("DeepCopyInto").Call([]reflect.Value{reflect.ValueOf(b)})
			// clear TypeMeta to match legacy reflective conversion
			b.(Object).GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{})
			return nil
		}); err != nil {
			panic(err)
		}
	}
}

// AddConversionFunc registers a function that converts between a and b by passing objects of those
// types to the provided function. The function *must* accept objects of a and b - this machinery will not enforce
// any other guarantee.
func (s *Scheme) AddConversionFunc(a, b interface{}, fn conversion.ConversionFunc) error {
	return s.converter.RegisterUntypedConversionFunc(a, b, fn)
}

// AddGeneratedConversionFunc registers a function that converts between a and b by passing objects of those
// types to the provided function. The function *must* accept objects of a and b - this machinery will not enforce
// any other guarantee.
func (s *Scheme) AddGeneratedConversionFunc(a, b interface{}, fn conversion.ConversionFunc) error {
	return s.converter.RegisterGeneratedUntypedConversionFunc(a, b, fn)
}

func (s *Scheme) addObservedVersion(version schema.GroupVersion) {
	if len(version.Version) == 0 || version.Version == APIVersionInternal {
		return
	}
	for _, observedVersion := range s.observedVersions {
		if observedVersion == version {
			return
		}
	}

	s.observedVersions = append(s.observedVersions, version)
}

// internalPackages are packages that ignored when creating a default reflector name. These packages are in the common
// call chains to NewReflector, so they'd be low entropy names for reflectors
var internalPackages = []string{"apimachinery/pkg/runtime/scheme.go"}
