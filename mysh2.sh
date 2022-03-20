#!/bin/bash
echo "Hello World !"

docker cp cli1:/opt/gopath/src/github.com/hyperledger/fabric/peer/mychaincode_channel12.tar.gz ./
docker cp mychaincode_channel12.tar.gz cli2:/opt/gopath/src/github.com/hyperledger/fabric/peer
docker exec cli1 peer lifecycle chaincode install mychaincode_channel12.tar.gz
docker exec cli2 peer lifecycle chaincode install mychaincode_channel12.tar.gz

docker exec cli1 peer lifecycle chaincode approveformyorg --channelID channel12 --name mychaincode_channel12 --version 1.0 --package-id mychaincode_channel12:60b780ba05898d193a5375c3aea7b2b31253f9f37b37895b7f2416340797bf9c --sequence 1 --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer1.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
docker exec cli2 peer lifecycle chaincode approveformyorg --channelID channel12 --name mychaincode_channel12 --version 1.0 --package-id mychaincode_channel12:60b780ba05898d193a5375c3aea7b2b31253f9f37b37895b7f2416340797bf9c --sequence 1 --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer1.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

docker exec cli1 peer lifecycle chaincode commit -o orderer1.example.com:7050 --channelID channel12 --name mychaincode_channel12 --version 1.0 --sequence 1 --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer1.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt

docker exec cli1 peer chaincode invoke -o orderer1.example.com:7050 --ordererTLSHostnameOverride orderer1.example.com --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer1.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C channel12 -n mychaincode_channel12 --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"Args":["initPrice" ,"CNIC","00001","10","4","1","1","1", "2", "0"]}'
