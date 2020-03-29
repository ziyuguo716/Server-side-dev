export DOCKER_BUILD_NAME="gzy123/gateway"
export MYSQL_BUILD_NAME="gzy123/db"
export ADDR=":443"
export TLSCERT="/etc/letsencrypt/live/api.ziyuguo.me/fullchain.pem"
export TLSKEY="/etc/letsencrypt/live/api.ziyuguo.me/privkey.pem"
export MYSQL_ROOT_PASSWORD="mypassword"
export SESSIONKEY="thisismykey"
export REDISADDR="redis:6379"
export DSN="root:mypassword@tcp(mysql:3306)/mydb"
export MESSAGESADDR="http://micro-messaging:4000"
export SUMMARYADDR="http://micro-summary:8080"

docker network create 441network

docker rm -f info441-api
docker rm -f mysql
docker rm -f redis
# docker rm -f rabbit

docker pull $DOCKER_BUILD_NAME
docker pull $MYSQL_BUILD_NAME

docker run -d \
    -p 6379:6379 \
    --name redis  \
    --network 441network \
    redis

docker run -d \
    -p 3306:3306 \
    --name mysql  \
    --network 441network \
    -e MYSQL_ROOT_PASSWORD=$MYSQL_ROOT_PASSWORD \
    gzy123/db

# docker run -d \
#     --name rabbit \
#     -p 5672:5672 -p 15672:15672 \
#     --network 441network \
#     rabbitmq:3-management 


docker run \
    -d \
    --name info441-api \
    --network 441network \
    -p 80:80 \
    -p 443:443 \
    -v /etc/letsencrypt:/etc/letsencrypt:ro \
    -e ADDR=$ADDR \
    -e TLSCERT=$TLSCERT \
    -e TLSKEY=$TLSKEY \
    -e SESSIONKEY=$SESSIONKEY \
    -e REDISADDR=$REDISADDR \
    -e MESSAGESADDR=$MESSAGESADDR \
    -e SUMMARYADDR=$SUMMARYADDR \
    -e MYSQL_ROOT_PASSWORD=$MYSQL_ROOT_PASSWORD \
    -e DSN=$DSN \
    $DOCKER_BUILD_NAME

docker system prune --all -f
docker volume prune -f
exit
