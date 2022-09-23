package surfstore

import (
	context "context"
	"math"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type RaftSurfstore struct {
	// Default Entries
	isLeader bool
	term     int64
	log      []*UpdateOperation

	metaStore *MetaStore

	// New entries
	commitIndex int64
	lastApplied int64

	ip       string
	ipList   []string
	serverId int64

	/*--------------- Chaos Monkey --------------*/
	isCrashed      bool
	isCrashedMutex *sync.RWMutex
	notCrashedCond *sync.Cond

	UnimplementedRaftSurfstoreServer
}

func (s *RaftSurfstore) GetFileInfoMap(ctx context.Context, empty *emptypb.Empty) (*FileInfoMap, error) {
	// Check if current server is crashed
	if isCrashed := s.isCrashed; isCrashed {
		return nil, ERR_SERVER_CRASHED
	}

	// Check if current server is leader
	if isLeader := s.isLeader; !isLeader {
		return nil, ERR_NOT_LEADER
	}

	// if majority of nodes are alive, return correct answer
	// if majority of nodes are dead, block until the majority of nodes recover
	for {
		ifMajorityAlive, _ := s.SendHeartbeat(ctx, empty)
		if ifMajorityAlive.Flag {
			break
		}
	}

	return s.metaStore.GetFileInfoMap(ctx, empty)
}

func (s *RaftSurfstore) GetBlockStoreAddr(ctx context.Context, empty *emptypb.Empty) (*BlockStoreAddr, error) {
	// Check if current server is crashed
	if isCrashed := s.isCrashed; isCrashed {
		return nil, ERR_SERVER_CRASHED
	}

	// Ch ueck if current server is leader
	if isLeader := s.isLeader; !isLeader {
		return nil, ERR_NOT_LEADER
	}

	// if majority of nodes are alive, return correct answer
	// if majority of nodes are dead, block until majority of nodes recover
	for {
		ifMajorityAlive, _ := s.SendHeartbeat(ctx, empty)
		if ifMajorityAlive.Flag {
			break
		}
	}

	return s.metaStore.GetBlockStoreAddr(ctx, empty)
}

// 1. Client calls UpdateFile to update metadata
// 2. if the node is not the leader , return NOT LEADER ERROR
// 3. if the node is leader, append the entry to its own log and try to replicate the entry to other nodes
// 4. if majority of nodes successfully append that entry, try to commit that entry
// 5. After commitment, in next heartbeat message, ask other nodes commit the entry too (using LeaderCommit param in AppendEntryInput)
func (s *RaftSurfstore) UpdateFile(ctx context.Context, filemeta *FileMetaData) (*Version, error) {
	// check if the node is crashed
	if s.isCrashed {
		return nil, ERR_SERVER_CRASHED
	}
	//2. check if the node is leader
	if !s.isLeader {
		return nil, ERR_NOT_LEADER
	}

	//3. if the node is leader, append the entry to its own log and call AppendEntriesRPC to replicate its entry
	// append the entry to its own log
	newEntry := UpdateOperation{
		Term:         s.term,
		FileMetaData: filemeta,
	}
	s.log = append(s.log, &newEntry)
	outputChan := make(chan *AppendEntryOutput, len(s.ipList)-1)
	// call AppendEntriesRPC
	for idx, addr := range s.ipList {
		if int64(idx) == s.serverId {
			continue
		}
		// concurrently call tryAppendEntries()
		go s.tryAppendEntries(addr, int64(len(s.log)-1), outputChan)
	}

	// receiving channels to see if the majority of nodes successfully append the entries
	succNum := 1
	for {
		output := <-outputChan
		if output != nil && output.Success {
			succNum++
		}
		if succNum > len(s.ipList)/2 {
			break // will block until the majority of nodes success
		}
	}

	//4. if the majority of nodes append the entries, commit the entry by calling its own metaStore.updateFile()
	ver, err := s.metaStore.UpdateFile(ctx, filemeta)
	if err != nil {
		return ver, err
	}
	s.commitIndex++

	//5. After commitment, in next heartbeat message, ask other nodes commit the entry too
	commitChan := make(chan *AppendEntryOutput, len(s.ipList)-1)
	// call AppendEntriesRPC
	for idx, addr := range s.ipList {
		if int64(idx) == s.serverId {
			continue
		}
		// concurrently call tryAppendEntries()
		go s.tryAppendEntries(addr, int64(len(s.log)-1), commitChan)
	}
	// receiving channels to see if the majority of nodes successfully append the entries
	succNum = 1
	for {
		output := <-commitChan
		if output != nil && output.Success {
			succNum++
		}
		if succNum > len(s.ipList)/2 {
			break // will block until the majority of nodes success
		}
	}

	return ver, nil
}

func (s *RaftSurfstore) tryAppendEntries(addr string, entryIdx int64, outputChan chan *AppendEntryOutput) {
	for {
		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return
		}
		client := NewRaftSurfstoreClient(conn)

		var prevLogTerm int64
		var prevLogIdx = entryIdx - 1
		if entryIdx == 0 {
			prevLogIdx = 0
			prevLogTerm = 0
		} else {
			prevLogTerm = s.log[entryIdx-1].Term
		}
		input := &AppendEntryInput{
			Term:         s.log[entryIdx].Term,
			PrevLogIndex: prevLogIdx,
			PrevLogTerm:  prevLogTerm,
			Entries:      s.log[:entryIdx+1],
			LeaderCommit: s.commitIndex,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		output, _ := client.AppendEntries(ctx, input)

		if output != nil {
			if output.Success {
				outputChan <- output
				return
			}
		}
	}
}

// AppendEntries 1. Reply false if term < currentTerm (§5.1)
// 2. Reply false if log doesn’t contain an entry at prevLogIndex whose term
// matches prevLogTerm (§5.3)
// 3. If an existing entry conflicts with a new one (same index but different
// terms), delete the existing entry and all that follow it (§5.3)
// 4. Append any new entries not already in the log
// 5. If leaderCommit > commitIndex, set commitIndex = min(leaderCommit, index
// of last new entry)
func (s *RaftSurfstore) AppendEntries(ctx context.Context, input *AppendEntryInput) (*AppendEntryOutput, error) {
	// Check if the server is crashed
	if isCrashed := s.isCrashed; isCrashed {
		return &AppendEntryOutput{
			ServerId: s.serverId,
			Success:  false,
		}, ERR_SERVER_CRASHED
	}

	//1. Reply false if term < currentTerm
	if input.Term < s.term {
		return &AppendEntryOutput{
			ServerId: s.serverId,
			Success:  false,
		}, nil
	} else if input.Term > s.term {
		s.term = input.Term
		s.isLeader = false
	}

	//2. if the follower does not find an entry in its log with the same index and term, then it refuses the new entries
	// search the log list if cannot find an entry matched with both prev index and prev term, return false
	flag := false
	matchedIndex := -1
	if input.PrevLogIndex != 0 && input.PrevLogTerm != 0 {
		for index, entry := range s.log {
			if int64(index) == input.PrevLogIndex && entry.Term == input.PrevLogTerm {
				flag = true
				matchedIndex = index
				break
			}
		}
	} else {
		flag = true
	}

	if !flag {
		return &AppendEntryOutput{
			ServerId:     s.serverId,
			Success:      false,
			MatchedIndex: 0, //TODO if matchedIndex == 0, previndex-- (in calling func)
			Term:         s.term,
		}, nil
	}

	//3. If an existing entry conflicts with a new one (same index but different terms),
	// delete the existing entry and all that folow it
	if matchedIndex != -1 {
		input.Entries = input.Entries[matchedIndex+1:]
	}
	if matchedIndex != len(s.log)-1 && input.Entries != nil {
		// delete all entries after matchedIndex
		s.log = s.log[:matchedIndex+1]
	}

	//4. Append any new entries not already in the log
	s.log = append(s.log, input.Entries...)

	//5. If leaderCommit > commitIndex, set commitIndex = min(leaderCommit, index of last new entry)
	if input.LeaderCommit > s.commitIndex {
		currCommitIdx := s.commitIndex
		s.commitIndex = int64(math.Min(float64(input.LeaderCommit), float64(len(s.log)-1))) // assume commit will success !!!

		for currCommitIdx < s.commitIndex {
			// commit the entry
			currCommitIdx++
			tempEntry := s.log[currCommitIdx]
			s.metaStore.UpdateFile(ctx, tempEntry.FileMetaData) // calling the server's UpdateFile, not the client's one
		}
	}

	output := AppendEntryOutput{
		ServerId:     s.serverId,
		Term:         s.term,
		Success:      true,
		MatchedIndex: int64(len(s.log) - 1),
	}

	return &output, nil
}

// SetLeader This should set the leader status and any related variables as if the node has just won an election
func (s *RaftSurfstore) SetLeader(ctx context.Context, _ *emptypb.Empty) (*Success, error) {
	if isCrashed := s.isCrashed; isCrashed {
		return &Success{Flag: false}, ERR_SERVER_CRASHED
	}

	s.term++
	s.isLeader = true
	return &Success{Flag: true}, nil
}

// SendHeartbeat Send a 'Heartbeat" (AppendEntries with no log entries) to the other servers
// Only leaders send heartbeats, if the node is not the leader you can return Success = false
func (s *RaftSurfstore) SendHeartbeat(ctx context.Context, _ *emptypb.Empty) (*Success, error) {
	if isCrashed := s.isCrashed; isCrashed {
		return &Success{Flag: false}, ERR_SERVER_CRASHED
	}

	if isLeader := s.isLeader; !isLeader {
		return &Success{Flag: false}, ERR_NOT_LEADER
	}

	ifMajorityAlive := false
	aliveNums := 1
	for idx, addr := range s.ipList {
		if int64(idx) == s.serverId {
			continue
		}

		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return &Success{Flag: false}, nil
		}
		client := NewRaftSurfstoreClient(conn)

		var prevTerm int64
		var prevIdx int64
		if len(s.log) == 0 {
			prevTerm = int64(0)
			prevIdx = int64(0)
		} else {
			prevTerm = s.log[len(s.log)-1].Term
			prevIdx = int64(len(s.log) - 1)
		}
		currInput := &AppendEntryInput{
			Term:         s.term,
			PrevLogIndex: prevIdx,
			PrevLogTerm:  prevTerm,
			Entries:      nil,
			LeaderCommit: s.commitIndex,
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		appendEntryOutput, _ := client.AppendEntries(ctx, currInput)
		if appendEntryOutput != nil {
			aliveNums++
			if aliveNums > len(s.ipList)/2 {
				ifMajorityAlive = true
			}
		}

	}
	return &Success{Flag: ifMajorityAlive}, nil
}

func (s *RaftSurfstore) Crash(ctx context.Context, _ *emptypb.Empty) (*Success, error) {
	s.isCrashedMutex.Lock()
	s.isCrashed = true
	s.isCrashedMutex.Unlock()

	return &Success{Flag: true}, nil
}

func (s *RaftSurfstore) Restore(ctx context.Context, _ *emptypb.Empty) (*Success, error) {
	s.isCrashedMutex.Lock()
	s.isCrashed = false
	s.notCrashedCond.Broadcast()
	s.isCrashedMutex.Unlock()

	return &Success{Flag: true}, nil
}

func (s *RaftSurfstore) IsCrashed(ctx context.Context, _ *emptypb.Empty) (*CrashedState, error) {
	return &CrashedState{IsCrashed: s.isCrashed}, nil
}

func (s *RaftSurfstore) GetInternalState(ctx context.Context, empty *emptypb.Empty) (*RaftInternalState, error) {
	fileInfoMap, _ := s.metaStore.GetFileInfoMap(ctx, empty)
	return &RaftInternalState{
		IsLeader: s.isLeader,
		Term:     s.term,
		Log:      s.log,
		MetaMap:  fileInfoMap,
	}, nil
}

var _ RaftSurfstoreInterface = new(RaftSurfstore)
