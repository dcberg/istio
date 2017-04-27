// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/golang/glog"
)

const (
	istioctlURL = "ISTIOCTL_URL"
)

var (
	remotePath = flag.String("istioctl_url", os.Getenv(istioctlURL), "URL to download istioctl")
)

// Istioctl gathers istioctl information
type Istioctl struct {
	remotePath string
	binaryPath string
	namespace  string
}

// NewIstioctl create a new istioctl by given temp dir
func NewIstioctl(tmpDir, namespace string) *Istioctl {
	return &Istioctl{
		remotePath: *remotePath,
		binaryPath: filepath.Join(tmpDir, "/istioctl"),
		namespace:  namespace,
	}
}

// DownloadIstioctl download Istioctl binary
func (i *Istioctl) DownloadIstioctl() error {
	var usr, err = user.Current()
	if err != nil {
		return err
	}
	homeDir := usr.HomeDir

	if err = HTTPDownload(i.binaryPath, i.remotePath+"/istioctl-linux"); err != nil {
		return err
	}
	err = os.Chmod(i.binaryPath, 0755) // #nosec
	if err != nil {
		return err
	}
	i.binaryPath = fmt.Sprintf("%s -c %s/.kube/config", i.binaryPath, homeDir)
	return nil
}

// KubeInject use istio kube-inject to create new yaml with a proxy as sidecar
func (i *Istioctl) KubeInject(yamlFile, svcName, yamlDir, proxyHub, proxyTag string) (string, error) {
	injectedYamlFile := filepath.Join(yamlDir, "injected-"+svcName+"-app.yaml")
	if _, err := Shell(fmt.Sprintf("%s kube-inject -f %s -o %s --hub %s --tag %s -n %s",
		i.binaryPath, yamlFile, injectedYamlFile, proxyHub, proxyTag, i.namespace)); err != nil {
		glog.Errorf("Kube-inject failed for service %s", svcName)
		return "", err
	}
	return injectedYamlFile, nil
}

// CreateRule create new rule(s)
func (i *Istioctl) CreateRule(rule string) error {
	_, err := Shell(fmt.Sprintf("%s -n %s create -f %s", i.binaryPath, i.namespace, rule))
	return err
}

// ReplaceRule replace rule(s)
func (i *Istioctl) ReplaceRule(rule string) error {
	_, err := Shell(fmt.Sprintf("%s -n %s replace -f %s", i.binaryPath, i.namespace, rule))
	return err
}

// DeleteRule Delete rule(s)
func (i *Istioctl) DeleteRule(rule string) error {
	_, err := Shell(fmt.Sprintf("%s -n %s delete -f %s", i.binaryPath, i.namespace, rule))
	return err
}