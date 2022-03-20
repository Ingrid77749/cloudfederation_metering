#!/bin/bash
echo "Hello World !"
docker-compose up -d
docker exec cli1 peer channel create -o orderer1.example.com:7050 -c channel13 -f ./channel-artifacts/channel13.tx --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/msp/tlscacerts/tlsca.example.com-cert.pem
docker cp cli1:/opt/gopath/src/github.com/hyperledger/fabric/peer/channel13.block ./
docker cp ./channel13.block cli3:/opt/gopath/src/github.com/hyperledger/fabric/peer

docker exec cli1 peer channel join -b channel13.block
docker exec cli1 peer channel update -o orderer1.example.com:7050 -c channel13 -f ./channel-artifacts/Org1MSPanchors13.tx --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer1.example.com/msp/tlscacerts/tlsca.example.com-cert.pem


docker exec cli3 peer channel join -b channel13.block
docker exec cli3 peer channel update -o orderer1.example.com:7050 -c channel13 -f ./channel-artifacts/Org3MSPanchors13.tx --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer1.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

