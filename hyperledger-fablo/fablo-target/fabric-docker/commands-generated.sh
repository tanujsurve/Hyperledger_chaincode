#!/usr/bin/env bash

generateArtifacts() {
  printHeadline "Generating basic configs" "U1F913"

  printItalics "Generating crypto material for Orderer" "U1F512"
  certsGenerate "$FABLO_NETWORK_ROOT/fabric-config" "crypto-config-orderer.yaml" "peerOrganizations/orderer.example.com" "$FABLO_NETWORK_ROOT/fabric-config/crypto-config/"

  printItalics "Generating crypto material for Org1" "U1F512"
  certsGenerate "$FABLO_NETWORK_ROOT/fabric-config" "crypto-config-org1.yaml" "peerOrganizations/org1.example.com" "$FABLO_NETWORK_ROOT/fabric-config/crypto-config/"

  printItalics "Generating crypto material for Org2" "U1F512"
  certsGenerate "$FABLO_NETWORK_ROOT/fabric-config" "crypto-config-org2.yaml" "peerOrganizations/org2.example.com" "$FABLO_NETWORK_ROOT/fabric-config/crypto-config/"

  printItalics "Generating crypto material for Org3" "U1F512"
  certsGenerate "$FABLO_NETWORK_ROOT/fabric-config" "crypto-config-org3.yaml" "peerOrganizations/org3.example.com" "$FABLO_NETWORK_ROOT/fabric-config/crypto-config/"

  printItalics "Generating genesis block for group group1" "U1F3E0"
  genesisBlockCreate "$FABLO_NETWORK_ROOT/fabric-config" "$FABLO_NETWORK_ROOT/fabric-config/config" "Group1Genesis"

  # Create directory for chaincode packages to avoid permission errors on linux
  mkdir -p "$FABLO_NETWORK_ROOT/fabric-config/chaincode-packages"
}

startNetwork() {
  printHeadline "Starting network" "U1F680"
  (cd "$FABLO_NETWORK_ROOT"/fabric-docker && docker-compose up -d)
  sleep 4
}

generateChannelsArtifacts() {
  printHeadline "Generating config for 'my-channel1'" "U1F913"
  createChannelTx "my-channel1" "$FABLO_NETWORK_ROOT/fabric-config" "MyChannel1" "$FABLO_NETWORK_ROOT/fabric-config/config"
  printHeadline "Generating config for 'my-channel2'" "U1F913"
  createChannelTx "my-channel2" "$FABLO_NETWORK_ROOT/fabric-config" "MyChannel2" "$FABLO_NETWORK_ROOT/fabric-config/config"
}

installChannels() {
  printHeadline "Creating 'my-channel1' on Org1/peer0" "U1F63B"
  docker exec -i cli.org1.example.com bash -c "source scripts/channel_fns.sh; createChannelAndJoin 'my-channel1' 'Org1MSP' 'peer0.org1.example.com:7041' 'crypto/users/Admin@org1.example.com/msp' 'orderer0.group1.orderer.example.com:7030';"

  printItalics "Joining 'my-channel1' on  Org2/peer0" "U1F638"
  docker exec -i cli.org2.example.com bash -c "source scripts/channel_fns.sh; fetchChannelAndJoin 'my-channel1' 'Org2MSP' 'peer0.org2.example.com:7061' 'crypto/users/Admin@org2.example.com/msp' 'orderer0.group1.orderer.example.com:7030';"
  printHeadline "Creating 'my-channel2' on Org2/peer1" "U1F63B"
  docker exec -i cli.org2.example.com bash -c "source scripts/channel_fns.sh; createChannelAndJoin 'my-channel2' 'Org2MSP' 'peer1.org2.example.com:7062' 'crypto/users/Admin@org2.example.com/msp' 'orderer0.group1.orderer.example.com:7030';"

  printItalics "Joining 'my-channel2' on  Org3/peer1" "U1F638"
  docker exec -i cli.org3.example.com bash -c "source scripts/channel_fns.sh; fetchChannelAndJoin 'my-channel2' 'Org3MSP' 'peer1.org3.example.com:7082' 'crypto/users/Admin@org3.example.com/msp' 'orderer0.group1.orderer.example.com:7030';"
}

installChaincodes() {
  if [ -n "$(ls "$CHAINCODES_BASE_DIR/./chaincodes/chaincode-go-NFT-Collectibles")" ]; then
    local version="0.0.1"
    printHeadline "Packaging chaincode 'chaincode_NFT_Collectibles'" "U1F60E"
    chaincodeBuild "chaincode_NFT_Collectibles" "golang" "$CHAINCODES_BASE_DIR/./chaincodes/chaincode-go-NFT-Collectibles" "16"
    chaincodePackage "cli.org1.example.com" "peer0.org1.example.com:7041" "chaincode_NFT_Collectibles" "$version" "golang" printHeadline "Installing 'chaincode_NFT_Collectibles' for Org1" "U1F60E"
    chaincodeInstall "cli.org1.example.com" "peer0.org1.example.com:7041" "chaincode_NFT_Collectibles" "$version" ""
    chaincodeApprove "cli.org1.example.com" "peer0.org1.example.com:7041" "my-channel1" "chaincode_NFT_Collectibles" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "" ""
    printHeadline "Installing 'chaincode_NFT_Collectibles' for Org2" "U1F60E"
    chaincodeInstall "cli.org2.example.com" "peer0.org2.example.com:7061" "chaincode_NFT_Collectibles" "$version" ""
    chaincodeApprove "cli.org2.example.com" "peer0.org2.example.com:7061" "my-channel1" "chaincode_NFT_Collectibles" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "" ""
    printItalics "Committing chaincode 'chaincode_NFT_Collectibles' on channel 'my-channel1' as 'Org1'" "U1F618"
    chaincodeCommit "cli.org1.example.com" "peer0.org1.example.com:7041" "my-channel1" "chaincode_NFT_Collectibles" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "" "peer0.org1.example.com:7041,peer0.org2.example.com:7061" "" ""
  else
    echo "Warning! Skipping chaincode 'chaincode_NFT_Collectibles' installation. Chaincode directory is empty."
    echo "Looked in dir: '$CHAINCODES_BASE_DIR/./chaincodes/chaincode-go-NFT-Collectibles'"
  fi
  if [ -n "$(ls "$CHAINCODES_BASE_DIR/./chaincodes/chaincode-go-DEX-AMM")" ]; then
    local version="0.0.1"
    printHeadline "Packaging chaincode 'chaincode_DEX_AMM'" "U1F60E"
    chaincodeBuild "chaincode_DEX_AMM" "golang" "$CHAINCODES_BASE_DIR/./chaincodes/chaincode-go-DEX-AMM" "16"
    chaincodePackage "cli.org2.example.com" "peer1.org2.example.com:7062" "chaincode_DEX_AMM" "$version" "golang" printHeadline "Installing 'chaincode_DEX_AMM' for Org2" "U1F60E"
    chaincodeInstall "cli.org2.example.com" "peer1.org2.example.com:7062" "chaincode_DEX_AMM" "$version" ""
    chaincodeApprove "cli.org2.example.com" "peer1.org2.example.com:7062" "my-channel2" "chaincode_DEX_AMM" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "" "collections/chaincode_DEX_AMM.json"
    printHeadline "Installing 'chaincode_DEX_AMM' for Org3" "U1F60E"
    chaincodeInstall "cli.org3.example.com" "peer1.org3.example.com:7082" "chaincode_DEX_AMM" "$version" ""
    chaincodeApprove "cli.org3.example.com" "peer1.org3.example.com:7082" "my-channel2" "chaincode_DEX_AMM" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "" "collections/chaincode_DEX_AMM.json"
    printItalics "Committing chaincode 'chaincode_DEX_AMM' on channel 'my-channel2' as 'Org2'" "U1F618"
    chaincodeCommit "cli.org2.example.com" "peer1.org2.example.com:7062" "my-channel2" "chaincode_DEX_AMM" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "" "peer1.org2.example.com:7062,peer1.org3.example.com:7082" "" "collections/chaincode_DEX_AMM.json"
  else
    echo "Warning! Skipping chaincode 'chaincode_DEX_AMM' installation. Chaincode directory is empty."
    echo "Looked in dir: '$CHAINCODES_BASE_DIR/./chaincodes/chaincode-go-DEX-AMM'"
  fi

}

installChaincode() {
  local chaincodeName="$1"
  if [ -z "$chaincodeName" ]; then
    echo "Error: chaincode name is not provided"
    exit 1
  fi

  local version="$2"
  if [ -z "$version" ]; then
    echo "Error: chaincode version is not provided"
    exit 1
  fi

  if [ "$chaincodeName" = "chaincode_NFT_Collectibles" ]; then
    if [ -n "$(ls "$CHAINCODES_BASE_DIR/./chaincodes/chaincode-go-NFT-Collectibles")" ]; then
      printHeadline "Packaging chaincode 'chaincode_NFT_Collectibles'" "U1F60E"
      chaincodeBuild "chaincode_NFT_Collectibles" "golang" "$CHAINCODES_BASE_DIR/./chaincodes/chaincode-go-NFT-Collectibles" "16"
      chaincodePackage "cli.org1.example.com" "peer0.org1.example.com:7041" "chaincode_NFT_Collectibles" "$version" "golang" printHeadline "Installing 'chaincode_NFT_Collectibles' for Org1" "U1F60E"
      chaincodeInstall "cli.org1.example.com" "peer0.org1.example.com:7041" "chaincode_NFT_Collectibles" "$version" ""
      chaincodeApprove "cli.org1.example.com" "peer0.org1.example.com:7041" "my-channel1" "chaincode_NFT_Collectibles" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "" ""
      printHeadline "Installing 'chaincode_NFT_Collectibles' for Org2" "U1F60E"
      chaincodeInstall "cli.org2.example.com" "peer0.org2.example.com:7061" "chaincode_NFT_Collectibles" "$version" ""
      chaincodeApprove "cli.org2.example.com" "peer0.org2.example.com:7061" "my-channel1" "chaincode_NFT_Collectibles" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "" ""
      printItalics "Committing chaincode 'chaincode_NFT_Collectibles' on channel 'my-channel1' as 'Org1'" "U1F618"
      chaincodeCommit "cli.org1.example.com" "peer0.org1.example.com:7041" "my-channel1" "chaincode_NFT_Collectibles" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "" "peer0.org1.example.com:7041,peer0.org2.example.com:7061" "" ""

    else
      echo "Warning! Skipping chaincode 'chaincode_NFT_Collectibles' install. Chaincode directory is empty."
      echo "Looked in dir: '$CHAINCODES_BASE_DIR/./chaincodes/chaincode-go-NFT-Collectibles'"
    fi
  fi
  if [ "$chaincodeName" = "chaincode_DEX_AMM" ]; then
    if [ -n "$(ls "$CHAINCODES_BASE_DIR/./chaincodes/chaincode-go-DEX-AMM")" ]; then
      printHeadline "Packaging chaincode 'chaincode_DEX_AMM'" "U1F60E"
      chaincodeBuild "chaincode_DEX_AMM" "golang" "$CHAINCODES_BASE_DIR/./chaincodes/chaincode-go-DEX-AMM" "16"
      chaincodePackage "cli.org2.example.com" "peer1.org2.example.com:7062" "chaincode_DEX_AMM" "$version" "golang" printHeadline "Installing 'chaincode_DEX_AMM' for Org2" "U1F60E"
      chaincodeInstall "cli.org2.example.com" "peer1.org2.example.com:7062" "chaincode_DEX_AMM" "$version" ""
      chaincodeApprove "cli.org2.example.com" "peer1.org2.example.com:7062" "my-channel2" "chaincode_DEX_AMM" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "" "collections/chaincode_DEX_AMM.json"
      printHeadline "Installing 'chaincode_DEX_AMM' for Org3" "U1F60E"
      chaincodeInstall "cli.org3.example.com" "peer1.org3.example.com:7082" "chaincode_DEX_AMM" "$version" ""
      chaincodeApprove "cli.org3.example.com" "peer1.org3.example.com:7082" "my-channel2" "chaincode_DEX_AMM" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "" "collections/chaincode_DEX_AMM.json"
      printItalics "Committing chaincode 'chaincode_DEX_AMM' on channel 'my-channel2' as 'Org2'" "U1F618"
      chaincodeCommit "cli.org2.example.com" "peer1.org2.example.com:7062" "my-channel2" "chaincode_DEX_AMM" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "" "peer1.org2.example.com:7062,peer1.org3.example.com:7082" "" "collections/chaincode_DEX_AMM.json"

    else
      echo "Warning! Skipping chaincode 'chaincode_DEX_AMM' install. Chaincode directory is empty."
      echo "Looked in dir: '$CHAINCODES_BASE_DIR/./chaincodes/chaincode-go-DEX-AMM'"
    fi
  fi
}

runDevModeChaincode() {
  local chaincodeName=$1
  if [ -z "$chaincodeName" ]; then
    echo "Error: chaincode name is not provided"
    exit 1
  fi

  if [ "$chaincodeName" = "chaincode_NFT_Collectibles" ]; then
    local version="0.0.1"
    printHeadline "Approving 'chaincode_NFT_Collectibles' for Org1 (dev mode)" "U1F60E"
    chaincodeApprove "cli.org1.example.com" "peer0.org1.example.com:7041" "my-channel1" "chaincode_NFT_Collectibles" "0.0.1" "orderer0.group1.orderer.example.com:7030" "" "false" "" ""
    printHeadline "Approving 'chaincode_NFT_Collectibles' for Org2 (dev mode)" "U1F60E"
    chaincodeApprove "cli.org2.example.com" "peer0.org2.example.com:7061" "my-channel1" "chaincode_NFT_Collectibles" "0.0.1" "orderer0.group1.orderer.example.com:7030" "" "false" "" ""
    printItalics "Committing chaincode 'chaincode_NFT_Collectibles' on channel 'my-channel1' as 'Org1' (dev mode)" "U1F618"
    chaincodeCommit "cli.org1.example.com" "peer0.org1.example.com:7041" "my-channel1" "chaincode_NFT_Collectibles" "0.0.1" "orderer0.group1.orderer.example.com:7030" "" "false" "" "peer0.org1.example.com:7041,peer0.org2.example.com:7061" "" ""

  fi
  if [ "$chaincodeName" = "chaincode_DEX_AMM" ]; then
    local version="0.0.1"
    printHeadline "Approving 'chaincode_DEX_AMM' for Org2 (dev mode)" "U1F60E"
    chaincodeApprove "cli.org2.example.com" "peer1.org2.example.com:7062" "my-channel2" "chaincode_DEX_AMM" "0.0.1" "orderer0.group1.orderer.example.com:7030" "" "false" "" "collections/chaincode_DEX_AMM.json"
    printHeadline "Approving 'chaincode_DEX_AMM' for Org3 (dev mode)" "U1F60E"
    chaincodeApprove "cli.org3.example.com" "peer1.org3.example.com:7082" "my-channel2" "chaincode_DEX_AMM" "0.0.1" "orderer0.group1.orderer.example.com:7030" "" "false" "" "collections/chaincode_DEX_AMM.json"
    printItalics "Committing chaincode 'chaincode_DEX_AMM' on channel 'my-channel2' as 'Org2' (dev mode)" "U1F618"
    chaincodeCommit "cli.org2.example.com" "peer1.org2.example.com:7062" "my-channel2" "chaincode_DEX_AMM" "0.0.1" "orderer0.group1.orderer.example.com:7030" "" "false" "" "peer1.org2.example.com:7062,peer1.org3.example.com:7082" "" "collections/chaincode_DEX_AMM.json"

  fi
}

upgradeChaincode() {
  local chaincodeName="$1"
  if [ -z "$chaincodeName" ]; then
    echo "Error: chaincode name is not provided"
    exit 1
  fi

  local version="$2"
  if [ -z "$version" ]; then
    echo "Error: chaincode version is not provided"
    exit 1
  fi

  if [ "$chaincodeName" = "chaincode_NFT_Collectibles" ]; then
    if [ -n "$(ls "$CHAINCODES_BASE_DIR/./chaincodes/chaincode-go-NFT-Collectibles")" ]; then
      printHeadline "Packaging chaincode 'chaincode_NFT_Collectibles'" "U1F60E"
      chaincodeBuild "chaincode_NFT_Collectibles" "golang" "$CHAINCODES_BASE_DIR/./chaincodes/chaincode-go-NFT-Collectibles" "16"
      chaincodePackage "cli.org1.example.com" "peer0.org1.example.com:7041" "chaincode_NFT_Collectibles" "$version" "golang" printHeadline "Installing 'chaincode_NFT_Collectibles' for Org1" "U1F60E"
      chaincodeInstall "cli.org1.example.com" "peer0.org1.example.com:7041" "chaincode_NFT_Collectibles" "$version" ""
      chaincodeApprove "cli.org1.example.com" "peer0.org1.example.com:7041" "my-channel1" "chaincode_NFT_Collectibles" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "" ""
      printHeadline "Installing 'chaincode_NFT_Collectibles' for Org2" "U1F60E"
      chaincodeInstall "cli.org2.example.com" "peer0.org2.example.com:7061" "chaincode_NFT_Collectibles" "$version" ""
      chaincodeApprove "cli.org2.example.com" "peer0.org2.example.com:7061" "my-channel1" "chaincode_NFT_Collectibles" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "" ""
      printItalics "Committing chaincode 'chaincode_NFT_Collectibles' on channel 'my-channel1' as 'Org1'" "U1F618"
      chaincodeCommit "cli.org1.example.com" "peer0.org1.example.com:7041" "my-channel1" "chaincode_NFT_Collectibles" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "" "peer0.org1.example.com:7041,peer0.org2.example.com:7061" "" ""

    else
      echo "Warning! Skipping chaincode 'chaincode_NFT_Collectibles' upgrade. Chaincode directory is empty."
      echo "Looked in dir: '$CHAINCODES_BASE_DIR/./chaincodes/chaincode-go-NFT-Collectibles'"
    fi
  fi
  if [ "$chaincodeName" = "chaincode_DEX_AMM" ]; then
    if [ -n "$(ls "$CHAINCODES_BASE_DIR/./chaincodes/chaincode-go-DEX-AMM")" ]; then
      printHeadline "Packaging chaincode 'chaincode_DEX_AMM'" "U1F60E"
      chaincodeBuild "chaincode_DEX_AMM" "golang" "$CHAINCODES_BASE_DIR/./chaincodes/chaincode-go-DEX-AMM" "16"
      chaincodePackage "cli.org2.example.com" "peer1.org2.example.com:7062" "chaincode_DEX_AMM" "$version" "golang" printHeadline "Installing 'chaincode_DEX_AMM' for Org2" "U1F60E"
      chaincodeInstall "cli.org2.example.com" "peer1.org2.example.com:7062" "chaincode_DEX_AMM" "$version" ""
      chaincodeApprove "cli.org2.example.com" "peer1.org2.example.com:7062" "my-channel2" "chaincode_DEX_AMM" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "" "collections/chaincode_DEX_AMM.json"
      printHeadline "Installing 'chaincode_DEX_AMM' for Org3" "U1F60E"
      chaincodeInstall "cli.org3.example.com" "peer1.org3.example.com:7082" "chaincode_DEX_AMM" "$version" ""
      chaincodeApprove "cli.org3.example.com" "peer1.org3.example.com:7082" "my-channel2" "chaincode_DEX_AMM" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "" "collections/chaincode_DEX_AMM.json"
      printItalics "Committing chaincode 'chaincode_DEX_AMM' on channel 'my-channel2' as 'Org2'" "U1F618"
      chaincodeCommit "cli.org2.example.com" "peer1.org2.example.com:7062" "my-channel2" "chaincode_DEX_AMM" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "" "peer1.org2.example.com:7062,peer1.org3.example.com:7082" "" "collections/chaincode_DEX_AMM.json"

    else
      echo "Warning! Skipping chaincode 'chaincode_DEX_AMM' upgrade. Chaincode directory is empty."
      echo "Looked in dir: '$CHAINCODES_BASE_DIR/./chaincodes/chaincode-go-DEX-AMM'"
    fi
  fi
}

notifyOrgsAboutChannels() {
  printHeadline "Creating new channel config blocks" "U1F537"
  createNewChannelUpdateTx "my-channel1" "Org1MSP" "MyChannel1" "$FABLO_NETWORK_ROOT/fabric-config" "$FABLO_NETWORK_ROOT/fabric-config/config"
  createNewChannelUpdateTx "my-channel1" "Org2MSP" "MyChannel1" "$FABLO_NETWORK_ROOT/fabric-config" "$FABLO_NETWORK_ROOT/fabric-config/config"
  createNewChannelUpdateTx "my-channel2" "Org2MSP" "MyChannel2" "$FABLO_NETWORK_ROOT/fabric-config" "$FABLO_NETWORK_ROOT/fabric-config/config"
  createNewChannelUpdateTx "my-channel2" "Org3MSP" "MyChannel2" "$FABLO_NETWORK_ROOT/fabric-config" "$FABLO_NETWORK_ROOT/fabric-config/config"

  printHeadline "Notyfing orgs about channels" "U1F4E2"
  notifyOrgAboutNewChannel "my-channel1" "Org1MSP" "cli.org1.example.com" "peer0.org1.example.com" "orderer0.group1.orderer.example.com:7030"
  notifyOrgAboutNewChannel "my-channel1" "Org2MSP" "cli.org2.example.com" "peer0.org2.example.com" "orderer0.group1.orderer.example.com:7030"
  notifyOrgAboutNewChannel "my-channel2" "Org2MSP" "cli.org2.example.com" "peer0.org2.example.com" "orderer0.group1.orderer.example.com:7030"
  notifyOrgAboutNewChannel "my-channel2" "Org3MSP" "cli.org3.example.com" "peer0.org3.example.com" "orderer0.group1.orderer.example.com:7030"

  printHeadline "Deleting new channel config blocks" "U1F52A"
  deleteNewChannelUpdateTx "my-channel1" "Org1MSP" "cli.org1.example.com"
  deleteNewChannelUpdateTx "my-channel1" "Org2MSP" "cli.org2.example.com"
  deleteNewChannelUpdateTx "my-channel2" "Org2MSP" "cli.org2.example.com"
  deleteNewChannelUpdateTx "my-channel2" "Org3MSP" "cli.org3.example.com"
}

printStartSuccessInfo() {
  printHeadline "Done! Enjoy your fresh network" "U1F984"
}

stopNetwork() {
  printHeadline "Stopping network" "U1F68F"
  (cd "$FABLO_NETWORK_ROOT"/fabric-docker && docker-compose stop)
  sleep 4
}

networkDown() {
  printHeadline "Destroying network" "U1F916"
  (cd "$FABLO_NETWORK_ROOT"/fabric-docker && docker-compose down)

  printf "Removing chaincode containers & images... \U1F5D1 \n"
  for container in $(docker ps -a | grep "dev-peer0.org1.example.com-chaincode_NFT_Collectibles" | awk '{print $1}'); do
    echo "Removing container $container..."
    docker rm -f "$container" || echo "docker rm of $container failed. Check if all fabric dockers properly was deleted"
  done
  for image in $(docker images "dev-peer0.org1.example.com-chaincode_NFT_Collectibles*" -q); do
    echo "Removing image $image..."
    docker rmi "$image" || echo "docker rmi of $image failed. Check if all fabric dockers properly was deleted"
  done
  for container in $(docker ps -a | grep "dev-peer0.org2.example.com-chaincode_NFT_Collectibles" | awk '{print $1}'); do
    echo "Removing container $container..."
    docker rm -f "$container" || echo "docker rm of $container failed. Check if all fabric dockers properly was deleted"
  done
  for image in $(docker images "dev-peer0.org2.example.com-chaincode_NFT_Collectibles*" -q); do
    echo "Removing image $image..."
    docker rmi "$image" || echo "docker rmi of $image failed. Check if all fabric dockers properly was deleted"
  done
  for container in $(docker ps -a | grep "dev-peer1.org2.example.com-chaincode_DEX_AMM" | awk '{print $1}'); do
    echo "Removing container $container..."
    docker rm -f "$container" || echo "docker rm of $container failed. Check if all fabric dockers properly was deleted"
  done
  for image in $(docker images "dev-peer1.org2.example.com-chaincode_DEX_AMM*" -q); do
    echo "Removing image $image..."
    docker rmi "$image" || echo "docker rmi of $image failed. Check if all fabric dockers properly was deleted"
  done
  for container in $(docker ps -a | grep "dev-peer1.org3.example.com-chaincode_DEX_AMM" | awk '{print $1}'); do
    echo "Removing container $container..."
    docker rm -f "$container" || echo "docker rm of $container failed. Check if all fabric dockers properly was deleted"
  done
  for image in $(docker images "dev-peer1.org3.example.com-chaincode_DEX_AMM*" -q); do
    echo "Removing image $image..."
    docker rmi "$image" || echo "docker rmi of $image failed. Check if all fabric dockers properly was deleted"
  done

  printf "Removing generated configs... \U1F5D1 \n"
  rm -rf "$FABLO_NETWORK_ROOT/fabric-config/config"
  rm -rf "$FABLO_NETWORK_ROOT/fabric-config/crypto-config"
  rm -rf "$FABLO_NETWORK_ROOT/fabric-config/chaincode-packages"

  printHeadline "Done! Network was purged" "U1F5D1"
}
