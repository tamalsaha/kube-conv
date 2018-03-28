package main

import (
	"bytes"
	"fmt"
	"io"

	api "github.com/appscode/stash/apis/stash/v1alpha1"
	"github.com/appscode/stash/client/clientset/versioned/scheme"
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

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kubernetes/pkg/apis/apps"
)

var ss = `{
  "metadata": {
    "creationTimestamp": null
  },
  "status": {
    "observedGeneration": 1,
    "replicas": 0
  }
}`

var rs = `apiVersion: stash.appscode.com/v1alpha1
kind: Restic
metadata:
  name: stash-demo
  namespace: default
spec:
  selector:
    matchLabels:
      app: stash-demo
  # type: offline
  fileGroups:
  - path: /source/data
    retentionPolicyName: 'keep-last-5'
  backend:
    local:
      mountPath: /safe/data
      hostPath:
        path: /data/stash-test/restic-repo
    storageSecretName: local-secret
  schedule: '@every 1m'
 # paused: false
  volumeMounts:
  - mountPath: /source/data
    name: source-data
  retentionPolicies:
  - name: 'keep-last-5'
    keepLast: 5
    prune: true`

func transform(obj runtime.Object) (runtime.Object, error) {
	var n int32 = 2
	r := obj.(*v1beta1.StatefulSet)
	r.Spec.Replicas = &n
	return obj, nil
}

func xyz() runtime.Codec {
	mediaType := "application/yaml"
	info, ok := runtime.SerializerInfoForMediaType(legacyscheme.Codecs.SupportedMediaTypes(), mediaType)
	if !ok {
		panic("unsupported media type " + mediaType)
	}
	return info.Serializer
}

var Serializer = xyz()

type VersionedCodec struct {
	scheme        *runtime.Scheme
	encodeVersion runtime.GroupVersioner
	decodeVersion runtime.GroupVersioner
}

func (c VersionedCodec) Encode(obj runtime.Object, w io.Writer) error {
	var out runtime.Object
	if c.encodeVersion == c.decodeVersion {
		out = obj
	} else {
		internal, err := c.scheme.UnsafeConvertToVersion(obj, runtime.InternalGroupVersioner)
		if err != nil {
			return err
		}

		out, err = c.scheme.UnsafeConvertToVersion(internal, c.encodeVersion)
		if err != nil {
			return err
		}
	}
	c.scheme.Default(out)

	return Serializer.Encode(out, w)
}

func (c VersionedCodec) Decode(data []byte, gvk *schema.GroupVersionKind, _ runtime.Object) (runtime.Object, *schema.GroupVersionKind, error) {
	in, gvk, err := Serializer.Decode(data, gvk, nil)
	if err != nil {
		return nil, gvk, err
	}
	if gvk.GroupVersion() != c.encodeVersion {
		return nil, gvk, fmt.Errorf("data expected to be of version %s, found %s", c.encodeVersion, gvk)
	}
	c.scheme.Default(in)
	in.GetObjectKind().SetGroupVersionKind(*gvk)

	if c.encodeVersion == c.decodeVersion {
		return in, gvk, err
	}

	internal, err := c.scheme.UnsafeConvertToVersion(in, runtime.InternalGroupVersioner)
	if err != nil {
		return nil, gvk, err
	}

	out, err := c.scheme.UnsafeConvertToVersion(internal, c.decodeVersion)
	return out, gvk, err
}

func main() {
	scheme.AddToScheme(legacyscheme.Scheme)

	raw := []byte(rs)
	gvk := api.SchemeGroupVersion.WithKind("Restic")

	cur, _, err := Serializer.Decode(raw, &gvk, nil)
	if err != nil {
		log.Fatal(err)
	}
	legacyscheme.Scheme.Default(cur)
	r := cur.(*api.Restic)
	fmt.Println("|" + r.Spec.Backend.Local.HostPath.String() + "|")

	var buf bytes.Buffer
	err = Serializer.Encode(cur, &buf)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(buf.String())
}

func main78() {
	scheme.AddToScheme(legacyscheme.Scheme)

	raw := []byte(ss)
	gvk := v1.SchemeGroupVersion.WithKind("StatefulSet")

	cur, k2, err := Serializer.Decode(raw, &gvk, nil)
	if err != nil {
		log.Fatal(err)
	}
	legacyscheme.Scheme.Default(cur)
	cur.GetObjectKind().SetGroupVersionKind(*k2)
	fmt.Println(k2)

	var buf bytes.Buffer
	err = Serializer.Encode(cur, &buf)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(buf.String())
}

func main6() {
	scheme.AddToScheme(legacyscheme.Scheme)
	//legacyscheme.Registry.AllPreferredGroupVersions()
	//legacyscheme.Registry.EnabledVersions()

	raw := []byte(rs)
	gvk := api.SchemeGroupVersion.WithKind("Restic")

	c := VersionedCodec{
		scheme:        legacyscheme.Scheme,
		encodeVersion: api.SchemeGroupVersion,
		decodeVersion: api.SchemeGroupVersion,
	}
	cur, gvk2, err := c.Decode(raw, &gvk, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(gvk2)

	var buf bytes.Buffer
	err = c.Encode(cur, &buf)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(buf.String())
}

func main4() {
	scheme.AddToScheme(legacyscheme.Scheme)
	//legacyscheme.Registry.AllPreferredGroupVersions()
	//legacyscheme.Registry.EnabledVersions()

	raw := []byte(ss)

	c := VersionedCodec{
		encodeVersion: v1.SchemeGroupVersion,
		decodeVersion: v1beta1.SchemeGroupVersion,
	}
	cur, _, err := c.Decode(raw, nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println(gvk)

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

	srcObj, srcGVK, err := Serializer.Decode(raw, nil, nil)
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
	err = Serializer.Encode(srcMod, &mod)
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
