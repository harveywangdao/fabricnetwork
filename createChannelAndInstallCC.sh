#!/bin/bash
#
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# Language defaults to "golang"
LANGUAGE="golang"

##set chaincode path
function setChaincodePath(){
    LANGUAGE=`echo "$LANGUAGE" | tr '[:upper:]' '[:lower:]'`
    case "$LANGUAGE" in
        "golang")
        CC_SRC_PATH="github.com/ocean/go"
        ;;
        "node")
        CC_SRC_PATH="$PWD/artifacts/src/github.com/ocean/node"
        ;;
        *) printf "\n ------ Language $LANGUAGE is not supported yet ------\n"$
        exit 1
    esac
}

setChaincodePath

echo "POST request Register Org1 ..."
echo
curl -s -X POST \
  http://localhost:4000/users \
  -H "content-type: application/json" \
  -d '{
    "username":"thomas",
    "orgname":"Org1"
}'
echo

echo "POST request Register Org2 ..."
echo
curl -s -X POST \
  http://localhost:4000/users \
  -H "content-type: application/json" \
  -d '{
    "username":"anli",
    "orgname":"Org2"
}'
echo

echo "POST request Create channel  ..."
echo
curl -s -X POST \
  http://localhost:4000/channels \
  -H "content-type: application/json" \
  -d '{
    "channelName":"mychannel",
    "channelConfigPath":"../artifacts/channel/mychannel.tx"
}'
echo
echo
sleep 5

echo "POST request Join channel on Org1"
echo
curl -s -X POST \
  http://localhost:4000/channels/mychannel/peers \
  -H "content-type: application/json" \
  -d '{
    "username":"thomas",
    "orgname":"Org1",
    "peers": ["peer0.org1.example.com","peer1.org1.example.com"]
}'
echo
echo

echo "POST request Join channel on Org2"
echo
curl -s -X POST \
  http://localhost:4000/channels/mychannel/peers \
  -H "content-type: application/json" \
  -d '{
    "username":"anli",
    "orgname":"Org2",
    "peers": ["peer0.org2.example.com","peer1.org2.example.com"]
}'
echo
echo

echo "POST Install chaincode on Org1"
echo
curl -s -X POST \
  http://localhost:4000/chaincodes \
  -H "content-type: application/json" \
  -d "{
    \"username\":\"thomas\",
    \"orgname\":\"Org1\",
    \"peers\": [\"peer0.org1.example.com\",\"peer1.org1.example.com\"],
    \"chaincodeName\":\"mycc\",
    \"chaincodePath\":\"$CC_SRC_PATH\",
    \"chaincodeType\": \"$LANGUAGE\",
    \"chaincodeVersion\":\"v0\"
}"
echo
echo

echo "POST Install chaincode on Org2"
echo
curl -s -X POST \
  http://localhost:4000/chaincodes \
  -H "content-type: application/json" \
  -d "{
    \"username\":\"anli\",
    \"orgname\":\"Org2\",
    \"peers\": [\"peer0.org2.example.com\",\"peer1.org2.example.com\"],
    \"chaincodeName\":\"mycc\",
    \"chaincodePath\":\"$CC_SRC_PATH\",
    \"chaincodeType\": \"$LANGUAGE\",
    \"chaincodeVersion\":\"v0\"
}"
echo
echo

echo "POST instantiate chaincode on Org1"
echo
curl -s -X POST \
  http://localhost:4000/channels/mychannel/chaincodes \
  -H "content-type: application/json" \
  -d "{
    \"username\":\"thomas\",
    \"orgname\":\"Org1\",
    \"chaincodeName\":\"mycc\",
    \"chaincodeVersion\":\"v0\",
    \"chaincodeType\": \"$LANGUAGE\",
    \"args\":[\"a\",\"100\",\"b\",\"200\"]
}"
echo
echo

echo "POST invoke chaincode on peers of Org1 and Org2"
echo
TRX_ID=$(curl -s -X POST \
  http://localhost:4000/channels/mychannel/chaincodes/mycc \
  -H "content-type: application/json" \
  -d '{
    "peers": ["peer0.org1.example.com","peer0.org2.example.com"],
    "fcn":"move",
    "args":["a","b","10"]
}')
echo "Transaction ID is $TRX_ID"
echo
echo

echo "GET query chaincode on peer1 of Org1"
echo
curl -s -X GET \
  "http://localhost:4000/channels/mychannel/chaincodes/mycc?peer=peer0.org1.example.com&fcn=query&args=%5B%22a%22%5D" \
  -H "content-type: application/json"
echo
echo
