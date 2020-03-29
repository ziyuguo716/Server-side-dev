export SUMMARY_BUILD_NAME="gzy123/messaging"

docker rm -f micro-messaging

docker pull $SUMMARY_BUILD_NAME

docker run -d --name micro-messaging --network 441network -p 4000:4000 $SUMMARY_BUILD_NAME


exit