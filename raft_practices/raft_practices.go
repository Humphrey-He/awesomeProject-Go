package raft_practices

import "sync"

type State int

const (
	Follower State = iota
	Candidate
	Leader
)

type LogEntry struct {
	Term    int
	Command string
}

type RequestVoteArgs struct {
	Term         int
	CandidateID  string
	LastLogIndex int
	LastLogTerm  int
}

type AppendEntriesArgs struct {
	Term         int
	LeaderID     string
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}

type RaftNode struct {
	mu          sync.Mutex
	id          string
	state       State
	currentTerm int
	votedFor    string
	log         []LogEntry
	commitIndex int
	lastApplied int
}

func NewNode(id string) *RaftNode {
	return &RaftNode{
		id:          id,
		state:       Follower,
		commitIndex: -1,
		lastApplied: -1,
	}
}

func (n *RaftNode) StartElection() RequestVoteArgs {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.currentTerm++
	n.state = Candidate
	n.votedFor = n.id
	return RequestVoteArgs{
		Term:         n.currentTerm,
		CandidateID:  n.id,
		LastLogIndex: n.lastLogIndexLocked(),
		LastLogTerm:  n.lastLogTermLocked(),
	}
}

func (n *RaftNode) HandleRequestVote(args RequestVoteArgs) bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	if args.Term < n.currentTerm {
		return false
	}
	if args.Term > n.currentTerm {
		n.currentTerm = args.Term
		n.state = Follower
		n.votedFor = ""
	}

	if n.votedFor != "" && n.votedFor != args.CandidateID {
		return false
	}

	if !n.isUpToDateLocked(args.LastLogIndex, args.LastLogTerm) {
		return false
	}

	n.votedFor = args.CandidateID
	return true
}

func (n *RaftNode) HandleAppendEntries(args AppendEntriesArgs) bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	if args.Term < n.currentTerm {
		return false
	}
	if args.Term > n.currentTerm {
		n.currentTerm = args.Term
	}
	n.state = Follower
	if args.PrevLogIndex >= 0 {
		if args.PrevLogIndex >= len(n.log) {
			return false
		}
		if n.log[args.PrevLogIndex].Term != args.PrevLogTerm {
			return false
		}
	}

	for i, entry := range args.Entries {
		idx := args.PrevLogIndex + 1 + i
		if idx < len(n.log) {
			if n.log[idx].Term != entry.Term {
				n.log = append(n.log[:idx], entry)
			}
		} else {
			n.log = append(n.log, entry)
		}
	}

	if args.LeaderCommit > n.commitIndex {
		last := len(n.log) - 1
		if args.LeaderCommit < last {
			n.commitIndex = args.LeaderCommit
		} else {
			n.commitIndex = last
		}
	}

	return true
}

func (n *RaftNode) lastLogIndexLocked() int {
	return len(n.log) - 1
}

func (n *RaftNode) lastLogTermLocked() int {
	if len(n.log) == 0 {
		return 0
	}
	return n.log[len(n.log)-1].Term
}

func (n *RaftNode) isUpToDateLocked(index, term int) bool {
	lastTerm := n.lastLogTermLocked()
	if term != lastTerm {
		return term > lastTerm
	}
	return index >= n.lastLogIndexLocked()
}
