#pssh -h cloudlab -A -P -I<./setup-node.sh
#pssh -i -h cloudlab sudo apt install -y protobuf-compiler
#pssh -i -h cloudlab sudo apt install -y golang-go
#pssh -i -h cloudlab go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
#pssh -i -h cloudlab go version
#pscp -h cloudlab /Users/pranaymundra/Downloads/localgraph-dp-cloudlab.zip /users/pmundra/
pssh -i -h cloudlab unzip localgraph-dp-cloudlab.zip && sudo rm *.zip
#pssh -i -h cloudlab
#pssh -i -h cloudlab