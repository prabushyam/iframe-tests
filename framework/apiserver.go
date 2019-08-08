package framework

import (
	"github.com/coreos/etcd/embed"
	"io/ioutil"
	"net"
	"os"

	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/rest"
	"k8s.io/kubernetes/cmd/kube-apiserver/app"
	"k8s.io/kubernetes/cmd/kube-apiserver/app/options"
)

type Framework struct {
}

// runEtcd will setup up the etcd configuration and run the etcd server
func runEtcd() error {
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
func runApiServer() (*rest.Config, error) {
	s := options.NewServerRunOptions()

	listener, err := net.Listen("tcp4", "127.0.0.1:9999")
	if err != nil {
		return nil, err
	}

	// TODO, remove the webhooks
	s.InsecureServing.BindAddress = net.ParseIP("127.0.0.1")
	s.InsecureServing.Listener = listener
	s.Etcd.StorageConfig.Transport.ServerList = []string{"http://127.0.0.1:2379"}
	s.SecureServing.ServerCert.CertDirectory = "/tmp"

	completedOptions, err := app.Complete(s)
	if err != nil {
		return nil, err
	}

	if errs := completedOptions.Validate(); len(errs) != 0 {
		return nil, utilerrors.NewAggregate(errs)
	}

	stopCh := genericapiserver.SetupSignalHandler()
	server, err := app.CreateServerChain(completedOptions, stopCh)
	if err != nil {
		return nil, err
	}

	err = server.PrepareRun().NonBlockingRun(stopCh)

	restConfig := &rest.Config{}
	restConfig.Host = "http://127.0.0.1:9999"
	return restConfig, err
}

// RunApiServer will run both etcd and api server together
func RunApiServer() (*rest.Config, error) {
	err := runEtcd()
	if err != nil {
		return nil, err
	}

	return runApiServer()
}

// ShutdownApiServer will request the api server to shutdown
func ShutdownApiServer() {
	genericapiserver.RequestShutdown()
}
