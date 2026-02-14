package orchestrator_mvp

import "testing"

func TestDemoOutput(t *testing.T) {
	c := NewCluster()
	c.AddNodeWithMeta("node-a", Resource{CPU: 2000, Mem: 4096}, map[string]string{"zone": "a"}, nil)
	c.AddNodeWithMeta("node-b", Resource{CPU: 1000, Mem: 2048}, map[string]string{"zone": "b"}, map[string]string{"dedicated": "gpu"})
	t.Logf("nodes: %+v", c.Nodes())

	_ = c.Deploy("web", 3, PodSpec{
		Image:        "nginx:latest",
		Resource:     Resource{CPU: 500, Mem: 256},
		Labels:       map[string]string{"app": "web"},
		NodeSelector: map[string]string{"zone": "a"},
	})
	t.Logf("pods after deploy: %+v", c.ListPods())

	c.ExposeService("web-svc", map[string]string{"app": "web"})
	ep1, _ := c.Route("web-svc")
	ep2, _ := c.Route("web-svc")
	ep3, _ := c.Route("web-svc")
	t.Logf("service route sequence: %s -> %s -> %s", ep1, ep2, ep3)

	_ = c.RolloutImage("web", "nginx:v2", 1, 1)
	st, _ := c.DeploymentStatus("web")
	t.Logf("deployment status after rollout: %+v", st)
}
