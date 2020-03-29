export DOCKER_BUILD_NAME="gzy123/summary"
export ADDR=":443"
export TLSCERT="/etc/letsencrypt/live/ziyuguo.me/fullchain.pem"
export TLSKEY="/etc/letsencrypt/live/ziyuguo.me/privkey.pem"

docker rm -f info441-clientserver

docker pull $DOCKER_BUILD_NAME

docker run \
    -d \
    --name info441-clientserver \
    -p 80:80 \
    -p 443:443 \
    -v /etc/letsencrypt:/etc/letsencrypt:ro \
    -e ADDR=$ADDR \
    -e TLSCERT=$TLSCERT \
    -e TLSKEY=$TLSKEY \
    $DOCKER_BUILD_NAME

exit
