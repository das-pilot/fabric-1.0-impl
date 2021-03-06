#!/bin/bash

echo
echo " ____    _____      _      ____    _____ "
echo "/ ___|  |_   _|    / \    |  _ \  |_   _|"
echo "\___ \    | |     / _ \   | |_) |   | |  "
echo " ___) |   | |    / ___ \  |  _ <    | |  "
echo "|____/    |_|   /_/   \_\ |_| \_\   |_|  "
echo
echo "Build your first network (BYFN) end-to-end test"
echo
CHANNEL_NAME="$1"
: ${CHANNEL_NAME:="daschannel"}
: ${TIMEOUT:="60"}
COUNTER=1
MAX_RETRY=5
ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/das-pilot.com/orderers/orderer.das-pilot.com/msp/tlscacerts/tlsca.das-pilot.com-cert.pem
RED='\033[0;31m'
NC='\033[0m' # No Color
echo "Channel name : "$CHANNEL_NAME

# verify the result of the end-to-end test
verifyResult () {
	if [ $1 -ne 0 ] ; then
		echo "!!!!!!!!!!!!!!! "$2" !!!!!!!!!!!!!!!!"
    echo "========= ERROR !!! FAILED to execute End-2-End Scenario ==========="
		echo
   		exit 1
	fi
}

setGlobals () {

	if [ $1 -eq 0 -o $1 -eq 1 ] ; then
		CORE_PEER_LOCALMSPID="Org1MSP"
		CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.das-pilot.com/peers/peer0.org1.das-pilot.com/tls/ca.crt
		CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.das-pilot.com/users/Admin@org1.das-pilot.com/msp
		if [ $1 -eq 0 ]; then
			CORE_PEER_ADDRESS=peer0.org1.das-pilot.com:7051
		else
			CORE_PEER_ADDRESS=peer1.org1.das-pilot.com:7051
			CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.das-pilot.com/users/Admin@org1.das-pilot.com/msp
		fi
	else
		CORE_PEER_LOCALMSPID="Org2MSP"
		CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.das-pilot.com/peers/peer0.org2.das-pilot.com/tls/ca.crt
		CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.das-pilot.com/users/Admin@org2.das-pilot.com/msp
		if [ $1 -eq 2 ]; then
			CORE_PEER_ADDRESS=peer0.org2.das-pilot.com:7051
		else
			CORE_PEER_ADDRESS=peer1.org2.das-pilot.com:7051
		fi
	fi

	env |grep CORE
}

createChannel() {
	setGlobals 0

  if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
        echo "creating channel $CHANNEL_NAME"
		peer channel create -o orderer.das-pilot.com:7050 -c $CHANNEL_NAME -f ./channel-artifacts/channel.tx >&log.txt
	else
	    echo "creating channel $CHANNEL_NAME tls $CORE_PEER_TLS_ENABLED orderer ca $ORDERER_CA"
		peer channel create -o orderer.das-pilot.com:7050 -c $CHANNEL_NAME -f ./channel-artifacts/channel.tx --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA >&log.txt
	fi
	res=$?
	cat log.txt
	verifyResult $res "Channel creation failed"
	echo "===================== Channel \"$CHANNEL_NAME\" is created successfully ===================== "
	echo
}

updateAnchorPeers() {
  PEER=$1
  setGlobals $PEER

  if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
		peer channel update -o orderer.das-pilot.com:7050 -c $CHANNEL_NAME -f ./channel-artifacts/${CORE_PEER_LOCALMSPID}anchors.tx >&log.txt
	else
		peer channel update -o orderer.das-pilot.com:7050 -c $CHANNEL_NAME -f ./channel-artifacts/${CORE_PEER_LOCALMSPID}anchors.tx --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA >&log.txt
	fi
	res=$?
	cat log.txt
	verifyResult $res "Anchor peer update failed"
	echo "===================== Anchor peers for org \"$CORE_PEER_LOCALMSPID\" on \"$CHANNEL_NAME\" is updated successfully ===================== "
	echo
}

## Sometimes Join takes time hence RETRY atleast for 5 times
joinWithRetry () {
	peer channel join -b $CHANNEL_NAME.block  >&log.txt
	res=$?
	cat log.txt
	if [ $res -ne 0 -a $COUNTER -lt $MAX_RETRY ]; then
		COUNTER=` expr $COUNTER + 1`
		echo "PEER$1 failed to join the channel, Retry after 2 seconds"
		sleep 2
		joinWithRetry $1
	else
		COUNTER=1
	fi
  verifyResult $res "After $MAX_RETRY attempts, PEER$ch has failed to Join the Channel"
}

joinChannel () {
	for ch in 0 1 2 3; do
		setGlobals $ch
		joinWithRetry $ch
		echo "===================== PEER$ch joined on the channel \"$CHANNEL_NAME\" ===================== "
		sleep 2
		echo
	done
}

installChaincode () {
	PEER=$1
	setGlobals $PEER
	peer chaincode install -n wallet -v 1.0 -p github.com/hyperledger/fabric/chaincodes/go/wallet >&log.txt
	res=$?
	cat log.txt
        verifyResult $res "Chaincode installation on remote peer PEER$PEER has Failed"
	echo "===================== Chaincode is installed on remote peer PEER$PEER ===================== "
	echo
}

instantiateChaincode () {
	PEER=$1
	setGlobals $PEER
	# while 'peer chaincode' command can get the orderer endpoint from the peer (if join was successful),
	# lets supply it directly as we know it using the "-o" option
	if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
		peer chaincode instantiate -o orderer.das-pilot.com:7050 -C $CHANNEL_NAME -n wallet -v 1.0 -c '{"Args":["init""]}' -P "OR	('Org1MSP.member','Org2MSP.member')" >&log.txt
	else
		peer chaincode instantiate -o orderer.das-pilot.com:7050 --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA -C $CHANNEL_NAME -n wallet -v 1.0 -c '{"Args":["init"]}' -P "OR	('Org1MSP.member','Org2MSP.member')" >&log.txt
	fi
	res=$?
	cat log.txt
	verifyResult $res "Chaincode instantiation on PEER$PEER on channel '$CHANNEL_NAME' failed"
	echo "===================== Chaincode Instantiation on PEER$PEER on channel '$CHANNEL_NAME' is successful ===================== "
	echo
}


chaincodeQuery () {
  PEER=$1
  SOURCE=$2
  DESTINATION=$3
  echo "===================== Querying on PEER$PEER on channel '$CHANNEL_NAME'... ===================== "
  setGlobals $PEER
  local rc=1
  local starttime=$(date +%s)

  # continue to poll
  # we either get a successful response, or reach TIMEOUT
  while test "$(($(date +%s)-starttime))" -lt "$TIMEOUT" -a $rc -ne 0
  do
     sleep 3
     echo "Attempting to Query PEER$PEER ...$(($(date +%s)-starttime)) secs"
     peer chaincode query -C $CHANNEL_NAME -n wallet -c '{"Args":["query","'${SOURCE}'","'${DESTINATION}'"]}' >&log.txt
     #test $? -eq 0 && VALUE=$(cat log.txt | awk '/Query Result/ {print $NF}')
     VALUE=$(cat log.txt | awk '/Query Result/ {print $NF}')
     printf "${RED}Balance for ${SOURCE}/${DESTINATION} is: ${VALUE}${NC}"
     let rc=0
  done
  echo
  cat log.txt
  if test $rc -eq 0 ; then
	echo "===================== Query on PEER$PEER on channel '$CHANNEL_NAME' is successful ===================== "
  else
	echo "!!!!!!!!!!!!!!! Query result on PEER$PEER is INVALID !!!!!!!!!!!!!!!!"
        echo "================== ERROR !!! FAILED to execute End-2-End Scenario =================="
	echo
	exit 1
  fi
}

chaincodeQueryHistory () {
  PEER=$1
  SOURCE=$2
  DESTINATION=$3
  echo "===================== Querying on PEER$PEER on channel '$CHANNEL_NAME'... ===================== "
  setGlobals $PEER
  local rc=1
  local starttime=$(date +%s)

  # continue to poll
  # we either get a successful response, or reach TIMEOUT
  while test "$(($(date +%s)-starttime))" -lt "$TIMEOUT" -a $rc -ne 0
  do
     sleep 3
     echo "Attempting to Query PEER$PEER ...$(($(date +%s)-starttime)) secs"
     peer chaincode query -C $CHANNEL_NAME -n wallet -c '{"Args":["queryHistory","'${SOURCE}'","'${DESTINATION}'"]}' >&log.txt
     #test $? -eq 0 && VALUE=$(cat log.txt | awk '/Query Result/ {print $NF}')
     VALUE=$(cat log.txt | grep 'Query Result')
     printf "${RED}History for ${SOURCE}/${DESTINATION} is: ${VALUE}${NC}"
     let rc=0
  done
  echo
  cat log.txt
  if test $rc -eq 0 ; then
	echo "===================== Query on PEER$PEER on channel '$CHANNEL_NAME' is successful ===================== "
  else
	echo "!!!!!!!!!!!!!!! Query result on PEER$PEER is INVALID !!!!!!!!!!!!!!!!"
        echo "================== ERROR !!! FAILED to execute End-2-End Scenario =================="
	echo
	exit 1
  fi
}

chainCodeCreateWallet () {
	PEER=$1
    WALLET=$2
	setGlobals $PEER
	# while 'peer chaincode' command can get the orderer endpoint from the peer (if join was successful),
	# lets supply it directly as we know it using the "-o" option
	if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
		peer chaincode invoke -o orderer.das-pilot.com:7050 -C $CHANNEL_NAME -n wallet -c '{"Args":["create","'${WALLET}'"]}' >&log.txt
	else
		peer chaincode invoke -o orderer.das-pilot.com:7050  --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA -C $CHANNEL_NAME -n wallet -c '{"Args":["create","'${WALLET}'"]}' >&log.txt
	fi
	res=$?
	cat log.txt
	verifyResult $res "Invoke execution on PEER$PEER failed "
	echo "===================== Invoke transaction on PEER$PEER on channel '$CHANNEL_NAME' is successful ===================== "
	echo
}

chainCodeCharge () {
	PEER=$1
	CHARGE_FROM=$2
	CHARGE_TO=$3
	AMOUNT=$4
	setGlobals $PEER
	# while 'peer chaincode' command can get the orderer endpoint from the peer (if join was successful),
	# lets supply it directly as we know it using the "-o" option
	if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
		peer chaincode invoke -o orderer.das-pilot.com:7050 -C $CHANNEL_NAME -n wallet -c '{"Args":["charge","'${CHARGE_FROM}'","'${CHARGE_TO}'","'${AMOUNT}'"]}' >&log.txt
	else
		peer chaincode invoke -o orderer.das-pilot.com:7050  --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA -C $CHANNEL_NAME -n wallet -c '{"Args":["charge","'${CHARGE_FROM}'","'${CHARGE_TO}'","'${AMOUNT}'"]}' >&log.txt
	fi
	res=$?
	cat log.txt
	verifyResult $res "Invoke execution on PEER$PEER failed "
	echo "===================== Invoke transaction on PEER$PEER on channel '$CHANNEL_NAME' is successful ===================== "
	echo
}
## Create channel
echo "Creating channel..."
createChannel

## Join all the peers to the channel
echo "Having all peers join the channel..."
joinChannel

## Set the anchor peers for each org in the channel
echo "Updating anchor peers for org1..."
updateAnchorPeers 0
echo "Updating anchor peers for org2..."
updateAnchorPeers 2

## Install chaincode on Peer0/Org1 and Peer2/Org2
echo "Installing chaincode on org1/peer0..."
installChaincode 0
echo "Installing chaincode on org2/peer0..."
installChaincode 2
#Instantiate chaincode on Peer2/Org2
echo "Instantiating chaincode on org1/peer0..."
instantiateChaincode 0
echo "Query 'total_amount'"
chaincodeQuery 0 "total_amount" "one"
echo "Creating wallet 'one'"
chainCodeCreateWallet 0 "one"
echo "Query chaincode on org1/peer0..."
chaincodeQuery 0 "one" "two"
echo "Query chaincode on org2/peer0..."
chaincodeQuery 2 "one" "two"
echo "Create wallet 'two' on org2/peer0..."
chainCodeCreateWallet 2 "two"
echo "Query wallet 'two'"
chaincodeQuery 2 "two" "one"
echo "Charge 'one'->'two' chaincode on org1/peer0..."
chainCodeCharge 0 "one" "two" "9.99"
echo "Query wallet 'one' on org2/peer0..."
chaincodeQuery 2 "one" "two"
echo "Charge 'two'->'one"
chainCodeCharge 2 "two" "one" "3.33"
sleep 5
chainCodeCharge 2 "two" "one" "3.33"
sleep 5
#chaincodeQuery 2 "one" "two"
#chainCodeCharge 2 "two" "one" "3.33"
#sleep 5
chaincodeQuery 2 "one" "two"

chaincodeQueryHistory 2 "one" "two"
echo
echo "========= All GOOD, BYFN execution completed =========== "
echo
echo
echo
echo " _____   _   _   ____   "
echo "| ____| | \ | | |  _ \  "
echo "|  _|   |  \| | | | | | "
echo "| |___  | |\  | | |_| | "
echo "|_____| |_| \_| |____/  "
echo
#/opt/gopath/src/github.com/hyperledger/fabric/block-listener/listener \
#  -events-address=peer0.org1.das-pilot.com:7053 \
#  -events-mspdir=$CORE_PEER_MSPCONFIGPATH \
#  -events-mspid=Org1MSP \
#  -update-reciever-url=127.0.0.1:8080 > /opt/gopath/src/github.com/hyperledger/fabric/block-listener/listener.log
#exit 0
