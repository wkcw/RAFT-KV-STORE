export GOPATH=$GOPATH:~/cse223b-RAFT-KV-STORE/
rm -rf ~/cse223b-RAFT-KV-STORE/
git clone https://SwimmingFish6:lxhxhd123@github.com/wkcw/cse223b-RAFT-KV-STORE.git -b master
source /etc/profile
cd cse223b-RAFT-KV-STORE/
go build src/main/raftkv_main.go
exit

