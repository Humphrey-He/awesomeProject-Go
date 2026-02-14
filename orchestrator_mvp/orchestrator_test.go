package orchestrator_mvp

import (
	"testing"
	"time"
)

func TestDeployAndRoute(t *testing.T) {
	c := NewCluster()
	c.AddNode("node-a", Resource{CPU: 2000, Mem: 4096})
	c.AddNode("node-b", Resource{CPU: 2000, Mem: 4096})

	err := c.Deploy("api", 3, PodSpec{
		Image:    "api:v1",
		Resource: Resource{CPU: 500, Mem: 256},
		Labels:   map[string]string{"app": "api"},
	})
	if err != nil {
		t.Fatalf("deploy err: %v", err)
	}
	pods := c.ListPods()
	if len(pods) != 3 {
		t.Fatalf("pods=%d want=3", len(pods))
	}

	c.ExposeService("api-svc", map[string]string{"app": "api"})
	ep1, err := c.Route("api-svc")
	if err != nil {
		t.Fatalf("route1 err: %v", err)
	}
	ep2, err := c.Route("api-svc")
	if err != nil {
		t.Fatalf("route2 err: %v", err)
	}
	if ep1 == ep2 && len(pods) > 1 {
		t.Fatalf("round robin should rotate endpoints, got same %s", ep1)
	}
}

func TestPendingWhenNoResource(t *testing.T) {
	c := NewCluster()
	c.AddNode("small", Resource{CPU: 500, Mem: 256})
	_ = c.Deploy("job", 2, PodSpec{
		Image:    "job:v1",
		Resource: Resource{CPU: 500, Mem: 256},
		Labels:   map[string]string{"app": "job"},
	})
	pods := c.ListPods()
	running := 0
	pending := 0
	for _, p := range pods {
		if p.Status == PodRunning {
			running++
		}
		if p.Status == PodPending {
			pending++
		}
	}
	if running != 1 || pending != 1 {
		t.Fatalf("running=%d pending=%d", running, pending)
	}
}

func TestScale(t *testing.T) {
	c := NewCluster()
	c.AddNode("n1", Resource{CPU: 3000, Mem: 4096})
	_ = c.Deploy("worker", 2, PodSpec{
		Image:    "worker:v1",
		Resource: Resource{CPU: 500, Mem: 128},
		Labels:   map[string]string{"app": "worker"},
	})
	if err := c.Scale("worker", 4); err != nil {
		t.Fatalf("scale up err: %v", err)
	}
	if len(c.ListPods()) != 4 {
		t.Fatalf("expected 4 pods")
	}
	if err := c.Scale("worker", 1); err != nil {
		t.Fatalf("scale down err: %v", err)
	}
	if len(c.ListPods()) != 1 {
		t.Fatalf("expected 1 pod")
	}
}

func TestSelectorTaintAndReconcile(t *testing.T) {
	c := NewCluster()
	c.AddNodeWithMeta("n1", Resource{CPU: 1000, Mem: 1024}, map[string]string{"zone": "a"}, map[string]string{"dedicated": "gpu"})
	c.AddNodeWithMeta("n2", Resource{CPU: 1000, Mem: 1024}, map[string]string{"zone": "b"}, nil)

	err := c.Deploy("ml", 1, PodSpec{
		Image:        "ml:v1",
		Resource:     Resource{CPU: 300, Mem: 128},
		Labels:       map[string]string{"app": "ml"},
		NodeSelector: map[string]string{"zone": "a"},
		// no tolerations -> should be pending on n1 taint
	})
	if err != nil {
		t.Fatalf("deploy err: %v", err)
	}
	pods := c.ListPods()
	if len(pods) != 1 || pods[0].Status != PodPending {
		t.Fatalf("expected pending pod, got %+v", pods)
	}

	// rollout-like change by scaling down/up with toleration in template
	_ = c.Scale("ml", 0)
	_ = c.Deploy("ml2", 1, PodSpec{
		Image:        "ml:v1",
		Resource:     Resource{CPU: 300, Mem: 128},
		Labels:       map[string]string{"app": "ml"},
		NodeSelector: map[string]string{"zone": "a"},
		Tolerations:  map[string]string{"dedicated": "gpu"},
	})
	if changed := c.ReconcilePending(); changed < 0 {
		t.Fatalf("invalid reconcile changed=%d", changed)
	}
}

func TestRolloutAndStatus(t *testing.T) {
	c := NewCluster()
	c.AddNode("n1", Resource{CPU: 4000, Mem: 4096})
	if err := c.Deploy("api", 3, PodSpec{
		Image:    "api:v1",
		Resource: Resource{CPU: 500, Mem: 128},
		Labels:   map[string]string{"app": "api"},
	}); err != nil {
		t.Fatalf("deploy err: %v", err)
	}

	if err := c.RolloutImage("api", "api:v2", 1, 1); err != nil {
		t.Fatalf("rollout err: %v", err)
	}
	st, err := c.DeploymentStatus("api")
	if err != nil {
		t.Fatalf("status err: %v", err)
	}
	if st.Desired != 3 || st.Current != 3 {
		t.Fatalf("unexpected status: %+v", st)
	}
	foundV2 := false
	for _, img := range st.CurrentImages {
		if img == "api:v2" {
			foundV2 = true
		}
	}
	if !foundV2 {
		t.Fatalf("expected rolled image api:v2 in status: %+v", st.CurrentImages)
	}
}

func TestPreemption(t *testing.T) {
	c := NewCluster()
	c.AddNode("n1", Resource{CPU: 1000, Mem: 1024})

	// low-priority deployment occupies node
	_ = c.Deploy("low", 1, PodSpec{
		Image:    "low:v1",
		Resource: Resource{CPU: 900, Mem: 512},
		Labels:   map[string]string{"app": "low"},
		Priority: 1,
	})
	// high-priority should preempt low and run
	_ = c.Deploy("high", 1, PodSpec{
		Image:    "high:v1",
		Resource: Resource{CPU: 800, Mem: 256},
		Labels:   map[string]string{"app": "high"},
		Priority: 10,
	})

	pods := c.ListPods()
	hasHighRunning := false
	hasLowFailed := false
	for _, p := range pods {
		if p.Owner == "high" && p.Status == PodRunning {
			hasHighRunning = true
		}
		if p.Owner == "low" && p.Status == PodFailed {
			hasLowFailed = true
		}
	}
	if !hasHighRunning || !hasLowFailed {
		t.Fatalf("preemption expected high running and low failed, pods=%+v", pods)
	}
}

func TestAffinityAndAntiAffinity(t *testing.T) {
	c := NewCluster()
	c.AddNode("n1", Resource{CPU: 2000, Mem: 2048})
	c.AddNode("n2", Resource{CPU: 2000, Mem: 2048})

	_ = c.Deploy("db", 1, PodSpec{
		Image:    "db:v1",
		Resource: Resource{CPU: 500, Mem: 256},
		Labels:   map[string]string{"app": "db"},
	})
	// pod with affinity to app=db should land on n1 where db likely scheduled first
	_ = c.Deploy("api", 1, PodSpec{
		Image:    "api:v1",
		Resource: Resource{CPU: 500, Mem: 256},
		Labels:   map[string]string{"app": "api"},
		Affinity: map[string]string{"app": "db"},
	})
	// pod with anti-affinity app=db should avoid that node
	_ = c.Deploy("worker", 1, PodSpec{
		Image:        "w:v1",
		Resource:     Resource{CPU: 500, Mem: 256},
		Labels:       map[string]string{"app": "worker"},
		AntiAffinity: map[string]string{"app": "db"},
	})

	var dbNode, apiNode, workerNode string
	for _, p := range c.ListPods() {
		if p.Owner == "db" {
			dbNode = p.NodeName
		}
		if p.Owner == "api" {
			apiNode = p.NodeName
		}
		if p.Owner == "worker" {
			workerNode = p.NodeName
		}
	}
	if dbNode == "" || apiNode == "" || workerNode == "" {
		t.Fatalf("unexpected pod placement")
	}
	if apiNode != dbNode {
		t.Fatalf("affinity pod should colocate with db: api=%s db=%s", apiNode, dbNode)
	}
	if workerNode == dbNode {
		t.Fatalf("anti-affinity pod should avoid db node: worker=%s db=%s", workerNode, dbNode)
	}
}

func TestMinReadyAndFailureBackoff(t *testing.T) {
	c := NewCluster()
	c.AddNode("n1", Resource{CPU: 2000, Mem: 2048})
	err := c.DeployAdvanced("svc", 1, PodSpec{
		Image:    "svc:v1",
		Resource: Resource{CPU: 500, Mem: 128},
		Labels:   map[string]string{"app": "svc"},
	}, 80*time.Millisecond, 120*time.Millisecond)
	if err != nil {
		t.Fatalf("deploy advanced err: %v", err)
	}
	st, _ := c.DeploymentStatus("svc")
	if st.ReadyRatio != 0 {
		t.Fatalf("ready should be 0 immediately, got %.2f", st.ReadyRatio)
	}
	time.Sleep(90 * time.Millisecond)
	st, _ = c.DeploymentStatus("svc")
	if st.ReadyRatio <= 0 {
		t.Fatalf("ready should become >0 after minReady, got %.2f", st.ReadyRatio)
	}

	pods := c.ListPods()
	var target string
	for _, p := range pods {
		if p.Owner == "svc" {
			target = p.Name
			break
		}
	}
	if target == "" {
		t.Fatalf("target pod not found")
	}
	if err := c.MarkPodFailed(target, "crash"); err != nil {
		t.Fatalf("mark failed err: %v", err)
	}
	changed := c.ReconcileFailed()
	if changed != 0 {
		t.Fatalf("should respect backoff before retry, changed=%d", changed)
	}
	time.Sleep(2 * time.Second)
	changed = c.ReconcileFailed()
	if changed == 0 {
		t.Fatalf("expected recovery after backoff")
	}
}
