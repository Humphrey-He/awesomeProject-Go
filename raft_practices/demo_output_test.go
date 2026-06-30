package raft_practices

import "testing"

func TestDemoOutput(t *testing.T) {
	n := NewNode("n1")
	vote := n.StartElection()
	t.Logf("start election: term=%d state=%d", vote.Term, n.state)
	ok := n.HandleAppendEntries(AppendEntriesArgs{Term: vote.Term + 1, LeaderID: "n2"})
	t.Logf("append entries accepted=%v term=%d state=%d", ok, n.currentTerm, n.state)
}
