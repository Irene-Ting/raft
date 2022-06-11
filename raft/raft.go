package raft

import (
	"context"
	"time"

	"github.com/justin0u0/raft/pb"
	"go.uber.org/zap"
)

type Raft struct {
	pb.UnimplementedRaftServer

	*raftState
	persister Persister

	id    uint32
	peers map[uint32]Peer

	config *Config
	logger *zap.Logger

	// lastHeartbeat stores the last time of a valid RPC received from the leader
	lastHeartbeat time.Time

	// rpcCh stores incoming RPCs
	rpcCh chan *rpc
	// applyCh stores logs that can be applied
	applyCh chan *pb.Entry
}

var _ pb.RaftServer = (*Raft)(nil)

func NewRaft(id uint32, peers map[uint32]Peer, persister Persister, config *Config, logger *zap.Logger) *Raft {
	raftState := &raftState{
		state:       Follower,
		currentTerm: 0,
		votedFor:    0,
		logs:        make([]*pb.Entry, 0),
		commitIndex: 0,
		lastApplied: 0,
		nextIndex:   make(map[uint32]uint64),
		matchIndex:  make(map[uint32]uint64),
	}

	return &Raft{
		raftState:     raftState,
		persister:     persister,
		id:            id,
		peers:         peers,
		config:        config,
		logger:        logger.With(zap.Uint32("id", id)),
		lastHeartbeat: time.Now(),
		rpcCh:         make(chan *rpc),
		applyCh:       make(chan *pb.Entry),
	}
}

// RPC handlers
// an interface for the client to apply new log to the Raft leader
func (r *Raft) applyCommand(req *pb.ApplyCommandRequest) (*pb.ApplyCommandResponse, error) {
	// DONE: (B.1)* - if not leader, reject client operation and returns `errNotLeader`
	if r.state != Leader {
		return nil, errNotLeader
	}
	// DONE: (B.1)* - create a new log entry, append to the local entries
	// Hint:
	// - use `getLastLog` to get the last log ID
	// - use `appendLogs` to append new log
	id, _ := r.getLastLog()
	newEntry := pb.Entry{Id: id + 1, Term: r.currentTerm, Data: req.Data}
	r.appendLogs([]*pb.Entry{&newEntry})

	// DONE: (B.1)* - return the new log entry
	return &pb.ApplyCommandResponse{Entry: &newEntry}, nil
}

func (r *Raft) appendEntries(req *pb.AppendEntriesRequest) (*pb.AppendEntriesResponse, error) {
	// DONE: (A.1) - reply false if term < currentTerm
	// Log: r.logger.Info("reject append entries since current term is older")
	if req.Term < r.currentTerm {
		r.logger.Info("reject append entries since current term is older", zap.Uint64("req.Term", req.Term), zap.Uint64("r.currentTerm", r.currentTerm))
		return &pb.AppendEntriesResponse{Term: r.currentTerm, Success: false}, nil
	}

	// DONE: (A.2)* - reset the `lastHeartbeat`
	// Description: start from the current line, the current request is a valid RPC
	r.lastHeartbeat = time.Now()

	// DONE: (A.3) - if RPC request or response contains term T > currentTerm: set currentTerm = T, convert to follower
	// Hint: use `toFollower` to convert to follower
	// Log: r.logger.Info("increase term since receive a newer one", zap.Uint64("term", r.currentTerm))
	if req.Term > r.currentTerm {
		r.toFollower(req.Term)
		r.logger.Info("increase term since receive a newer one", zap.Uint64("term", r.currentTerm))
	}

	// DONE: (A.4) - if AppendEntries RPC received from new leader: convert to follower
	// Log: r.logger.Info("receive request from leader, fallback to follower", zap.Uint64("term", r.currentTerm))
	// for example, a candidate receive request from leader and then convert to follower
	if req.Term >= r.currentTerm && r.state == Candidate {
		r.toFollower(req.Term)
		r.logger.Info("receive request from leader, fallback to follower", zap.Uint64("term", r.currentTerm))
	}

	prevLogId := req.GetPrevLogId()
	prevLogTerm := req.GetPrevLogTerm()
	if prevLogId != 0 && prevLogTerm != 0 {
		// DONE: (B.2) - reply false if log doesn’t contain an entry at prevLogIndex whose term matches prevLogTerm
		// Hint: use `getLog` to get log with ID equals to prevLogId
		// Log: r.logger.Info("the given previous log from leader is missing or mismatched", zap.Uint64("prevLogId", prevLogId), zap.Uint64("prevLogTerm", prevLogTerm), zap.Uint64("logTerm", log.GetTerm()))
		log := r.getLog(prevLogId)
		if (log == nil) || (log.Term != prevLogTerm) {
			r.logger.Info("the given previous log from leader is missing or mismatched", zap.Uint64("prevLogId", prevLogId), zap.Uint64("prevLogTerm", prevLogTerm), zap.Uint64("logTerm", log.GetTerm()))
			return &pb.AppendEntriesResponse{Term: r.currentTerm, Success: false}, nil
		}
	}

	if len(req.GetEntries()) != 0 {
		// DONE: (B.3) - if an existing entry conflicts with a new one (same index but different terms), delete the existing entry and all that follow it
		entries := req.GetEntries()
		startIndex := len(entries) - 1
		for i, reqEntry := range entries {
			if rEntry := r.getLog(reqEntry.Id); rEntry != nil {
				if rEntry.Term != reqEntry.Term {
					startIndex = i
					r.deleteLogs(reqEntry.Id - 1) // deleteLogs(id) doesn't delete log with id
					break
				}
			} else {
				startIndex = i
				break
			}
		}
		// DONE: (B.4) - append any new entries not already in the log
		// Hint: use `deleteLogs` follows by `appendLogs`
		// Log: r.logger.Info("receive and append new entries", zap.Int("newEntries", len(req.GetEntries())), zap.Int("numberOfEntries", len(r.logs)))
		logs := req.Entries[startIndex:]
		r.appendLogs(logs)
		r.logger.Info("receive and append new entries", zap.Int("newEntries", len(req.GetEntries())), zap.Int("numberOfEntries", len(r.logs)))
	}

	// DONE: (B.5) - if leaderCommit > commitIndex, set commitIndex = min(leaderCommit, index of last new entry)
	// that is, leader committed some entries but follower hasn't done yet
	// Hint: use `getLastLog` to get the index of last new entry
	// Hint: use `applyLogs` to apply(commit) new logs in background
	// Log: r.logger.Info("update commit index from leader", zap.Uint64("commitIndex", r.commitIndex))
	if req.LeaderCommitId > r.commitIndex {
		leaderCommitId := req.LeaderCommitId
		lastId, _ := r.getLastLog()

		if leaderCommitId < lastId {
			r.setCommitIndex(leaderCommitId)
		} else {
			r.setCommitIndex(lastId) // all logs become committed
		}
		r.applyLogs(r.applyCh)
		r.logger.Info("update commit index from leader", zap.Uint64("commitIndex", r.commitIndex))
	}
	return &pb.AppendEntriesResponse{Term: r.currentTerm, Success: true}, nil
}

func (r *Raft) requestVote(req *pb.RequestVoteRequest) (*pb.RequestVoteResponse, error) {
	// DONE: (A.5) - reply false if term < currentTerm
	// Log: r.logger.Info("reject request vote since current term is older")
	if req.Term < r.currentTerm {
		r.logger.Info("reject request vote since current term is older")
		return &pb.RequestVoteResponse{Term: r.currentTerm, VoteGranted: false}, nil
	}

	// DONE: (A.6) - if RPC request or response contains term T > currentTerm: set currentTerm = T, convert to follower
	// Hint: use `toFollower` to convert to follower
	// Log: r.logger.Info("increase term since receive a newer one", zap.Uint64("term", r.currentTerm))
	if req.Term > r.currentTerm {
		r.toFollower(req.Term)
		r.logger.Info("increase term since receive a newer one", zap.Uint64("term", r.currentTerm))
	}

	if r.votedFor != 0 && r.votedFor != req.CandidateId {
		// DONE: (A.7) - if votedFor is null or candidateId, and candidate’s log is at least as up-to-date as receiver’s log, grant vote
		// Hint: (fix the condition) if already vote for another candidate, reply false

		r.logger.Info("reject since already vote for another candidate",
			zap.Uint64("term", r.currentTerm),
			zap.Uint32("votedFor", r.votedFor))

		return &pb.RequestVoteResponse{Term: r.currentTerm, VoteGranted: false}, nil
	}

	localLastId, localLastTerm := r.getLastLog()
	candidateLastId := req.LastLogId
	candidateLastTerm := req.LastLogTerm
	if (localLastTerm > candidateLastTerm) || (localLastTerm == candidateLastTerm && localLastId > candidateLastId) {
		// DONE: (A.7) - if votedFor is null or candidateId, and candidate’s log is at least as up-to-date as receiver’s log, grant vote
		// Hint: (fix the condition) if the local last entry is more up-to-date than the candidate's last entry, reply false
		// Hint: use `getLastLog` to get the last log entry
		r.logger.Info("reject since last entry is more up-to-date")

		return &pb.RequestVoteResponse{Term: r.currentTerm, VoteGranted: false}, nil
	}

	// DONE: (A.7) - if votedFor is null or candidateId, and candidate’s log is at least as up-to-date as receiver’s log, grant vote
	// Hint: now vote should be granted, use `voteFor` to set votedFor
	r.voteFor(req.GetCandidateId(), false)
	r.logger.Info("vote for another candidate", zap.Uint32("votedFor", r.votedFor))

	// DONE: (A.8)* - reset the `lastHeartbeat`
	// Description: start from the current line, the current request is a valid RPC
	r.lastHeartbeat = time.Now()

	return &pb.RequestVoteResponse{Term: r.currentTerm, VoteGranted: true}, nil
}

// raft main loop

func (r *Raft) Run(ctx context.Context) {
	if err := r.loadRaftState(r.persister); err != nil {
		r.logger.Error("fail to load raft state", zap.Error(err))
		return
	}

	r.logger.Info("starting raft",
		zap.Uint64("term", r.currentTerm),
		zap.Uint32("votedFor", r.votedFor),
		zap.Int("logs", len(r.logs)))

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("raft server stopped gracefully")
			return
		default:
		}

		switch r.state {
		case Follower:
			r.runFollower(ctx)
		case Candidate:
			r.runCandidate(ctx)
		case Leader:
			r.runLeader(ctx)
		}
	}
}

func (r *Raft) ApplyCh() <-chan *pb.Entry {
	return r.applyCh
}

// follower related

func (r *Raft) runFollower(ctx context.Context) {
	r.logger.Info("running follower")

	timeoutCh := randomTimeout(r.config.HeartbeatTimeout)

	for r.state == Follower {
		select {
		case <-ctx.Done():
			return

		case <-timeoutCh:
			timeoutCh = randomTimeout(r.config.HeartbeatTimeout)

			if time.Now().Sub(r.lastHeartbeat) > r.config.HeartbeatTimeout {
				r.handleFollowerHeartbeatTimeout()
			}

		case rpc := <-r.rpcCh:
			r.handleRPCRequest(rpc)
		}
	}
}

func (r *Raft) handleFollowerHeartbeatTimeout() {
	// DONE: (A.9) - if election timeout elapses without receiving AppendEntries RPC from current leader or granting vote to candidate: convert to candidate
	// Hint: use `toCandidate` to convert to candidate
	r.toCandidate()
	r.logger.Info("heartbeat timeout, change state from follower to candidate")
}

// candidate related

type voteResult struct {
	*pb.RequestVoteResponse
	peerId uint32
}

func (r *Raft) runCandidate(ctx context.Context) {
	r.logger.Info("running candidate")

	grantedVotes := 0
	votesNeeded := (len(r.peers)+1)/2 + 1

	// vote for itself
	r.voteForSelf(&grantedVotes)

	// request votes from peers
	voteCh := make(chan *voteResult, len(r.peers))
	r.broadcastRequestVote(ctx, voteCh)

	// wait until:
	// 1. it wins the election
	// 2. another server establishes itself as leader (see AppendEntries)
	// 3. election timeout

	timeoutCh := randomTimeout(r.config.ElectionTimeout)

	for r.state == Candidate {
		select {
		case <-ctx.Done():
			return

		case vote := <-voteCh:
			r.handleVoteResult(vote, &grantedVotes, votesNeeded)

		case <-timeoutCh:
			r.logger.Info("election timeout reached, restarting election")
			return

		case rpc := <-r.rpcCh:
			r.handleRPCRequest(rpc)
		}
	}
}

func (r *Raft) voteForSelf(grantedVotes *int) {
	// TODO: (A.10) increment currentTerm
	// TODO: (A.10) vote for self
	// Hint: use `voteFor` to vote for self
	r.voteFor(r.id, true) // this function increases currentTerm
	(*grantedVotes)++
	r.logger.Info("vote for self", zap.Uint64("term", r.currentTerm))
}

func (r *Raft) broadcastRequestVote(ctx context.Context, voteCh chan *voteResult) {
	r.logger.Info("broadcast request vote", zap.Uint64("term", r.currentTerm))

	lastId, lastTerm := r.getLastLog()
	req := &pb.RequestVoteRequest{
		// DONE: (A.11) - send RequestVote RPCs to all other servers (set all fields of `req`)
		// Hint: use `getLastLog` to get the last log entry
		Term:        r.currentTerm,
		CandidateId: r.id,
		LastLogId:   lastId,
		LastLogTerm: lastTerm,
	}

	// DONE: (A.11) - send RequestVote RPCs to all other servers (modify the code to send `RequestVote` RPCs in parallel)
	for peerId, peer := range r.peers {
		peerId := peerId
		peer := peer
		go func() {
			resp, err := peer.RequestVote(ctx, req)
			if err != nil {
				r.logger.Error("fail to send RequestVote RPC", zap.Error(err), zap.Uint32("peer", peerId))
				return
			}

			voteCh <- &voteResult{RequestVoteResponse: resp, peerId: peerId}
		}()
	}
}

func (r *Raft) handleVoteResult(vote *voteResult, grantedVotes *int, votesNeeded int) {
	// DONE: (A.12) - if RPC request or response contains term T > currentTerm: set currentTerm = T, convert to follower
	// Hint: use `toFollower` to convert to follower
	// Log: r.logger.Info("receive new term on RequestVote response, fallback to follower", zap.Uint32("peer", vote.peerId))
	if vote.RequestVoteResponse.Term > r.currentTerm {
		r.toFollower(vote.RequestVoteResponse.Term)
		r.logger.Info("receive new term on RequestVote response, fallback to follower", zap.Uint32("peer", vote.peerId))
	}
	if vote.VoteGranted {
		(*grantedVotes)++
		r.logger.Info("vote granted", zap.Uint32("peer", vote.peerId), zap.Int("grantedVote", (*grantedVotes)), zap.Int("votesNeeded", (votesNeeded)))
	}

	// DONE: (A.13) - if votes received from majority of servers: become leader
	// Log: r.logger.Info("election won", zap.Int("grantedVote", (*grantedVotes)), zap.Uint64("term", r.currentTerm))
	// Hint: use `toLeader` to convert to leader
	if *grantedVotes >= votesNeeded {
		r.toLeader()
		r.logger.Info("election won", zap.Int("grantedVote", (*grantedVotes)), zap.Uint64("term", r.currentTerm))
	}
}

// leader related

type appendEntriesResult struct {
	*pb.AppendEntriesResponse
	req    *pb.AppendEntriesRequest
	peerId uint32
}

func (r *Raft) runLeader(ctx context.Context) {
	timeoutCh := randomTimeout(r.config.HeartbeatInterval)

	appendEntriesResultCh := make(chan *appendEntriesResult, len(r.peers))

	// reset `nextIndex` and `matchIndex`
	lastLogId, _ := r.getLastLog()
	for peerId := range r.peers {
		r.nextIndex[peerId] = lastLogId + 1
		r.matchIndex[peerId] = 0
	}

	for r.state == Leader {
		select {
		case <-ctx.Done():
			return

		case <-timeoutCh:
			timeoutCh = randomTimeout(r.config.HeartbeatInterval)
			r.broadcastAppendEntries(ctx, appendEntriesResultCh)

		case result := <-appendEntriesResultCh:
			r.handleAppendEntriesResult(result)

		case rpc := <-r.rpcCh:
			r.handleRPCRequest(rpc)
		}
	}
}

func (r *Raft) broadcastAppendEntries(ctx context.Context, appendEntriesResultCh chan *appendEntriesResult) {
	r.logger.Info("broadcast append entries")
	for peerId, peer := range r.peers {
		peerId := peerId
		peer := peer

		// DONE: (A.14) - send initial empty AppendEntries RPCs (heartbeat) to each server; repeat during idle periods to prevent election timeouts
		// Hint: set `req` with the correct fields (entries, prevLogId, prevLogTerm can be ignored for heartbeat)
		// DONE: (B.6) - send AppendEntries RPC with log entries starting at nextIndex
		// Hint: set `req` with the correct fields (entries, prevLogId and prevLogTerm MUST be set)
		// Hint: use `getLog` to get specific log, `getLogs` to get all logs after and include the specific log Id
		// Log: r.logger.Debug("send append entries", zap.Uint32("peer", peerId), zap.Any("request", req), zap.Int("entries", len(entries)))
		req := &pb.AppendEntriesRequest{}
		nextIndex := r.nextIndex[peerId]
		matchIndex := r.matchIndex[peerId]
		if lastId, _ := r.getLastLog(); matchIndex < lastId {
			entries := r.getLogs(nextIndex)
			prev := r.getLog(nextIndex - 1)
			if prev == nil {
				req = &pb.AppendEntriesRequest{Term: r.currentTerm, LeaderId: r.id, LeaderCommitId: r.commitIndex, Entries: entries, PrevLogId: 0, PrevLogTerm: 0}
			} else {
				req = &pb.AppendEntriesRequest{Term: r.currentTerm, LeaderId: r.id, LeaderCommitId: r.commitIndex, Entries: entries, PrevLogId: prev.Id, PrevLogTerm: prev.Term}
			}
			r.logger.Debug("send append entries", zap.Uint32("peer", peerId), zap.Any("request", req), zap.Int("entries", len(entries)))
		} else {
			req = &pb.AppendEntriesRequest{Term: r.currentTerm, LeaderId: r.id, LeaderCommitId: r.commitIndex}
		}

		// DONE: (A.14) & (B.6)
		// Hint: modify the code to send `AppendEntries` RPCs in parallel
		go func() {
			resp, err := peer.AppendEntries(ctx, req)
			if err != nil {
				r.logger.Error("fail to send AppendEntries RPC", zap.Error(err), zap.Uint32("peer", peerId))
				// connection issue, should not be handled
				return
			}

			appendEntriesResultCh <- &appendEntriesResult{
				AppendEntriesResponse: resp,
				req:                   req,
				peerId:                peerId,
			}
		}()

	}
}

func (r *Raft) handleAppendEntriesResult(result *appendEntriesResult) {
	// DONE: (A.15) - if RPC request or response contains term T > currentTerm: set currentTerm = T, convert to follower
	// Hint: use `toFollower` to convert to follower
	// Log: r.logger.Info("receive new term on AppendEntries response, fallback to follower", zap.Uint32("peer", result.peerId))
	if result.Term > r.currentTerm {
		r.toFollower(result.Term)
		r.logger.Info("receive new term on AppendEntries response, fallback to follower", zap.Uint32("peer", result.peerId))
	}
	entries := result.req.GetEntries()

	if !result.GetSuccess() {
		// DONE: (B.7) - if AppendEntries fails because of log inconsistency: decrement nextIndex and retry
		// Hint: use `setNextAndMatchIndex` to decrement nextIndex
		// Log: logger.Info("append entries failed, decrease next index", zap.Uint64("nextIndex", nextIndex), zap.Uint64("matchIndex", matchIndex))
		peerId := result.peerId
		nextIndex := r.nextIndex[peerId] - 1
		matchIndex := r.matchIndex[peerId]
		r.setNextAndMatchIndex(peerId, nextIndex, matchIndex)
		r.logger.Info("append entries failed, decrease next index", zap.Uint64("nextIndex", nextIndex), zap.Uint64("matchIndex", matchIndex))
	} else if len(entries) != 0 {
		// DONE: (B.8) - if successful: update nextIndex and matchIndex for follower
		// Hint: use `setNextAndMatchIndex` to update nextIndex and matchIndex
		// Log: logger.Info("append entries successfully, set next index and match index", zap.Uint32("peer", result.peerId), zap.Uint64("nextIndex", nextIndex), zap.Uint64("matchIndex", matchIndex))
		peerId := result.peerId
		nextIndex := entries[len(entries)-1].Id + 1
		matchIndex := entries[len(entries)-1].Id
		r.setNextAndMatchIndex(peerId, nextIndex, matchIndex)
		r.logger.Info("append entries successfully, set next index and match index", zap.Uint32("peer", result.peerId), zap.Uint64("nextIndex", nextIndex), zap.Uint64("matchIndex", matchIndex))
	}

	replicasNeeded := (len(r.peers)+1)/2 + 1

	logs := r.getLogs(r.commitIndex + 1)
	for i := len(logs) - 1; i >= 0; i-- {
		// DONE: (B.9) if there exist an N such that N > commitIndex, a majority of matchIndex[i] >= N, and log[N].term == currentTerm: set commitIndex = N
		// Hint: find if such N exists
		// Hint: if such N exists, use `setCommitIndex` to set commit index
		// Hint: if such N exists, use `applyLogs` to apply logs
		N := logs[i].Id
		if N <= r.commitIndex {
			break
		}

		if logs[i].Term != r.currentTerm {
			continue
		}

		replicas := 1

		for peerId := range r.peers {
			if r.matchIndex[peerId] >= N {
				replicas++
			}
		}

		if replicas >= replicasNeeded {
			r.setCommitIndex(N)
			r.applyLogs(r.applyCh)
			break
		}
	}
}
