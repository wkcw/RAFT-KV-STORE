package service

type raftConfig struct{
	heartbeatInterval int
	peers []string
	ID string
}