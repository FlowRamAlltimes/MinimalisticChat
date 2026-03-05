FROM alpine:3.20.3
# do not use with tag latest
RUN apk --no-cache add ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Europe/Moscow /etc/localtime
WORKDIR /root/
COPY chat .
COPY server.crt .
COPY server.key .
#copies cert files
RUN touch banlist.txt mutelist.txt server.log
# touches files
EXPOSE 1358 1359
# random ports
RUN chmod +x chat
CMD ["./chat", "-addr=0.0.0.0", "-p=1358", "-pRst=1359"]
# I have already checked, Dockerfile works fine
