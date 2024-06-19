package defaults

import (
	"testing"

	"github.com/Baiyuani/kubectl-neatx/pkg/testutil"
)

func TestComputeDefault(t *testing.T) {
	cases := []struct {
		title  string
		path   string
		data   string
		expect string
	}{
		{
			title: "PullPolicyAlways",
			path:  "spec.containers.0.imagePullPolicy",
			data: `{
				"apiVersion": "v1",
				"kind": "Pod",
				"metadata": {
					"name": "myapp",
					"namespace": "default"
				},
				"spec": {
					"containers": [
						{
							"image": "foo",
							"name": "myapp"
						}
					]
				}
			}`,
			expect: "Always",
		},
		{
			title: "PullPolicyIfNotPresent",
			path:  "spec.containers.0.imagePullPolicy",
			data: `{
				"apiVersion": "v1",
				"kind": "Pod",
				"metadata": {
					"name": "myapp",
					"namespace": "default"
				},
				"spec": {
					"containers": [
						{
							"image": "foo:bar",
							"name": "myapp"
						}
					]
				}
			}`,
			expect: "IfNotPresent",
		},
		{
			title: "RestartPolicy",
			path:  "spec.restartPolicy",
			data: `{
				"apiVersion": "v1",
				"kind": "Pod",
				"metadata": {
					"name": "myapp",
					"namespace": "default"
				},
				"spec": {
					"containers": [
						{
							"image": "foo:bar",
							"name": "myapp"
						}
					]
				}
			}`,
			expect: "Always",
		},
		{
			title: "TerminationMessagePath",
			path:  "spec.containers.0.terminationMessagePath",
			data: `{
				"apiVersion": "v1",
				"kind": "Pod",
				"metadata": {
					"name": "myapp",
					"namespace": "default"
				},
				"spec": {
					"containers": [
						{
							"image": "foo:bar",
							"name": "myapp"
						}
					]
				}
			}`,
			expect: "/dev/termination-log",
		},
	}
	for _, c := range cases {
		res, err := computeDefault(c.path, c.data)
		if err != nil {
			t.Errorf("error in computeDefault for case '%s': %v", c.title, err)
		}
		if res != c.expect {
			t.Errorf("test case '%s' failed. want: '%s' have: '%s'", c.title, c.expect, res)
		}
	}
}

func TestIsDefault(t *testing.T) {
	cases := []struct {
		title  string
		path   string
		value  interface{}
		object string
		expect bool
	}{
		{
			title: "PullPolicyAlways",
			path:  "spec.containers.0.imagePullPolicy",
			object: `{
				"apiVersion": "v1",
				"kind": "Pod",
				"metadata": {
					"name": "myapp",
					"namespace": "default"
				},
				"spec": {
					"containers": [
						{
							"image": "foo",
							"name": "myapp"
						}
					]
				}
			}`,
			value:  "Always",
			expect: true,
		},
		{
			title: "PullPolicyIfNotPresent",
			path:  "spec.containers.0.imagePullPolicy",
			object: `{
				"apiVersion": "v1",
				"kind": "Pod",
				"metadata": {
					"name": "myapp",
					"namespace": "default"
				},
				"spec": {
					"containers": [
						{
							"image": "foo:bar",
							"name": "myapp"
						}
					]
				}
			}`,
			value:  "IfNotPresent",
			expect: true,
		},
		{
			title: "RestartPolicy",
			path:  "spec.restartPolicy",
			object: `{
				"apiVersion": "v1",
				"kind": "Pod",
				"metadata": {
					"name": "myapp",
					"namespace": "default"
				},
				"spec": {
					"containers": [
						{
							"image": "foo:bar",
							"name": "myapp"
						}
					]
				}
			}`,
			value:  "Always",
			expect: true,
		},
		{
			title: "TerminationMessagePath",
			path:  "spec.containers.0.terminationMessagePath",
			object: `{
				"apiVersion": "v1",
				"kind": "Pod",
				"metadata": {
					"name": "myapp",
					"namespace": "default"
				},
				"spec": {
					"containers": [
						{
							"image": "foo:bar",
							"name": "myapp"
						}
					]
				}
			}`,
			value:  "/dev/termination-log",
			expect: true,
		},
	}
	for _, c := range cases {
		res, err := isDefault(c.path, c.value, c.object)
		if err != nil {
			t.Errorf("error in isDefault for case '%s': %v", c.title, err)
		}
		if res != c.expect {
			t.Errorf("test case '%s' failed. want: '%v' have: '%v'", c.title, c.expect, res)
		}
	}
}

func TestNeatDefault(t *testing.T) {
	cases := []struct {
		title  string
		data   string
		expect string
	}{
		{
			title: "PullPolicyAlways",
			data: `{
				"apiVersion": "v1",
				"kind": "Pod",
				"metadata": {
					"name": "myapp",
					"namespace": "default"
				},
				"spec": {
					"containers": [
						{
							"image": "foo",
							"imagePullPolicy": "Always",
							"name": "myapp"
						}
					]
				}
			}`,
			expect: `{
				"apiVersion": "v1",
				"kind": "Pod",
				"metadata": {
					"name": "myapp",
					"namespace": "default"
				},
				"spec": {
					"containers": [
						{
							"image": "foo",
							"name": "myapp"
						}
					]
				}
			}`,
		},
		{
			title: "PullPolicyIfNotPresent",
			data: `{
				"apiVersion": "v1",
				"kind": "Pod",
				"metadata": {
					"name": "myapp",
					"namespace": "default"
				},
				"spec": {
					"containers": [
						{
							"image": "foo:bar",
							"imagePullPolicy": "IfNotPresent",
							"name": "myapp"
						}
					]
				}
			}`,
			expect: `{
				"apiVersion": "v1",
				"kind": "Pod",
				"metadata": {
					"name": "myapp",
					"namespace": "default"
				},
				"spec": {
					"containers": [
						{
							"image": "foo:bar",
							"name": "myapp"
						}
					]
				}
			}`,
		},
		{
			title: "RestartPolicy",
			data: `{
				"apiVersion": "v1",
				"kind": "Pod",
				"metadata": {
					"name": "myapp",
					"namespace": "default"
				},
				"spec": {
					"restartPolicy": "Always",
					"containers": [
						{
							"image": "foo:bar",
							"name": "myapp"
						}
					]
				}
			}`,
			expect: `{
				"apiVersion": "v1",
				"kind": "Pod",
				"metadata": {
					"name": "myapp",
					"namespace": "default"
				},
				"spec": {
					"containers": [
						{
							"image": "foo:bar",
							"name": "myapp"
						}
					]
				}
			}`,
		},
		{
			title: "TerminationMessagePath",
			data: `{
				"apiVersion": "v1",
				"kind": "Pod",
				"metadata": {
					"name": "myapp",
					"namespace": "default"
				},
				"spec": {
					"containers": [
						{
							"terminationMessagePath": "/dev/termination-log",
							"image": "foo:bar",
							"name": "myapp"
						}
					]
				}
			}`,
			expect: `{
				"apiVersion": "v1",
				"kind": "Pod",
				"metadata": {
					"name": "myapp",
					"namespace": "default"
				},
				"spec": {
					"containers": [
						{
							"image": "foo:bar",
							"name": "myapp"
						}
					]
				}
			}`,
		},
		{
			title: "CRD",
			data: `{
				"apiVersion": "networking.istio.io/v1alpha3",
				"kind": "DestinationRule",
				"metadata": {
					"annotations": {
						"kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"networking.istio.io/v1alpha3\",\"kind\":\"DestinationRule\",\"metadata\":{\"annotations\":{},\"name\":\"default\",\"namespace\":\"one\"},\"spec\":{\"host\":\"*.one.svc.cluster.local\",\"trafficPolicy\":{\"tls\":{\"mode\":\"ISTIO_MUTUAL\"}}}}\n"
					},
					"creationTimestamp": "2019-11-06T20:14:07Z",
					"generation": 1,
					"name": "default",
					"namespace": "one",
					"resourceVersion": "314732",
					"selfLink": "/apis/networking.istio.io/v1alpha3/namespaces/one/destinationrules/default",
					"uid": "fca04858-00d1-11ea-84b3-025000000001"
				},
				"spec": {
					"host": "*.one.svc.cluster.local",
					"trafficPolicy": {
						"tls": {
							"mode": "ISTIO_MUTUAL"
						}
					}
				}
			}`,
			expect: `{
				"apiVersion": "networking.istio.io/v1alpha3",
				"kind": "DestinationRule",
				"metadata": {
					"annotations": {
						"kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"networking.istio.io/v1alpha3\",\"kind\":\"DestinationRule\",\"metadata\":{\"annotations\":{},\"name\":\"default\",\"namespace\":\"one\"},\"spec\":{\"host\":\"*.one.svc.cluster.local\",\"trafficPolicy\":{\"tls\":{\"mode\":\"ISTIO_MUTUAL\"}}}}\n"
					},
					"creationTimestamp": "2019-11-06T20:14:07Z",
					"generation": 1,
					"name": "default",
					"namespace": "one",
					"resourceVersion": "314732",
					"selfLink": "/apis/networking.istio.io/v1alpha3/namespaces/one/destinationrules/default",
					"uid": "fca04858-00d1-11ea-84b3-025000000001"
				},
				"spec": {
					"host": "*.one.svc.cluster.local",
					"trafficPolicy": {
						"tls": {
							"mode": "ISTIO_MUTUAL"
						}
					}
				}
			}`,
		},
	}
	for _, c := range cases {
		resJSON, err := NeatDefaults(c.data)
		if err != nil {
			t.Errorf("error in neatDefaults for case '%s': %v", c.title, err)
			continue
		}
		equal, err := testutil.JSONEqual(resJSON, c.expect)
		if err != nil {
			t.Errorf("error in JSONEqual for case '%s': %v", c.title, err)
			continue
		}
		if !equal {
			t.Errorf("test case '%s' failed. want: '%s' have: '%s'", c.title, c.expect, resJSON)
		}
	}
}
