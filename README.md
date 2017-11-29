# fabric-1.0-impl
Setup of sample hyperledger network, with channels and smart contracts and REST API container

#### Prerequisits
1. Docker
2. docker-compose utility
3. GO

#### Setup
1. clone sources of fabric-1.0-impl and das-go into you $GOPATH so the folders should be $GOPATH/src/github.com/das-pilot/fabric-1.0-impl/ and $GOPATH/src/github.com/das-pilot/das-go
2. execute following:
```
cd $GOPATH/src/github.com/das-pilot/fabric-1.0-impl/network
chmod +x ../../das-go/start.sh
chmod +x *.sh */*.sh
./get-binaries.sh
cd bin
sudo ./get-docker-images.sh
cd ../
sudo docker pull hyperledger/fabric-ca:x86_64-1.0.0
sudo ./setup.sh -m tag
./setup.sh -m gopkgs
./setup.sh -m generate
sudo ./setup.sh -m up
```