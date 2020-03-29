export SUMMARY_BUILD_NAME="gzy123/summary"

bash build.sh
ssh -i ~/.ssh/EricKey.pem ec2-user@api.ziyuguo.me < update.sh
