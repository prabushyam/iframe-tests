/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"testing"

	"github.com/prabushyam/iframe-tests/framework"
	"time"
)

func TestApiServer(t *testing.T) {
	f, err := framework.RunApiServer()
	if err != nil {
		t.Error(err)
	}
	defer f.ShutdownApiServer()

	f.CreateCrd()
	time.Sleep(time.Second * 1)
	f.CreateCr()
	//time.Sleep(time.Second * 15)
}
