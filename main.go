package main

import (
	"k8s.io/apimachinery/pkg/runtime"
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
	"io"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kubernetes/pkg/apis/apps"
)

var ss = `{
  "apiVersion": "apps/v1",
  "kind": "StatefulSet",
  "metadata": {
    "creationTimestamp": null
  },
  "spec": {
    "selector": null,
    "serviceName": "",
    "template": {
      "metadata": {
        "creationTimestamp": null
      },
      "spec": {
        "containers": null
      }
    },
    "updateStrategy": {}
  },
  "status": {
    "observedGeneration": 1,
    "replicas": 0
  }
}`

func transform(obj runtime.Object) (runtime.Object, error) {
	var n int32 = 2
	r := obj.(*v1beta1.StatefulSet)
	r.Spec.Replicas = &n
	return obj, nil
}

func xyz() runtime.Codec {
	mediaType := "application/json"
	info, ok := runtime.SerializerInfoForMediaType(legacyscheme.Codecs.SupportedMediaTypes(), mediaType)
	if !ok {
		panic("unsupported media type " + mediaType)
	}
	return info.Serializer
}

var Codec = xyz()

type VersionedCodec struct {
	encodeVersion runtime.GroupVersioner
	decodeVersion runtime.GroupVersioner
}

func (c VersionedCodec) Encode(obj runtime.Object, w io.Writer) error {
	internal, err := legacyscheme.Scheme.UnsafeConvertToVersion(obj, runtime.InternalGroupVersioner)
	if err != nil {
		return err
	}

	out, err := legacyscheme.Scheme.UnsafeConvertToVersion(internal, c.encodeVersion)
	if err != nil {
		return err
	}
	legacyscheme.Scheme.Default(out)

	return Codec.Encode(out, w)
}

func (c VersionedCodec) Decode(data []byte, _ *schema.GroupVersionKind, _ runtime.Object) (runtime.Object, *schema.GroupVersionKind, error) {
	in, gvk, err := Codec.Decode(data, nil, nil)
	if err != nil {
		return nil, gvk, err
	}
	legacyscheme.Scheme.Default(in)

	internal, err := legacyscheme.Scheme.UnsafeConvertToVersion(in, runtime.InternalGroupVersioner)
	if err != nil {
		return nil, gvk, err
	}

	out, err := legacyscheme.Scheme.UnsafeConvertToVersion(internal, c.decodeVersion)
	return out, gvk, err
}

func main() {
	raw := []byte(ss)

	c := VersionedCodec{
		encodeVersion: v1.SchemeGroupVersion,
		decodeVersion: v1beta1.SchemeGroupVersion,
	}
	cur, gvk, err := c.Decode(raw, nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(gvk)

	mod, err := transform(cur)
	if err != nil {
		log.Fatal(err)
	}

	var buf bytes.Buffer
	err = c.Encode(mod, &buf)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(buf.String())
}

func main3() {
	// legacyscheme.Codecs.CodecForVersions()

	raw := []byte(ss)

	srcObj, srcGVK, err := Codec.Decode(raw, nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(srcGVK)
	legacyscheme.Scheme.Default(srcObj)

	srcInternal, err := legacyscheme.Scheme.UnsafeConvertToVersion(srcObj, runtime.InternalGroupVersioner)
	if err != nil {
		log.Fatal(err)
	}

	dstObj, err := legacyscheme.Scheme.UnsafeConvertToVersion(srcInternal, v1beta1.SchemeGroupVersion)
	if err != nil {
		log.Fatal(err)
	}

	dstMod, err := transform(dstObj)
	if err != nil {
		log.Fatal(err)
	}

	dstModInternal, err := legacyscheme.Scheme.UnsafeConvertToVersion(dstMod, runtime.InternalGroupVersioner)
	if err != nil {
		log.Fatal(err)
	}

	srcMod, err := legacyscheme.Scheme.UnsafeConvertToVersion(dstModInternal, v1.SchemeGroupVersion)
	if err != nil {
		log.Fatal(err)
	}
	legacyscheme.Scheme.Default(srcMod)

	var mod bytes.Buffer
	err = Codec.Encode(srcMod, &mod)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(mod.String())

	//codec := legacyscheme.Codecs.CodecForVersions(Codec, Codec, nil, nil)
	//obj, gvk, err := codec.Decode(raw, nil,  nil)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Println(gvk)

	//codec := legacyscheme.Codecs.CodecForVersions(info.Serializer, info.Serializer, v1beta1.SchemeGroupVersion, nil)
	//v1Obj := &v1.StatefulSet{
	//	Status: v1.StatefulSetStatus{
	//		ObservedGeneration: 1,
	//	},
	//}
	//var b1 bytes.Buffer
	//err := codec.Encode(v1Obj, &b1)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Println(b1.String())

	// legacyscheme.Codecs.LegacyCodec(legacyscheme.Registry.EnabledVersions()...)
}

func main2() {
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

type GroupVersions struct{}

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
