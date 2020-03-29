export SUMMARY_BUILD_NAME="gzy123/summary"

docker rm -f micro-summary

docker pull $SUMMARY_BUILD_NAME

docker run -d --name micro-summary --network 441network -p 8080:8080 $SUMMARY_BUILD_NAME

exit