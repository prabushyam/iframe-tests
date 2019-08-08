package framework

import (
	"github.com/coreos/etcd/embed"
	"io/ioutil"
	"net"
	"os"

	athenzClientset "github.com/yahoo/k8s-athenz-syncer/pkg/client/clientset/versioned"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/rest"
	"k8s.io/kubernetes/cmd/kube-apiserver/app"
	"k8s.io/kubernetes/cmd/kube-apiserver/app/options"
)

type Framework struct {
	restConfig   *rest.Config
	crdClientset *apiextensionsclient.Clientset
	crClientset  *athenzClientset.Clientset
}

// runEtcd will setup up the etcd configuration and run the etcd server
func (f *Framework) runEtcd() error {
	etcdDataDir, err := ioutil.TempDir(os.TempDir(), "integration_test_etcd_data")
	if err != nil {
		return err
	}

	config := embed.NewConfig()
	config.Dir = etcdDataDir
	_, err = embed.StartEtcd(config)
	return err
}

// runApiServer will setup the api configuration and run the api server
func (f *Framework) runApiServer() error {
	s := options.NewServerRunOptions()

	listener, err := net.Listen("tcp4", "127.0.0.1:9999")
	if err != nil {
		return err
	}

	// TODO, remove the webhooks
	s.InsecureServing.BindAddress = net.ParseIP("127.0.0.1")
	s.InsecureServing.Listener = listener
	s.Etcd.StorageConfig.Transport.ServerList = []string{"http://127.0.0.1:2379"}
	s.SecureServing.ServerCert.CertDirectory = "/tmp"

	completedOptions, err := app.Complete(s)
	if err != nil {
		return err
	}

	if errs := completedOptions.Validate(); len(errs) != 0 {
		return utilerrors.NewAggregate(errs)
	}

	stopCh := genericapiserver.SetupSignalHandler()
	server, err := app.CreateServerChain(completedOptions, stopCh)
	if err != nil {
		return err
	}

	err = server.PrepareRun().NonBlockingRun(stopCh)

	restConfig := &rest.Config{}
	restConfig.Host = "http://127.0.0.1:9999"
	f.restConfig = restConfig
	return err
}

func (f *Framework) createClients() error {
	rClientset, err := apiextensionsclient.NewForConfig(f.restConfig)
	if err != nil {
		return err
	}
	f.crdClientset = rClientset

	versiondClient, err := athenzClientset.NewForConfig(f.restConfig)
	if err != nil {
		return err
	}
	f.crClientset = versiondClient
	return nil
}

// RunApiServer will run both etcd and api server together
func RunApiServer() (*Framework, error) {
	f := &Framework{}
	err := f.runEtcd()
	if err != nil {
		return f, err
	}

	err = f.runApiServer()
	if err != nil {
		return f, err
	}

	err = f.createClients()
	if err != nil {
		return f, err
	}

	return f, nil
}

// ShutdownApiServer will request the api server to shutdown
func (f *Framework) ShutdownApiServer() {
	genericapiserver.RequestShutdown()
}
