#!/bin/bash
set -e

log() {
  echo ""
  echo "##################################"
  echo "-------> $1"
  echo "##################################"
}

setup() {
  export DOCKER_REGISTRY_ADDRESS="127.0.0.1:5000"
  export DOCKER_REGISTRY_USER="testuser"
  export DOCKER_REGISTRY_PASS="testpwd"
  export DOCKER_PRIVATE_IMAGE="127.0.0.1:5000/my-private-service:v1"
  sh scripts/testing/setup_private_registry.sh
}

run() {
  # Run the acc test suite
  #make testacc
  # for a single test
  #TF_LOG=INFO TF_ACC=1 go test -v -timeout 240s github.com/terraform-providers/terraform-provider-docker/docker -run ^TestAccDockerService_updateImage$
  #TF_LOG=INFO TF_ACC=1 go test -v -timeout 240s github.com/terraform-providers/terraform-provider-docker/docker -run ^TestAccDockerService_updateConfigForMultipleServices$
  #TF_LOG=INFO TF_ACC=1 go test -v -timeout 240s github.com/terraform-providers/terraform-provider-docker/docker -run ^TestAccDockerService_updateConfigAndDecreaseReplicas$
  TF_LOG=INFO TF_ACC=1 go test -v -timeout 240s github.com/terraform-providers/terraform-provider-docker/docker -run ^TestAccDockerService_updateConfigReplicasImageAndHealthIncreaseAndDecreaseReplicas$
  
  # keep the return for the scripts to fail and clean properly
  return $?
}

cleanup() {
  unset DOCKER_REGISTRY_ADDRESS DOCKER_REGISTRY_USER DOCKER_REGISTRY_PASS DOCKER_PRIVATE_IMAGE
  echo "### unsetted env ###"
  rm -f scripts/testing/auth/htpasswd
  rm -f scripts/testing/certs/registry_auth.*
  echo "### removed auth and certs ###"
  docker stop private_registry
  echo "### stopped private registry ###"
  docker rmi -f $(docker images -aq 127.0.0.1:5000/my-private-service)
  echo "### removed my-private-service images ###"
  # consider running this manually to clean up the
  # updateabe configs and secrets
  #docker config rm $(docker config ls -q)
  #docker secret rm $(docker secret ls -q)
}

## main
log "setup" && setup 
log "run" && run && echo $?
if [ $? -ne 0 ]; then
  log "cleanup" && cleanup 
  exit 1
fi
log "cleanup" && cleanup