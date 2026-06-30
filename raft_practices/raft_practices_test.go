package raft_practices

import "testing"

func TestRequestVote(t *testing.T) {
	n := NewNode("n1")
	n.log = []LogEntry{{Term: 1, Command: "a"}}
	n.currentTerm = 1

	args := RequestVoteArgs{Term: 2, CandidateID: "n2", LastLogIndex: 0, LastLogTerm: 1}
	if !n.HandleRequestVote(args) {
		t.Fatalf("expected vote granted")
	}
	if n.votedFor != "n2" {
		t.Fatalf("votedFor mismatch")
	}

	old := RequestVoteArgs{Term: 1, CandidateID: "n3", LastLogIndex: 0, LastLogTerm: 1}
	if n.HandleRequestVote(old) {
		t.Fatalf("should reject old term")
	}
}

func TestAppendEntries(t *testing.T) {
	n := NewNode("n1")
	n.currentTerm = 1
	n.log = []LogEntry{{Term: 1, Command: "a"}}

	ok := n.HandleAppendEntries(AppendEntriesArgs{
		Term:         2,
		LeaderID:     "n2",
		PrevLogIndex: 0,
		PrevLogTerm:  1,
		Entries:      []LogEntry{{Term: 2, Command: "b"}},
		LeaderCommit: 1,
	})
	if !ok {
		t.Fatalf("append should succeed")
	}
	if len(n.log) != 2 || n.log[1].Command != "b" {
		t.Fatalf("log mismatch")
	}
}
