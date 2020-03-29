#!/bin/bash
# Env variables
export DOCKER_BUILD_NAME="gzy123/gateway"
export MYSQL_BUILD_NAME="gzy123/db"
export ADDR=":443"
export TLSCERT="/etc/letsencrypt/live/api.ziyuguo.me/fullchain.pem"
export TLSKEY="/etc/letsencrypt/live/api.ziyuguo.me/privkey.pem"
export MESSAGESADDR="http://micro-messaging:4000"
export SUMMARYADDR="http://micro-summary:8080"

bash build.sh

docker push $DOCKER_BUILD_NAME;
docker push $MYSQL_BUILD_NAME;

# ssh -i ~/.ssh/EricKey.pem ec2-user@ec2-54-159-132-26.compute-1.amazonaws.com < update.sh
ssh -i ~/.ssh/EricKey.pem ec2-user@api.ziyuguo.me < update.sh
