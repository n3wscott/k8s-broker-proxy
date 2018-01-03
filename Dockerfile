FROM alpine:3.4

RUN apk -U add ca-certificates

EXPOSE 8080

ADD proxy /bin/proxy
ADD config.yml /etc/proxy/config.yml

CMD ["proxy", "-config", "/etc/proxy/config.yml"]
