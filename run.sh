export GOPATH=$GOPATH:~/cse223b-RAFT-KV-STORE/
cd cse223b-RAFT-KV-STORE/
nohup ./raftkv_main #node_num &
ps -ef | grep raft
exit