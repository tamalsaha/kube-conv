package main

import (
	"github.com/appscode/go/log"
	"k8s.io/api/apps/v1"
	"k8s.io/api/apps/v1beta1"
	_ "k8s.io/api/extensions/v1beta1"
	// "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	_ "k8s.io/kubernetes/pkg/apis/apps/install"
	_ "k8s.io/kubernetes/pkg/apis/batch/install"
	_ "k8s.io/kubernetes/pkg/apis/core/install"
	_ "k8s.io/kubernetes/pkg/apis/extensions/install"
	"k8s.io/kubernetes/pkg/api/legacyscheme"


	"k8s.io/kubernetes/pkg/apis/apps"
)

func main() {
	v1Obj := &v1.StatefulSet{
		Status: v1.StatefulSetStatus{
			ObservedGeneration: 1,
		},
	}

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

	//transform := func(in *v1beta1.Deployment) *v1beta1.Deployment { return in}
	//
	//masterURL := ""
	//kubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube/config")
	//
	//config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
	//if err != nil {
	//	log.Fatalf("Could not get Kubernetes config: %s", err)
	//}
	//
	//kc := kubernetes.NewForConfigOrDie(config)
	//
	//in_v1, err := kc.AppsV1().Deployments("kube-system").Get("pack-server", metav1.GetOptions{})
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//// v1 -> v1beta1
	//in_v1beta1 := &v1beta1.Deployment{}
	//err = scheme.Scheme.Convert(in_v1, in_v1beta1, nil)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//mod_v1beta1 := transform(in_v1beta1.DeepCopy())
	//
	//mod_v1 := &v1.Deployment{}
	//err = scheme.Scheme.Convert(mod_v1beta1, mod_v1, nil)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//meta.CreateStrategicPatch(in_v1, mod_v1)
	//
	//
	//data_v1b1, err := meta.MarshalToYAML(in_v1beta1, v1beta2.SchemeGroupVersion)
	//fmt.Println(string(data_v1b1))
}
