
FROM alpine
RUN apk add --no-cache bash ca-certificates
COPY summary /summary
EXPOSE 8080
ENTRYPOINT ["/summary"]
