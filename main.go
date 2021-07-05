package main

import (
	"crypto/rand"
	"crypto/rsa"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/client-go/util/homedir"

	"golang.org/x/crypto/ssh"
)

const (
	bitSize = 4096
)

func getClientSet() (*kubernetes.Clientset, *rest.Config) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return clientset, config
}

func sshPortforward(clientset *kubernetes.Clientset) {
	stopCh := make(<-chan struct{})
	readyCh := make(chan struct{})

	reqURL := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace("7db81398-eacc-404c-a6dc-2d70d32659bc").
		Name("pod-0gpu-0805-172608").
		SubResource("portforward").URL()

	transport, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		panic(err)
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, reqURL)
	fw, err := portforward.New(dialer, []string{"8080:80"}, stopCh, readyCh, os.Stdout, os.Stdout)
	if err != nil {
		panic(err)
	}
	if err := fw.ForwardPorts(); err != nil {
		panic(err)
	}
}

//func cleanUp() {
//
//}

func generatePrivateKey() (*rsa.PrivateKey, error) {
	// Private Key generation
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, err
	}

	// Validate Private Key
	err = privateKey.Validate()
	if err != nil {
		return nil, err
	}

	log.Println("Private Key generated")
	return privateKey, nil
}

func generatePublicKey(privatekey *rsa.PrivateKey) ([]byte, error) {
	publicRsaKey, err := ssh.NewPublicKey(privatekey)
	if err != nil {
		return nil, err
	}

	pubKeyBytes := ssh.MarshalAuthorizedKey(publicRsaKey)

	log.Println("Public key generated")
	return pubKeyBytes, nil
}

func createSecret(pass string, cr *api.PredixMessageQueue) {
	kubeClient := k8s{}
	secret := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "apps/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rabbitmq-secret",
			Namespace: cr.Namespace,
		},
		Data: map[string][]byte{
			"rabbitmq-secret": []byte(pass),
		},
		Type: "Opaque",
	}
	secretOut, err := kubeClient.clientset.CoreV1().Secrets(cr.Namespace).Create(secret)
	if err != nil {
		fmt.Println("ERROR:", err)
	}
}

func main() {

	id_rsa, err := generatePrivateKey()
	if err != nil {
		panic(err)
	}

	id_rsa_pub, err := generatePublicKey(id_rsa)
	if err != nil {
		panic(err)
	}

	go sshPortforward()

	for {
		fmt.Printf("test\n")

		time.Sleep(10 * time.Second)
	}
}
