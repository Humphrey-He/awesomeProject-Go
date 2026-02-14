// Package etcd_mvp 提供一个极简版 etcd/raft 核心模型，用于理解选主和心跳机制。
package etcd_mvp

import (
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	ErrNotLeader   = errors.New("not leader")
	ErrNoLeader    = errors.New("no leader")
	ErrMemberExist = errors.New("member exists")
	ErrMemberMiss  = errors.New("member not found")
)

type Role string

const (
	RoleFollower  Role = "Follower"
	RoleCandidate Role = "Candidate"
	RoleLeader    Role = "Leader"
)

// Member 表示一个集群节点的元数据。
type Member struct {
	ID            string    // 节点唯一 ID
	Role          Role      // 角色：Follower/Candidate/Leader
	Term          int64     // 当前任期
	LastHeartbeat time.Time // 最近一次心跳时间
}

// Cluster 是一个简化版 etcd/raft 实现：
// - 按成员 ID 排序做确定性的“选主”
// - Leader 通过心跳维持租约
// - 租约超时触发重新选主
// - 只允许 Leader 写入（Leader-only write）
// - 带 revision 的 KV 存储 + 简单 Watch/Lease/Lock 能力
type Cluster struct {
	mu sync.Mutex

	heartbeatTimeout time.Duration
	members          map[string]*Member
	leaderID         string
	term             int64
	// KV 存储相关
	kv       map[string]KV
	history  map[string][]KV // 简单历史存储，用于按 revision 查询
	revision int64
	// Watcher：按 key 订阅变更
	watchers map[string][]chan WatchEvent
	// 租约管理
	leases      map[int64]*Lease
	nextLeaseID int64
}

// KV 表示一个键在当前版本下的值信息。
type KV struct {
	Key      string
	Value    string
	Revision int64
	Version  int64 // 针对同一个 key 的版本号（递增）
	LeaseID  int64 // 绑定的租约 ID（0 表示无租约）
	ModTime  time.Time
	Deleted  bool
}

// EventType 用于区分 Put/Delete 事件。
type EventType int

const (
	EventPut EventType = iota + 1
	EventDelete
)

// WatchEvent 为 Watch 回调提供的事件结构。
type WatchEvent struct {
	Type     EventType
	Key      string
	Value    string
	Revision int64
}

// Lease 是一个简单的 TTL 租约。
type Lease struct {
	ID       int64
	TTL      time.Duration
	ExpireAt time.Time
	Keys     map[string]struct{}
}

func NewCluster(heartbeatTimeout time.Duration) *Cluster {
	if heartbeatTimeout <= 0 {
		heartbeatTimeout = 300 * time.Millisecond
	}
	return &Cluster{
		heartbeatTimeout: heartbeatTimeout,
		members:          make(map[string]*Member),
		kv:               make(map[string]KV),
		history:          make(map[string][]KV),
		watchers:         make(map[string][]chan WatchEvent),
		leases:           make(map[int64]*Lease),
	}
}

func (c *Cluster) AddMember(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.members[id]; ok {
		return ErrMemberExist
	}
	c.members[id] = &Member{
		ID:            id,
		Role:          RoleFollower,
		Term:          c.term,
		LastHeartbeat: time.Now(),
	}
	// 第一个加入的节点直接作为 Leader，简化启动过程。
	if c.leaderID == "" {
		c.term++
		c.leaderID = id
		c.members[id].Role = RoleLeader
		c.members[id].Term = c.term
		c.members[id].LastHeartbeat = time.Now()
	}
	return nil
}

func (c *Cluster) RemoveMember(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.members[id]; !ok {
		return ErrMemberMiss
	}
	delete(c.members, id)
	if c.leaderID == id {
		c.leaderID = ""
		c.electLocked(time.Now())
	}
	return nil
}

func (c *Cluster) Leader() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.leaderID == "" {
		return "", ErrNoLeader
	}
	return c.leaderID, nil
}

func (c *Cluster) Heartbeat(leaderID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	m, ok := c.members[leaderID]
	if !ok {
		return ErrMemberMiss
	}
	// 若当前没有 leader，允许通过心跳快速确立 leader。
	if c.leaderID == "" {
		c.term++
		c.leaderID = leaderID
	}
	if c.leaderID != leaderID {
		return ErrNotLeader
	}
	m.Role = RoleLeader
	m.Term = c.term
	m.LastHeartbeat = time.Now()
	for id, other := range c.members {
		if id == leaderID {
			continue
		}
		other.Role = RoleFollower
		other.Term = c.term
	}
	return nil
}

// Tick 用于推进时间：检查 leader 租约，如超时则发起新一轮选举。
func (c *Cluster) Tick(now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.leaderID == "" {
		c.electLocked(now)
		return
	}
	leader := c.members[c.leaderID]
	if leader == nil || now.Sub(leader.LastHeartbeat) > c.heartbeatTimeout {
		c.leaderID = ""
		c.electLocked(now)
	}
}

func (c *Cluster) Put(leaderID, key, value string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.leaderID == "" {
		return ErrNoLeader
	}
	if c.leaderID != leaderID {
		return ErrNotLeader
	}
	c.putKVLocked(key, value, 0)
	return nil
}

func (c *Cluster) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	kv, ok := c.kv[key]
	if !ok || kv.Deleted {
		return "", false
	}
	return kv.Value, true
}

func (c *Cluster) SnapshotMembers() []Member {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]Member, 0, len(c.members))
	for _, m := range c.members {
		out = append(out, *m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (c *Cluster) electLocked(now time.Time) {
	if len(c.members) == 0 {
		return
	}
	// deterministic "vote winner": lexicographically smallest alive member.
	ids := make([]string, 0, len(c.members))
	for id := range c.members {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	winner := ids[0]
	c.term++
	c.leaderID = winner
	for _, id := range ids {
		m := c.members[id]
		m.Term = c.term
		m.LastHeartbeat = now
		if id == winner {
			m.Role = RoleLeader
		} else {
			m.Role = RoleFollower
		}
	}
}

// putKVLocked 在持有锁的前提下写入 KV，并递增 revision、生成历史及通知 watcher。
func (c *Cluster) putKVLocked(key, value string, leaseID int64) {
	c.revision++
	now := time.Now()
	prev, existed := c.kv[key]
	var ver int64 = 1
	if existed {
		ver = prev.Version + 1
	}
	entry := KV{
		Key:      key,
		Value:    value,
		Revision: c.revision,
		Version:  ver,
		LeaseID:  leaseID,
		ModTime:  now,
	}
	c.kv[key] = entry
	c.history[key] = append(c.history[key], entry)
	if leaseID != 0 {
		if l, ok := c.leases[leaseID]; ok {
			if l.Keys == nil {
				l.Keys = make(map[string]struct{})
			}
			l.Keys[key] = struct{}{}
		}
	}
	c.broadcastLocked(WatchEvent{Type: EventPut, Key: key, Value: value, Revision: entry.Revision})
}

// deleteKeyLocked 逻辑删除 key，用于租约到期或主动删除。
func (c *Cluster) deleteKeyLocked(key string) {
	entry, ok := c.kv[key]
	if !ok || entry.Deleted {
		return
	}
	c.revision++
	entry.Deleted = true
	entry.Revision = c.revision
	entry.ModTime = time.Now()
	c.kv[key] = entry
	c.history[key] = append(c.history[key], entry)
	c.broadcastLocked(WatchEvent{Type: EventDelete, Key: key, Value: entry.Value, Revision: entry.Revision})
}

// GetAtRevision 返回指定 revision 及之前最近一次写入的值。
func (c *Cluster) GetAtRevision(key string, rev int64) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	h := c.history[key]
	if len(h) == 0 {
		return "", false
	}
	for i := len(h) - 1; i >= 0; i-- {
		if h[i].Revision <= rev && !h[i].Deleted {
			return h[i].Value, true
		}
	}
	return "", false
}

// WatchKey 订阅某个 key 的变更事件。
func (c *Cluster) WatchKey(key string, buffer int) (<-chan WatchEvent, func()) {
	if buffer <= 0 {
		buffer = 8
	}
	ch := make(chan WatchEvent, buffer)
	c.mu.Lock()
	c.watchers[key] = append(c.watchers[key], ch)
	c.mu.Unlock()
	cancel := func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		wlist := c.watchers[key]
		for i, w := range wlist {
			if w == ch {
				wlist = append(wlist[:i], wlist[i+1:]...)
				break
			}
		}
		if len(wlist) == 0 {
			delete(c.watchers, key)
		} else {
			c.watchers[key] = wlist
		}
		close(ch)
	}
	return ch, cancel
}

func (c *Cluster) broadcastLocked(ev WatchEvent) {
	for _, ch := range c.watchers[ev.Key] {
		select {
		case ch <- ev:
		default:
			// 丢弃慢消费者事件，避免阻塞 leader。
		}
	}
}

// LeaseGrant 创建一个新的租约。
func (c *Cluster) LeaseGrant(ttl time.Duration) int64 {
	if ttl <= 0 {
		ttl = time.Second
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.nextLeaseID++
	id := c.nextLeaseID
	c.leases[id] = &Lease{
		ID:       id,
		TTL:      ttl,
		ExpireAt: time.Now().Add(ttl),
		Keys:     make(map[string]struct{}),
	}
	return id
}

// LeaseRevoke 主动撤销租约并删除绑定的 keys。
func (c *Cluster) LeaseRevoke(id int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	l, ok := c.leases[id]
	if !ok {
		return
	}
	for key := range l.Keys {
		c.deleteKeyLocked(key)
	}
	delete(c.leases, id)
}

// LeaseTick 检查租约是否到期，到期则删除绑定 key。
func (c *Cluster) LeaseTick(now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for id, l := range c.leases {
		if now.After(l.ExpireAt) {
			for key := range l.Keys {
				c.deleteKeyLocked(key)
			}
			delete(c.leases, id)
		}
	}
}

// PutWithLease 带租约的写入，用于实现 TTL Key、服务注册、分布式锁等。
func (c *Cluster) PutWithLease(leaderID, key, value string, leaseID int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.leaderID == "" {
		return ErrNoLeader
	}
	if c.leaderID != leaderID {
		return ErrNotLeader
	}
	if leaseID != 0 {
		if _, ok := c.leases[leaseID]; !ok {
			return errors.New("lease not found")
		}
	}
	c.putKVLocked(key, value, leaseID)
	return nil
}

// AcquireLock 通过租约+Key 实现分布式锁。
func (c *Cluster) AcquireLock(leaderID, name string, ttl time.Duration) (int64, bool, error) {
	lockKey := "/locks/" + name
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.leaderID == "" {
		return 0, false, ErrNoLeader
	}
	if c.leaderID != leaderID {
		return 0, false, ErrNotLeader
	}
	if kv, ok := c.kv[lockKey]; ok && !kv.Deleted {
		return 0, false, nil
	}
	if ttl <= 0 {
		ttl = time.Second
	}
	c.nextLeaseID++
	id := c.nextLeaseID
	lease := &Lease{
		ID:       id,
		TTL:      ttl,
		ExpireAt: time.Now().Add(ttl),
		Keys:     map[string]struct{}{lockKey: {}},
	}
	c.leases[id] = lease
	c.revision++
	c.kv[lockKey] = KV{
		Key:      lockKey,
		Value:    leaderID,
		Revision: c.revision,
		Version:  1,
		LeaseID:  id,
		ModTime:  time.Now(),
	}
	c.broadcastLocked(WatchEvent{Type: EventPut, Key: lockKey, Value: leaderID, Revision: c.revision})
	return id, true, nil
}

// ReleaseLock 通过撤销租约释放锁。
func (c *Cluster) ReleaseLock(id int64) {
	c.LeaseRevoke(id)
}

// RegisterService 将实例注册到 /services 前缀下，依赖租约自动摘除。
func (c *Cluster) RegisterService(leaderID, svc, id, endpoint string, ttl time.Duration) (int64, error) {
	leaseID := c.LeaseGrant(ttl)
	key := "/services/" + svc + "/" + id
	if err := c.PutWithLease(leaderID, key, endpoint, leaseID); err != nil {
		return 0, err
	}
	return leaseID, nil
}

// DiscoverService 直接扫描 /services/{svc}/ 前缀，返回 endpoint 列表。
func (c *Cluster) DiscoverService(svc string) []string {
	prefix := "/services/" + svc + "/"
	c.mu.Lock()
	defer c.mu.Unlock()
	var out []string
	for k, v := range c.kv {
		if v.Deleted {
			continue
		}
		if strings.HasPrefix(k, prefix) {
			out = append(out, v.Value)
		}
	}
	sort.Strings(out)
	return out
}
