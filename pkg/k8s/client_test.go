/*
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

package k8s

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestNewK8sClient(t *testing.T) {
	// Test with GO_TEST_MODE=true
	os.Setenv("GO_TEST_MODE", "true")
	defer os.Unsetenv("GO_TEST_MODE")

	client, err := NewK8sClient()
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.cl)
	assert.NotNil(t, client.dCl)
	assert.NotNil(t, client.rCl)
}

func TestNewK8sClientDoOrDie(t *testing.T) {
	// Test with GO_TEST_MODE=true
	os.Setenv("GO_TEST_MODE", "true")
	defer os.Unsetenv("GO_TEST_MODE")

	client := NewK8sClientDoOrDie()
	assert.NotNil(t, client)
	assert.NotNil(t, client.cl)
	assert.NotNil(t, client.dCl)
	assert.NotNil(t, client.rCl)
}

func TestGetConfigMap(t *testing.T) {
	// Create a fake clientset
	clientset := fake.NewSimpleClientset(
		&v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-config",
				Namespace: "default",
			},
			Data: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
	)

	// Create a client
	client := &Client{
		cl: clientset,
	}

	// Test retrieving the config map
	cm := client.GetConfigMap(context.Background(), "default", "test-config")
	assert.NotNil(t, cm)
	assert.Equal(t, "test-config", cm.Name)
	assert.Equal(t, "default", cm.Namespace)
	assert.Equal(t, "value1", cm.Data["key1"])
	assert.Equal(t, "value2", cm.Data["key2"])

	// Test retrieving a non-existent config map
	cm = client.GetConfigMap(context.Background(), "default", "non-existent")
	assert.Nil(t, cm)
}

func TestClientInterface(t *testing.T) {
	// Create a fake clientset
	clientset := fake.NewSimpleClientset()

	// Create a client
	client := &Client{
		cl: clientset,
	}

	// Test retrieving the client interface
	cl := client.ClientInterface()
	assert.NotNil(t, cl)
	assert.Equal(t, clientset, cl)
}

func TestGetNamespace(t *testing.T) {
	// Create a fake clientset with a namespace
	clientset := fake.NewSimpleClientset(
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-namespace",
			},
		},
	)

	// Create a client
	client := &Client{
		cl: clientset,
	}

	// Test retrieving the namespace
	ns, err := client.GetNamespace(context.Background(), "test-namespace")
	assert.NoError(t, err)
	assert.NotNil(t, ns)
	assert.Equal(t, "test-namespace", ns.Name)

	// Test retrieving a non-existent namespace
	ns, err = client.GetNamespace(context.Background(), "non-existent")
	assert.Error(t, err)
	assert.Nil(t, ns)
}

func TestGetServiceAccount(t *testing.T) {
	// Create a fake clientset with a service account
	clientset := fake.NewSimpleClientset(
		&v1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-sa",
				Namespace: "default",
			},
		},
	)

	// Create a client
	client := &Client{
		cl: clientset,
	}

	// Test retrieving the service account
	sa := client.GetServiceAccount(context.Background(), "default", "test-sa")
	assert.NotNil(t, sa)
	assert.Equal(t, "test-sa", sa.Name)
	assert.Equal(t, "default", sa.Namespace)

	// Test retrieving a non-existent service account
	sa = client.GetServiceAccount(context.Background(), "default", "non-existent")
	assert.Nil(t, sa)
}

func TestNewK8sManagerClient(t *testing.T) {
	// Create a fake controller runtime client
	fakeClient := ctrlclient.NewClientBuilder().Build()

	// Create a client
	client := NewK8sManagerClient(fakeClient)
	assert.NotNil(t, client)
	assert.NotNil(t, client.rCl)
	assert.Equal(t, fakeClient, client.rCl)
}

func TestSetUpEventHandler(t *testing.T) {
	// Create a fake clientset
	clientset := fake.NewSimpleClientset()

	// Create a client
	client := &Client{
		cl: clientset,
	}

	// Test setting up the event handler
	eventRecorder := client.SetUpEventHandler(context.Background())
	assert.NotNil(t, eventRecorder)
}
