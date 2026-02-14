package orchestrator_mvp

// 本包实现一个类 K8s 的调度/部署 MVP，用于学习核心机制：
// - 资源调度（含优先级抢占）
// - 亲和/反亲和调度约束
// - Deployment 状态 / 最小就绪时间
// - Pod 失败重建与指数退避

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

var (
	ErrNoScheduleNode = errors.New("no schedulable node")
	ErrNotFound       = errors.New("not found")
)

type Resource struct {
	CPU int // CPU 单位（例如 millicores）
	Mem int // 内存单位（例如 MB）
}

// Fits 判断节点剩余资源是否满足 Pod 需求。
func (r Resource) Fits(req Resource) bool {
	return r.CPU >= req.CPU && r.Mem >= req.Mem
}

// Sub 计算资源差值（剩余资源）。
func (r Resource) Sub(req Resource) Resource {
	return Resource{CPU: r.CPU - req.CPU, Mem: r.Mem - req.Mem}
}

type Node struct {
	Name      string
	Capacity  Resource            // 节点总资源
	Allocated Resource            // 已分配资源
	PodNames  map[string]struct{} // 运行在该节点上的 Pod 名称集合
	Labels    map[string]string   // 节点标签（zone/role 等）
	Taints    map[string]string   // 简化版污点
	Ready     bool                // 是否可调度
}

type PodStatus string

const (
	PodPending PodStatus = "Pending"
	PodRunning PodStatus = "Running"
	PodFailed  PodStatus = "Failed"
)

type PodSpec struct {
	Image        string   // 镜像
	Resource     Resource // 资源需求
	Labels       map[string]string
	Env          map[string]string
	NodeSelector map[string]string // 约束调度到具备特定 label 的节点
	Tolerations  map[string]string // 污点容忍
	Priority     int               // 调度优先级（用于抢占）
	// 亲和/反亲和：基于同节点上的 Pod 标签做简单判断。
	Affinity     map[string]string
	AntiAffinity map[string]string
}

type Pod struct {
	Name     string
	NodeName string
	Spec     PodSpec
	Status   PodStatus
	Owner    string
	// 状态管理：记录启动/失败时间以及重试控制信息。
	StartedAt    time.Time
	FailedAt     time.Time
	NextRetryAt  time.Time
	RestartCount int
	FailReason   string
}

type Deployment struct {
	Name       string
	Replicas   int
	Template   PodSpec
	Pods       []string
	Generation int64
	// MinReadySeconds：Pod 处于 Running 状态多久才视为就绪。
	MinReadySeconds time.Duration
	// MaxRetryBackoff：Pod 失败重建的最大退避时间。
	MaxRetryBackoff time.Duration
}

type Service struct {
	Name      string
	Selector  map[string]string
	Endpoints []string
	next      int
}

type Cluster struct {
	mu          sync.Mutex
	nodes       map[string]*Node
	pods        map[string]*Pod
	deployments map[string]*Deployment
	services    map[string]*Service
	nextPodID   int64
}

func NewCluster() *Cluster {
	return &Cluster{
		nodes:       make(map[string]*Node),
		pods:        make(map[string]*Pod),
		deployments: make(map[string]*Deployment),
		services:    make(map[string]*Service),
	}
}

func (c *Cluster) AddNode(name string, cap Resource) {
	c.AddNodeWithMeta(name, cap, nil, nil)
}

func (c *Cluster) AddNodeWithMeta(name string, cap Resource, labels, taints map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.nodes[name] = &Node{
		Name:     name,
		Capacity: cap,
		PodNames: make(map[string]struct{}),
		Labels:   clone(labels),
		Taints:   clone(taints),
		Ready:    true,
	}
}

func (c *Cluster) SetNodeReady(name string, ready bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	n, ok := c.nodes[name]
	if !ok {
		return ErrNotFound
	}
	n.Ready = ready
	return nil
}

func (c *Cluster) Deploy(name string, replicas int, spec PodSpec) error {
	return c.DeployAdvanced(name, replicas, spec, 0, 30*time.Second)
}

// DeployAdvanced 支持设置运行就绪时间和失败重试上限。
func (c *Cluster) DeployAdvanced(name string, replicas int, spec PodSpec, minReadySeconds, maxRetryBackoff time.Duration) error {
	if replicas <= 0 {
		return nil
	}
	if maxRetryBackoff <= 0 {
		maxRetryBackoff = 30 * time.Second
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	dep := &Deployment{
		Name:            name,
		Replicas:        replicas,
		Template:        spec,
		Generation:      1,
		MinReadySeconds: minReadySeconds,
		MaxRetryBackoff: maxRetryBackoff,
	}
	c.deployments[name] = dep

	for i := 0; i < replicas; i++ {
		podName := fmt.Sprintf("%s-%d", name, i)
		c.createPodLocked(podName, name, spec)
		dep.Pods = append(dep.Pods, podName)
	}
	c.refreshServiceEndpointsLocked()
	return nil
}

func (c *Cluster) Scale(deploy string, replicas int) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	d, ok := c.deployments[deploy]
	if !ok {
		return ErrNotFound
	}
	if replicas == d.Replicas {
		return nil
	}
	if replicas > d.Replicas {
		start := d.Replicas
		for i := start; i < replicas; i++ {
			podName := fmt.Sprintf("%s-%d", deploy, i)
			c.createPodLocked(podName, deploy, d.Template)
			d.Pods = append(d.Pods, podName)
		}
	} else {
		// scale down: remove tail pods
		for i := d.Replicas - 1; i >= replicas; i-- {
			podName := fmt.Sprintf("%s-%d", deploy, i)
			c.deletePodLocked(podName)
			d.Pods = d.Pods[:len(d.Pods)-1]
		}
	}
	d.Replicas = replicas
	c.refreshServiceEndpointsLocked()
	return nil
}

func (c *Cluster) ExposeService(name string, selector map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services[name] = &Service{Name: name, Selector: clone(selector)}
	c.refreshServiceEndpointsLocked()
}

func (c *Cluster) Route(service string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	svc, ok := c.services[service]
	if !ok {
		return "", ErrNotFound
	}
	if len(svc.Endpoints) == 0 {
		return "", ErrNoScheduleNode
	}
	ep := svc.Endpoints[svc.next%len(svc.Endpoints)]
	svc.next++
	return ep, nil
}

func (c *Cluster) ListPods() []Pod {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]Pod, 0, len(c.pods))
	for _, p := range c.pods {
		out = append(out, *p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (c *Cluster) Nodes() []Node {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]Node, 0, len(c.nodes))
	for _, n := range c.nodes {
		out = append(out, *n)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// ReconcilePending retries scheduling pending pods when resources become available.
func (c *Cluster) ReconcilePending() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	changed := 0
	for _, p := range c.pods {
		if p.Status != PodPending {
			continue
		}
		node, err := c.pickNodeLocked(p.Spec)
		if err != nil {
			continue
		}
		node.Allocated.CPU += p.Spec.Resource.CPU
		node.Allocated.Mem += p.Spec.Resource.Mem
		node.PodNames[p.Name] = struct{}{}
		p.NodeName = node.Name
		p.Status = PodRunning
		changed++
	}
	if changed > 0 {
		c.refreshServiceEndpointsLocked()
	}
	return changed
}

type DeploymentStatus struct {
	Name          string
	Generation    int64
	Desired       int
	Current       int
	Running       int
	Pending       int
	ReadyRatio    float64
	CurrentImages []string
}

func (c *Cluster) DeploymentStatus(name string) (DeploymentStatus, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	d, ok := c.deployments[name]
	if !ok {
		return DeploymentStatus{}, ErrNotFound
	}
	st := DeploymentStatus{Name: d.Name, Generation: d.Generation, Desired: d.Replicas}
	imgSet := map[string]struct{}{}
	for _, pn := range d.Pods {
		p := c.pods[pn]
		if p == nil {
			continue
		}
		st.Current++
		imgSet[p.Spec.Image] = struct{}{}
		if p.Status == PodRunning {
			st.Running++
		}
		if p.Status == PodPending {
			st.Pending++
		}
		if p.Status == PodFailed {
			st.Pending++
		}
	}
	if st.Desired > 0 {
		ready := 0
		for _, pn := range d.Pods {
			p := c.pods[pn]
			if p == nil || p.Status != PodRunning {
				continue
			}
			if d.MinReadySeconds == 0 || !time.Now().Before(p.StartedAt.Add(d.MinReadySeconds)) {
				ready++
			}
		}
		st.ReadyRatio = float64(ready) / float64(st.Desired)
	}
	for img := range imgSet {
		st.CurrentImages = append(st.CurrentImages, img)
	}
	sort.Strings(st.CurrentImages)
	return st, nil
}

// RolloutImage performs simplified rolling update with maxUnavailable/maxSurge.
func (c *Cluster) RolloutImage(deploy, newImage string, maxUnavailable, maxSurge int) error {
	if newImage == "" {
		return errors.New("new image is empty")
	}
	if maxUnavailable <= 0 {
		maxUnavailable = 1
	}
	if maxSurge <= 0 {
		maxSurge = 1
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	d, ok := c.deployments[deploy]
	if !ok {
		return ErrNotFound
	}
	d.Generation++
	d.Template.Image = newImage

	oldPods := make([]string, 0, len(d.Pods))
	for _, pn := range d.Pods {
		if p := c.pods[pn]; p != nil && p.Spec.Image != newImage {
			oldPods = append(oldPods, pn)
		}
	}

	for _, old := range oldPods {
		if len(d.Pods) < d.Replicas+maxSurge {
			newPodName := fmt.Sprintf("%s-r%d-%d", deploy, d.Generation, c.nextPodID)
			c.nextPodID++
			c.createPodLocked(newPodName, deploy, d.Template)
			d.Pods = append(d.Pods, newPodName)
		}

		running := 0
		for _, pn := range d.Pods {
			if p := c.pods[pn]; p != nil && p.Status == PodRunning {
				running++
			}
		}
		unavailable := d.Replicas - running
		if unavailable >= maxUnavailable {
			for _, p := range c.pods {
				if p.Status != PodPending {
					continue
				}
				node, err := c.pickNodeLocked(p.Spec)
				if err != nil {
					continue
				}
				node.Allocated.CPU += p.Spec.Resource.CPU
				node.Allocated.Mem += p.Spec.Resource.Mem
				node.PodNames[p.Name] = struct{}{}
				p.NodeName = node.Name
				p.Status = PodRunning
			}
		}

		c.deletePodLocked(old)
		d.Pods = removePodName(d.Pods, old)
	}

	for len(d.Pods) > d.Replicas {
		pn := d.Pods[len(d.Pods)-1]
		c.deletePodLocked(pn)
		d.Pods = d.Pods[:len(d.Pods)-1]
	}

	c.refreshServiceEndpointsLocked()
	return nil
}

func (c *Cluster) pickNodeLocked(spec PodSpec) (*Node, error) {
	names := make([]string, 0, len(c.nodes))
	for name := range c.nodes {
		names = append(names, name)
	}
	sort.Strings(names)
	var best *Node
	bestScore := -1
	for _, name := range names {
		n := c.nodes[name]
		if !c.nodePassesFiltersLocked(n, spec) {
			continue
		}
		free := n.Capacity.Sub(n.Allocated)
		if !free.Fits(spec.Resource) {
			continue
		}
		score := free.CPU + free.Mem
		if score > bestScore {
			bestScore = score
			best = n
		}
	}
	if best != nil {
		return best, nil
	}
	// Try preemption as fallback.
	if node, ok := c.tryPreemptLocked(spec, names); ok {
		return node, nil
	}
	return nil, ErrNoScheduleNode
}

func (c *Cluster) nodePassesFiltersLocked(n *Node, spec PodSpec) bool {
	if !n.Ready {
		return false
	}
	if !matchLabels(n.Labels, spec.NodeSelector) {
		return false
	}
	if !tolerates(n.Taints, spec.Tolerations) {
		return false
	}
	// Affinity: require at least one running pod on node matching labels.
	if len(spec.Affinity) > 0 && !c.hasMatchingPodOnNodeLocked(n.Name, spec.Affinity, false) {
		return false
	}
	// AntiAffinity: require no running pod on node matching labels.
	if len(spec.AntiAffinity) > 0 && c.hasMatchingPodOnNodeLocked(n.Name, spec.AntiAffinity, true) {
		return false
	}
	return true
}

func (c *Cluster) hasMatchingPodOnNodeLocked(nodeName string, labels map[string]string, anyMatch bool) bool {
	found := false
	for _, p := range c.pods {
		if p.NodeName != nodeName || p.Status != PodRunning {
			continue
		}
		match := matchLabels(p.Spec.Labels, labels)
		if anyMatch && match {
			return true
		}
		if !anyMatch && match {
			found = true
		}
	}
	return found
}

func (c *Cluster) tryPreemptLocked(spec PodSpec, names []string) (*Node, bool) {
	for _, name := range names {
		n := c.nodes[name]
		if !c.nodePassesFiltersLocked(n, spec) {
			continue
		}
		free := n.Capacity.Sub(n.Allocated)
		if free.Fits(spec.Resource) {
			return n, true
		}
		neededCPU := spec.Resource.CPU - free.CPU
		neededMem := spec.Resource.Mem - free.Mem
		if neededCPU < 0 {
			neededCPU = 0
		}
		if neededMem < 0 {
			neededMem = 0
		}
		// collect evictable lower-priority pods on this node.
		type victim struct {
			name string
			cpu  int
			mem  int
			pri  int
		}
		var victims []victim
		for podName := range n.PodNames {
			p := c.pods[podName]
			if p == nil || p.Status != PodRunning {
				continue
			}
			if p.Spec.Priority >= spec.Priority {
				continue
			}
			victims = append(victims, victim{
				name: podName, cpu: p.Spec.Resource.CPU, mem: p.Spec.Resource.Mem, pri: p.Spec.Priority,
			})
		}
		sort.Slice(victims, func(i, j int) bool { return victims[i].pri < victims[j].pri })
		accCPU, accMem := 0, 0
		toEvict := make([]string, 0)
		for _, v := range victims {
			toEvict = append(toEvict, v.name)
			accCPU += v.cpu
			accMem += v.mem
			if accCPU >= neededCPU && accMem >= neededMem {
				break
			}
		}
		if accCPU >= neededCPU && accMem >= neededMem {
			for _, podName := range toEvict {
				c.markPodFailedLocked(podName, "preempted")
			}
			return n, true
		}
	}
	return nil, false
}

func (c *Cluster) refreshServiceEndpointsLocked() {
	for _, svc := range c.services {
		var eps []string
		for _, p := range c.pods {
			if p.Status != PodRunning {
				continue
			}
			if matchLabels(p.Spec.Labels, svc.Selector) {
				eps = append(eps, p.Name)
			}
		}
		sort.Strings(eps)
		svc.Endpoints = eps
	}
}

func matchLabels(labels, selector map[string]string) bool {
	for k, v := range selector {
		if labels[k] != v {
			return false
		}
	}
	return true
}

func clone(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func tolerates(taints, tolerations map[string]string) bool {
	for k, v := range taints {
		tv, ok := tolerations[k]
		if !ok || tv != v {
			return false
		}
	}
	return true
}

func removePodName(pods []string, name string) []string {
	for i := range pods {
		if pods[i] == name {
			return append(pods[:i], pods[i+1:]...)
		}
	}
	return pods
}

func (c *Cluster) createPodLocked(podName, owner string, spec PodSpec) {
	node, err := c.pickNodeLocked(spec)
	if err != nil {
		c.pods[podName] = &Pod{Name: podName, Spec: spec, Status: PodPending, Owner: owner}
		return
	}
	node.Allocated.CPU += spec.Resource.CPU
	node.Allocated.Mem += spec.Resource.Mem
	node.PodNames[podName] = struct{}{}
	c.pods[podName] = &Pod{
		Name:      podName,
		NodeName:  node.Name,
		Spec:      spec,
		Status:    PodRunning,
		Owner:     owner,
		StartedAt: time.Now(),
	}
}

func (c *Cluster) deletePodLocked(podName string) {
	p := c.pods[podName]
	if p != nil && p.Status == PodRunning {
		if node := c.nodes[p.NodeName]; node != nil {
			node.Allocated.CPU -= p.Spec.Resource.CPU
			node.Allocated.Mem -= p.Spec.Resource.Mem
			delete(node.PodNames, podName)
		}
	}
	delete(c.pods, podName)
}

// MarkPodFailed marks a pod failed and releases node resources.
func (c *Cluster) MarkPodFailed(podName, reason string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.pods[podName]; !ok {
		return ErrNotFound
	}
	c.markPodFailedLocked(podName, reason)
	return nil
}

func (c *Cluster) markPodFailedLocked(podName, reason string) {
	p := c.pods[podName]
	if p == nil {
		return
	}
	if p.Status == PodRunning {
		if node := c.nodes[p.NodeName]; node != nil {
			node.Allocated.CPU -= p.Spec.Resource.CPU
			node.Allocated.Mem -= p.Spec.Resource.Mem
			delete(node.PodNames, podName)
		}
	}
	p.Status = PodFailed
	p.FailReason = reason
	p.FailedAt = time.Now()
	p.NodeName = ""
	p.RestartCount++
	backoff := time.Duration(math.Pow(2, float64(p.RestartCount-1))) * time.Second
	if dep, ok := c.deployments[p.Owner]; ok && dep.MaxRetryBackoff > 0 && backoff > dep.MaxRetryBackoff {
		backoff = dep.MaxRetryBackoff
	}
	p.NextRetryAt = time.Now().Add(backoff)
}

// ReconcileFailed retries failed pods with exponential backoff.
func (c *Cluster) ReconcileFailed() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	changed := 0
	now := time.Now()
	for _, p := range c.pods {
		if p.Status != PodFailed {
			continue
		}
		if now.Before(p.NextRetryAt) {
			continue
		}
		node, err := c.pickNodeLocked(p.Spec)
		if err != nil {
			// still unschedulable: move next retry window.
			backoff := time.Duration(math.Pow(2, float64(p.RestartCount))) * time.Second
			if dep, ok := c.deployments[p.Owner]; ok && dep.MaxRetryBackoff > 0 && backoff > dep.MaxRetryBackoff {
				backoff = dep.MaxRetryBackoff
			}
			p.NextRetryAt = now.Add(backoff)
			continue
		}
		node.Allocated.CPU += p.Spec.Resource.CPU
		node.Allocated.Mem += p.Spec.Resource.Mem
		node.PodNames[p.Name] = struct{}{}
		p.NodeName = node.Name
		p.Status = PodRunning
		p.StartedAt = now
		p.FailReason = ""
		p.FailedAt = time.Time{}
		changed++
	}
	if changed > 0 {
		c.refreshServiceEndpointsLocked()
	}
	return changed
}
