package SurfTest

import (
	context "context"
	"cse224/proj5/pkg/surfstore"
	"google.golang.org/grpc"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"
)

type TestInfo struct {
	CfgPath    string
	Ips        []string
	Context    context.Context
	CancelFunc context.CancelFunc
	Procs      []*exec.Cmd
	Conns      []*grpc.ClientConn
	Clients    []surfstore.RaftSurfstoreClient
}

func InitTest(cfgPath, blockStorePort string) TestInfo {
	cfg := surfstore.LoadRaftConfigFile(cfgPath)

	procs := make([]*exec.Cmd, 0)
	procs = append(procs, InitBlockStore(blockStorePort))
	procs = append(procs, InitRaftServers(cfgPath)...)

	conns := make([]*grpc.ClientConn, 0)
	clients := make([]surfstore.RaftSurfstoreClient, 0)
	for _, addr := range cfg {
		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			log.Fatal("Error connecting to clients ", err)
		}
		client := surfstore.NewRaftSurfstoreClient(conn)

		conns = append(conns, conn)
		clients = append(clients, client)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)

	return TestInfo{
		CfgPath:    cfgPath,
		Ips:        cfg,
		Context:    ctx,
		CancelFunc: cancel,
		Procs:      procs,
		Conns:      conns,
		Clients:    clients,
	}
}

func EndTest(test TestInfo) {
	test.CancelFunc()

	for _, conn := range test.Conns {
		conn.Close()
	}

	for _, server := range test.Procs {
		_ = server.Process.Kill()
	}

	exec.Command("pkill SurfstoreRaftServerExec*")

	time.Sleep(100 * time.Millisecond)
}

func InitBlockStore(blockStorePort string) *exec.Cmd {
	blockCmd := exec.Command("_bin/SurfstoreServerExec", "-s", "block", "-p", blockStorePort, "-l")
	blockCmd.Stderr = os.Stderr
	blockCmd.Stdout = os.Stdout
	err := blockCmd.Start()
	if err != nil {
		log.Fatal("Error starting BlockStore ", err)
	}

	return blockCmd
}

func InitRaftServers(cfgPath string) []*exec.Cmd {
	cfg := surfstore.LoadRaftConfigFile(cfgPath)
	cmdList := make([]*exec.Cmd, 0)
	for idx, _ := range cfg {
		cmd := exec.Command("_bin/SurfstoreRaftServerExec", "-f", cfgPath, "-i", strconv.Itoa(idx), "-b", "localhost:8080")
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmdList = append(cmdList, cmd)
	}

	for _, cmd := range cmdList {
		err := cmd.Start()
		if err != nil {
			log.Fatal("Error starting servers", err)
		}
	}

	time.Sleep(2 * time.Second)

	return cmdList
}

func SameOperation(op1, op2 *surfstore.UpdateOperation) bool {
	if op1 == nil && op2 == nil {
		return true
	}
	if op1 == nil || op2 == nil {
		return false
	}
	if op1.Term != op2.Term {
		return false
	}
	if op1.FileMetaData == nil && op2.FileMetaData != nil ||
		op1.FileMetaData != nil && op2.FileMetaData == nil {
		return false
	}
	if op1.FileMetaData.Version != op2.FileMetaData.Version {
		return false
	}
	if !SameHashList(op1.FileMetaData.BlockHashList, op2.FileMetaData.BlockHashList) {
		return false
	}
	return true
}

func SameLog(log1, log2 []*surfstore.UpdateOperation) bool {
	if len(log1) != len(log2) {
		return false
	}
	for idx, entry1 := range log1 {
		entry2 := log2[idx]
		if !SameOperation(entry1, entry2) {
			return false
		}
	}
	return true
}
