export GOPATH=$GOPATH:~/cse223b-RAFT-KV-STORE/
rm -rf ~/cse223b-RAFT-KV-STORE/
source /etc/profile
cd cse223b-RAFT-KV-STORE/
go build src/main/raftkv_main.go
exit

