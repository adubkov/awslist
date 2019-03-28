FROM gliderlabs/alpine:latest
RUN apk-install ca-certificates
ADD ./bin/awslist.linux.amd64 /opt/awslist
CMD /opt/awslist -service
EXPOSE 8080
