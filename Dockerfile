FROM ghcr.io/linuxserver/baseimage-ubuntu:bionic AS base
LABEL maintainer="zxc25077667@pm.me"

# set version label
ARG BUILD_DATE
ARG VERSION
ARG WIREGUARD_RELEASE
LABEL build_version="Linuxserver.io version:- ${VERSION} Build-date:- ${BUILD_DATE}"
LABEL maintainer="aptalca"

ENV DEBIAN_FRONTEND="noninteractive"
ENV KEEPALIVE="25"

RUN \
	echo "**** install dependencies ****" && \
	apt-get update && \
	apt-get install -y --no-install-recommends \
	bc \
	build-essential \
	curl \
	dkms \
	git \
	gnupg \ 
	ifupdown \
	iproute2 \
	iptables \
	iputils-ping \
	jq \
	libc6 \
	libelf-dev \
	net-tools \
	openresolv \
	perl \
	pkg-config \
	qrencode && \
	echo "**** install wireguard-tools ****" && \
	if [ -z ${WIREGUARD_RELEASE+x} ]; then \
	WIREGUARD_RELEASE=$(curl -sX GET "https://api.github.com/repos/WireGuard/wireguard-tools/tags" \
	| jq -r .[0].name); \
	fi && \
	cd /app && \
	git clone https://git.zx2c4.com/wireguard-linux-compat && \
	git clone https://git.zx2c4.com/wireguard-tools && \
	cd wireguard-tools && \
	git checkout "${WIREGUARD_RELEASE}" && \
	make -C src -j$(nproc) && \
	make -C src install && \
	echo "**** install CoreDNS ****" && \
	COREDNS_VERSION=$(curl -sX GET "https://api.github.com/repos/coredns/coredns/releases/latest" \
	| awk '/tag_name/{print $4;exit}' FS='[""]' | awk '{print substr($1,2); }') && \
	curl -o \
	/tmp/coredns.tar.gz -L \
	"https://github.com/coredns/coredns/releases/download/v${COREDNS_VERSION}/coredns_${COREDNS_VERSION}_linux_amd64.tgz" && \
	tar xf \
	/tmp/coredns.tar.gz -C \
	/app && \
	echo "**** clean up ****" && \
	rm -rf \
	/tmp/* \
	/var/lib/apt/lists/* \
	/var/tmp/*

# add local files
COPY /root /

# build the agent
FROM golang:alpine AS builder
RUN apk --no-cache add libcap ca-certificates
COPY ./vpn-agent /src/vpn-agent
WORKDIR /src/vpn-agent
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o agent monitor.go


FROM base
COPY --from=builder /src/vpn-agent/agent /app
ENV PUID=1000
ENV PGID=1000
ENV TZ=Asia/Taipei
ENV SERVERPORT=51820
ENV PEERDNS=auto
ENV INTERNAL_SUBNET=10.13.13.0
ENV ALLOWEDIPS=0.0.0.0/0

# for agent
EXPOSE 8080
# for vpn service
EXPOSE 51820/udp

VOLUME [ "/lib/modules" ]