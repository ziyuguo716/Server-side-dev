# add the necessary instructions
# to create a Docker container image
# for your summary client using the
# NGINX web server image as the base

# Dockerfile for client
FROM nginx
# /usr/share/nginx/html is where
# nginx serves its static files
COPY . /usr/share/nginx/html/
ADD default.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
EXPOSE 443