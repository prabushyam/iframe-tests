// Copyright 2019, Verizon Media Inc.
// Licensed under the terms of the 3-Clause BSD license. See LICENSE file in
// github.com/yahoo/k8s-athenz-istio-auth for terms.

// +build integration

package framework

import (
	"io/ioutil"
	"log"
	"net"
	"os"

	"github.com/coreos/etcd/embed"
	"github.com/prabushyam/iframe-tests/fixtures"
	athenzdomainclientset "github.com/yahoo/k8s-athenz-syncer/pkg/client/clientset/versioned"

	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/rest"
	"k8s.io/kubernetes/cmd/kube-apiserver/app"
	"k8s.io/kubernetes/cmd/kube-apiserver/app/options"
)

type Framework struct {
	AthenzDomainClientset *athenzdomainclientset.Clientset
	etcd                  *embed.Etcd
	stopCh                chan struct{}
}

// runEtcd will setup up the etcd configuration and run the etcd server
func runEtcd() (*embed.Etcd, error) {
	etcdDataDir, err := ioutil.TempDir(os.TempDir(), "integration_test_etcd_data")
	if err != nil {
		return nil, err
	}

	config := embed.NewConfig()
	config.Dir = etcdDataDir
	return embed.StartEtcd(config)
}

// runApiServer will setup the api configuration and run the api server
func runApiServer(certDir string) (*rest.Config, chan struct{}, error) {
	s := options.NewServerRunOptions()

	// TODO, remove the webhooks and api server certs
	s.InsecureServing.BindAddress = net.ParseIP("127.0.0.1")
	s.InsecureServing.BindPort = 9999
	s.Etcd.StorageConfig.Transport.ServerList = []string{"http://127.0.0.1:2379"}
	s.SecureServing.ServerCert.CertDirectory = certDir

	completedOptions, err := app.Complete(s)
	if err != nil {
		return nil, nil, err
	}

	if errs := completedOptions.Validate(); len(errs) != 0 {
		return nil, nil, errors.NewAggregate(errs)
	}

	stopCh := make(chan struct{})
	server, err := app.CreateServerChain(completedOptions, stopCh)
	if err != nil {
		return nil, nil, err
	}

	restConfig := &rest.Config{}
	restConfig.Host = "http://127.0.0.1:9999"
	return restConfig, stopCh, server.PrepareRun().NonBlockingRun(stopCh)
}

// Setup will run both etcd and api server together
func Setup() (*Framework, error) {
	etcd, err := runEtcd()
	if err != nil {
		return nil, err
	}

	restConfig, stopCh, err := runApiServer(etcd.Server.Cfg.DataDir)
	if err != nil {
		return nil, err
	}

	crdClientset, err := apiextensionsclient.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	err = fixtures.CreateCrds(crdClientset)
	if err != nil {
		return nil, err
	}

	athenzDomainClientset, err := athenzdomainclientset.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return &Framework{
		AthenzDomainClientset: athenzDomainClientset,
		etcd:                  etcd,
		stopCh:                stopCh,
	}, nil
}

// Teardown will request the api server to shutdown
func (f *Framework) Teardown() {
	err := os.RemoveAll(f.etcd.Server.Cfg.DataDir)
	if err != nil {
		log.Println(err)
	}

	close(f.stopCh)
	f.etcd.Close()
}
