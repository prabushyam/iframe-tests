// +build integration

package main

import (
	"testing"

	"github.com/prabushyam/iframe-tests/fixtures"
	"github.com/prabushyam/iframe-tests/framework"
)

func TestApiServer(t *testing.T) {
	f, err := framework.RunApiServer()
	if err != nil {
		t.Error(err)
	}
	defer framework.ShutdownApiServer()
	fixtures.CreateAthenzDomain(f.AthenzDomainClientset)
}
