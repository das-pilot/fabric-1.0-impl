# fabric-1.0-impl
Setup of sample hyperledger network, with channels and smart contracts and REST API container

#### Prerequisits
1. Docker
2. docker-compose utility
3. GO

#### Setup
1. clone sources of fabric-1.0-impl & das-go into you $GOPATH so the folders should be $GOPATH/src/github.com/fabric-1.0-impl/ & $GOPATH/src/github.com/das-go
2. execute following:
```
cd $GOPATH/src/github.com/fabric-1.0-impl/network
./get-binaries.sh
cd bin
sudo ./get-docker-images.sh
cd ../
sudo docker pull hyperledger/fabric-ca:x86_64-1.0.0
sudo ./setup.sh -m tag

```