# cloudfederation_metering
a prototype system about cloud federation metering based on Hyperledger Fabric
# Cloud Federation Metering

本项目是基于联盟链的科研云联邦计量原型系统的部分代码开源，该原型系统底层基于Hyperledger Fabric框架研发，智能合约使用Goland开发，上层采用Goland、Python(Django)框架开发，当前原型系统正在进一步开发中。

## 说明

本系统部署目前指定3个组织（Org1、Org2、Org3），每个组织2个peer节点（Peer0、Peer1），3个排序节点（Orderer1、Orderer2、Orderer3）、构建4个应用通道（channelall、channel12、channel13、channel23），保证计量数据的隐私与安全。
1. 下载代码后修改configtx.yaml 文件中的msp/raft路径
2. 生成创世块 在指定路径下输出创世块文件
    ```
    configtxgen -profile ThreeOrgsOrdererGenesis -outputBlock ./channel-artifacts/genesis.block -channelID fabric-channelsys
    ```
3. 生成通道文件
    ```
    configtxgen -profile ThreeOrgsChannel -outputCreateChannelTx ./channel-artifacts/channelall.tx -channelID channelall
    configtxgen -profile Channel12 -outputCreateChannelTx ./channel-artifacts/channel12.tx -channelID channel12
    configtxgen -profile Channel13 -outputCreateChannelTx ./channel-artifacts/channel13.tx -channelID channel13
    configtxgen -profile Channel23 -outputCreateChannelTx ./channel-artifacts/channel23.tx -channelID channel23
    ```
4. 生成锚节点文件
    ```
    configtxgen -profile ThreeOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org1MSPanchors.tx -channelID channelall -asOrg Org1MSP

    configtxgen -profile ThreeOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org2MSPanchors.tx -channelID channelall -asOrg Org2MSP

    configtxgen -profile ThreeOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org3MSPanchors.tx -channelID channelall -asOrg Org3MSP

    configtxgen -profile Channel12 -outputAnchorPeersUpdate ./channel-artifacts/Org1MSPanchors12.tx -channelID channel12 -asOrg Org1MSP
    configtxgen -profile Channel12 -outputAnchorPeersUpdate ./channel-artifacts/Org2MSPanchors12.tx -channelID channel12 -asOrg Org2MSP

    configtxgen -profile Channel13 -outputAnchorPeersUpdate ./channel-artifacts/Org1MSPanchors13.tx -channelID channel13 -asOrg Org1MSP
    configtxgen -profile Channel13 -outputAnchorPeersUpdate ./channel-artifacts/Org3MSPanchors13.tx -channelID channel13 -asOrg Org3MSP

    configtxgen -profile Channel23 -outputAnchorPeersUpdate ./channel-artifacts/Org2MSPanchors23.tx -channelID channel23 -asOrg Org2MSP
    configtxgen -profile Channel23 -outputAnchorPeersUpdate ./channel-artifacts/Org3MSPanchors23.tx -channelID channel23 -asOrg Org3MSP
    ```
5. 修改docker-compose文件中的路径名、网络名称；用于利用容器创建peer节点和客户端节点
6. 启动网络
    ```
    Docker-compose up -d
    ```
    (停掉网络 使用 docker-compose down
    删除挂在内容 使用 docker volume prune)
7. 进入终端节点创建通道
    ```
    docker exec -it cli1 bash
    peer channel create -o orderer2.example.com:7050 -c channel12 -f ./channel-artifacts/channel12.tx --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/msp/tlscacerts/tlsca.example.com-cert.pem
    peer channel create -o orderer.example.com:7050 -c channel13 -f ./channel-artifacts/channel13.tx --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/msp/tlscacerts/tlsca.example.com-cert.pem
    peer channel create -o orderer.example.com:7050 -c channel23 -f ./channel-artifacts/channel23.tx --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/msp/tlscacerts/tlsca.example.com-cert.pem
    ```
    退出终端
8. 复制区块文件到同通道的另一个节点
    ```
    docker cp cli1:/opt/gopath/src/github.com/hyperledger/fabric/peer/channel12.block ./
    docker cp ./channel12.block cli2:/opt/gopath/src/github.com/hyperledger/fabric/peer

    docker cp cli1:/opt/gopath/src/github.com/hyperledger/fabric/peer/channel13.block ./
    docker cp ./channel13.block cli3:/opt/gopath/src/github.com/hyperledger/fabric/peer

    docker cp cli2:/opt/gopath/src/github.com/hyperledger/fabric/peer/channel23.block ./
    docker cp ./channel23.block cli3:/opt/gopath/src/github.com/hyperledger/fabric/peer
    ```

9. 同时进入两个终端 加入通道（以channel12为例）
    ```
    peer channel join -b channel12.block
    ```

10. 更新各自锚节点

    【组织1】
    ```
    peer channel update -o orderer1.example.com:7050 -c channel12 -f ./channel-artifacts/Org1MSPanchors12.tx --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer1.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
    ```
    【组织2】
    ```
    peer channel update -o orderer1.example.com:7050 -c channel12 -f ./channel-artifacts/Org2MSPanchors12.tx --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer1.example.com/msp/tlscacerts/tlsca.example.
    com-cert.pem
    ```

11. 链码相关操作

    
    ```
    # 进入终端节点1
    Docker exec -it cli1 bash
    Cd /opt/gopath/src/github.com/hyperledger/fabric-cluster/chaincode/go
    # 安装依赖
    go env -w GOPROXY=https://goproxy.cn,direct
    go mod init
    go mod vendor
    # 回到工作目录 打包链码
    cd /opt/gopath/src/github.com/hyperledger/fabric/peer/
    peer lifecycle chaincode package mychaincode_channel12.tar.gz \
    --path github.com/hyperledger/fabric-cluster/chaincode/go/ \
    --label mychaincode_channel12
    # 复制到cli2的节点终端里面【此处是退出终端1 在本机操作的】
    docker cp cli1:/opt/gopath/src/github.com/hyperledger/fabric/peer/mychaincode_channel12.tar.gz ./
    docker cp mychaincode_channel12.tar.gz cli2:/opt/gopath/src/github.com/hyperledger/fabric/peer
    # 安装链码【两个节点都要安装】
    peer lifecycle chaincode install mychaincode_channel12.tar.gz
    # 组织批准【在组织1、2上都要执行批准 approve】
    # 通道ID 链码名称、版本号、是否需要初始化、包ID（上面已经给出【也能用命令查package-id 【peer lifecycle chaincode queryinstalled】）、序列号、tls、证书文件目录
    peer lifecycle chaincode approveformyorg --channelID channel12 --name mychaincode_channel12 --version 1.0 --package-id mychaincode_channel12:60b780ba05898d193a5375c3aea7b2b31253f9f37b37895b7f2416340797bf9c --sequence 1 --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer3.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
    # 查看是否批准成果
    peer lifecycle chaincode checkcommitreadiness --channelID channel12 -name mychaincode_channel12 --version 1.0 --sequence 1 --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem –output json
    
    # 提交链码到链上
    peer lifecycle chaincode commit -o orderer.example.com:7050 --channelID channel12 --name mychaincode_channel12 --version 1.0 --sequence 1 --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
    ```
 

12. 链码调用

    云服务方对不同的云服务使用方发布配额信息，包括最长服务持有时限、vCPU、内存、硬盘容量、IP个数等等内容。

    ```
    peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C channel12 -n mychaincode --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"Args":["initInitialQuota","000001","CNIC","00001","CNNIC","00002","365","10","40", "40", "100", "2","10","20"]}'
    ```
    
    云服务提供方公开发布价格信息，价格信息对所有用户透明，可以帮助用户选择合适的云资源，防止服务提供方之间恶意竞争。

    ```
    peer chaincode invoke -o orderer2.example.com:7050 --ordererTLSHostnameOverride orderer2.example.com --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer2.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C channel12 -n mychaincode_channel12 --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"Args":["initPrice" ,"CNIC","00001","10","4","1","1","1", "2", "0"]}'
    ```

    链码设计了读函数，用户可以查询本单位配额和所有提供方的价格信息。
    ```
    peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C channel12 -n mychaincode --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"Args":["readResource","quota","000001"]}'
    ```

    用户提交云主机资源申请，，链码调用initECSResource函数，上链信息包括提供方和用户双方信息，以及云主机的vCPU、内存、网络带宽等基础信息；也同样提供了读函数，方便用户调用查看资源信息。
    ```
    peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C channel12 -n mychaincode_021503 --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"Args":["initECSResource","202202150002","CNIC","00001","CNNIC","00002001","ECS","public","2","2","4", "8", "365"]}'
    ```
    用户使用资源完毕后，释放资源，系统自动调用billingECS函数，如图5.17所示，将信息上链计入账单，并生成相应的账单金额。
    ```
    peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C channel12 -n mychaincode_channel12 --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"Args":["billedECS","202202160001","00001","00002001"]}'
    ```
    管理员可以通过调用queryProviderBill和queryProviderTotalCost函数查看本单位的所有账单以及合计金额，方便后续的结算模块使用数据。
    ```
    peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C channel12 -n mychaincode_021504 --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"Args":["queryProviderBill","00001","ECSBill"]}' 
    ```
