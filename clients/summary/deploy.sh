bash build.sh

docker login
docker push gzy123/summary

ssh -i ~/.ssh/EricKey.pem ec2-user@ziyuguo.me < update.sh