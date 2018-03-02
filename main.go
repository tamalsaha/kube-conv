package main

import (
	"path/filepath"

	"github.com/appscode/kutil/meta"
	"github.com/appscode/go/log"
	"k8s.io/api/apps/v1beta2"
	"k8s.io/api/apps/v1beta1"
	_ "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"fmt"

	"k8s.io/api/apps/v1"
)

func main() {
	transform := func(in *v1beta1.Deployment) *v1beta1.Deployment { return in}


	masterURL := ""
	kubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube/config")

	config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
	if err != nil {
		log.Fatalf("Could not get Kubernetes config: %s", err)
	}

	kc := kubernetes.NewForConfigOrDie(config)

	in_v1, err := kc.AppsV1().Deployments("kube-system").Get("pack-server", metav1.GetOptions{})
	if err != nil {
		log.Fatal(err)
	}

	// v1 -> v1beta1
	in_v1beta1 := &v1beta1.Deployment{}
	err = scheme.Scheme.Convert(in_v1, in_v1beta1, nil)
	if err != nil {
		log.Fatal(err)
	}

	mod_v1beta1 := transform(in_v1beta1.DeepCopy())

	mod_v1 := &v1.Deployment{}
	err = scheme.Scheme.Convert(mod_v1beta1, mod_v1, nil)
	if err != nil {
		log.Fatal(err)
	}

	meta.CreateStrategicPatch(in_v1, mod_v1)


	data_v1b1, err := meta.MarshalToYAML(in_v1beta1, v1beta2.SchemeGroupVersion)
	fmt.Println(string(data_v1b1))


}
