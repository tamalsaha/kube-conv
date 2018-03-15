package main

import (
	// batchv1 "k8s.io/api/batch/v1"
	// batchv1beta1 "k8s.io/api/batch/v1beta1"
	"github.com/appscode/go/log"
	"k8s.io/api/apps/v1"
	"k8s.io/api/apps/v1beta1"
	_ "k8s.io/api/extensions/v1beta1"
	// "k8s.io/client-go/kubernetes/scheme"
	// "k8s.io/api/apps/v1beta2"
	// core "k8s.io/api/core/v1"
	// extensions "k8s.io/api/extensions/v1beta1"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	// "k8s.io/kubernetes/pkg/apis/apps"
	_ "k8s.io/kubernetes/pkg/apis/apps/install"
	_ "k8s.io/kubernetes/pkg/apis/batch/install"
	_ "k8s.io/kubernetes/pkg/apis/core/install"
	_ "k8s.io/kubernetes/pkg/apis/extensions/install"
	//	"k8s.io/apimachinery/pkg/runtime"
	"bytes"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kubernetes/pkg/apis/apps"
)

func main() {
	v1Obj := &v1.StatefulSet{
		Status: v1.StatefulSetStatus{
			ObservedGeneration: 1,
		},
	}

	/*
	// working example:

	// src -> internal -> dst
	internalObj := &apps.StatefulSet{}
	err := legacyscheme.Scheme.Convert(v1Obj, internalObj, nil)
	if err != nil {
		log.Fatal(err)
	}

	v1beta1Obj := &v1beta1.StatefulSet{}
	err = legacyscheme.Scheme.Convert(internalObj, v1beta1Obj, nil)
	if err != nil {
		log.Fatal(err)
	}
	*/

	var err error

	// apimachinery/pkg/runtime/serializer/versioning/versioning.go
	// Thanks for @sttts
	codec := legacyscheme.Codecs.LegacyCodec(legacyscheme.Registry.EnabledVersions()...)

	var v1Buf bytes.Buffer
	err = codec.Encode(v1Obj, &v1Buf)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(v1Buf.String())

	gvk := v1beta1.SchemeGroupVersion.WithKind("StatefulSet")

	internalObj := &apps.StatefulSet{}
	o1, o2, err := codec.Decode(v1Buf.Bytes(), &gvk, internalObj)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(o1)
	fmt.Println(o2)

	v1beta1Obj := &v1beta1.StatefulSet{}
	// PANIC !!!
	// panic: reflect: call of reflect.Value.IsNil on int64 Value
	o1, o2, err = codec.Decode(v1Buf.Bytes(), &gvk, v1beta1Obj)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(o1)
	fmt.Println(o2)

	//d2 := legacyscheme.Codecs.DecoderToVersion(codec, GroupVersions{})
	//v1beta1Obj := &v1beta1.StatefulSet{}
	//o1, o2, err := d2.Decode(v1Buf.Bytes(), &gvk, v1beta1Obj)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Println(o1)
	//fmt.Println(o2)

	//decoder := legacyscheme.Codecs.DecoderToVersion(codec, v1beta1.SchemeGroupVersion)
	//v1beta1Obj := &v1beta1.StatefulSet{}
	//o1, o2, err := decoder.Decode(v1Buf.Bytes(), &gvk, v1beta1Obj)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Println(o1)
	//fmt.Println(o2)

	//v1beta1Obj := &v1beta1.StatefulSet{}
	//err = legacyscheme.Scheme.Convert(v1Obj, v1beta1Obj, runtime.InternalGroupVersioner)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//

	//// src -> internal -> dst
	//internalObj := &apps.StatefulSet{}
	//err := legacyscheme.Scheme.Convert(v1Obj, internalObj, nil)
	//if err != nil {
	//	log.Fatal(err)
	//}
}

type GroupVersions struct {}

// KindForGroupVersionKinds identifies the preferred GroupVersionKind out of a list. It returns ok false
// if none of the options match the group.
func (gvs GroupVersions) KindForGroupVersionKinds(kinds []schema.GroupVersionKind) (schema.GroupVersionKind, bool) {
	for _, kind := range kinds {
		if kind.Version == "v1beta1" {
			return kind, true
		}
	}
	for _, kind := range kinds {
		return schema.GroupVersionKind{Group: kind.Group, Version: "v1beta1", Kind: kind.Kind}, true
	}
	return schema.GroupVersionKind{}, false
}
