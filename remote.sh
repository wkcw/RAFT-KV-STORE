rm -rf ~/cse223b-RAFT-KV-STORE/
git clone https://github.com/wkcw/cse223b-RAFT-KV-STORE.git -b first 
source /etc/profile
cd ~/cse223b-RAFT-KV-STORE/src/main
/usr/local/go/bin/go build server_main.go
./server_main $element:9527 & 
ps -ef | grep server
exit

