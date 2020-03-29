bash build.sh
docker push gzy123/messages

ssh -i ~/.ssh/EricKey.pem ec2-user@ziyuguo.me < update.sh