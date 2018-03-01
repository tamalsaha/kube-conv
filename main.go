package main

import (
	"path/filepath"

	"github.com/appscode/kutil/meta"
	"github.com/appscode/go/log"
	"k8s.io/api/apps/v1beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"fmt"

)

func main() {
	masterURL := ""
	kubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube/config")

	config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
	if err != nil {
		log.Fatalf("Could not get Kubernetes config: %s", err)
	}

	kc := kubernetes.NewForConfigOrDie(config)

	dep_v1, err := kc.AppsV1().Deployments("kube-system").Get("pack-server", metav1.GetOptions{})
	if err != nil {
		log.Fatal(err)
	}
	dep_v1beta2 := &v1beta2.Deployment{}

	err = scheme.Scheme.Convert(dep_v1, dep_v1beta2, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(meta.MarshalToYAML(dep_v1beta2, v1beta2.SchemeGroupVersion))


}
