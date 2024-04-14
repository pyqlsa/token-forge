# (ref): https://github.com/jarlave/docker-tor-relay-nyx/blob/master/Dockerfile
FROM docker.io/library/alpine:edge

# add group/user tor with ID
RUN addgroup -g 1000 -S tor && \
    adduser -u 1000 -S tor -G tor

RUN echo '@edge https://dl-cdn.alpinelinux.org/alpine/edge/community' >> /etc/apk/repositories && \
    echo '@edge https://dl-cdn.alpinelinux.org/alpine/edge/testing'   >> /etc/apk/repositories && \
    apk update && \
    apk upgrade && \
    apk add --update-cache tor@edge obfs4proxy@edge && \
    rm -rf /var/cache/apk/* && \
    rm -rf /etc/tor/torrc.d && \
    mkdir /etc/tor/torrc.d

COPY ./proxy/torrc.local /etc/tor/torrc
COPY ./proxy/priv.d/ /etc/tor/torrc.d/

# make sure files are owned by tor user
RUN chown -R tor:root /etc/tor
RUN chown -R tor:root /var/lib/tor

USER tor
EXPOSE 9050/tcp 9080/tcp 8853/udp

CMD [ "/usr/bin/tor", "-f", "/etc/tor/torrc" ]
