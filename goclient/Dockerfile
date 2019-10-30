FROM ubuntu:bionic

ARG DEBIAN_FRONTEND=noninteractive

RUN apt-get update \
 && apt-get install gnupg -y

RUN echo "deb [arch=amd64] http://repo.sawtooth.me/ubuntu/ci bionic universe" >> /etc/apt/sources.list \
 && echo "deb http://archive.ubuntu.com/ubuntu bionic-backports universe" >> /etc/apt/sources.list \
 && echo "deb [arch=amd64] http://repo.sawtooth.me/ubuntu/bumper/stable xenial universe" >> /etc/apt/sources.list \
 && echo 'deb http://ppa.launchpad.net/gophers/archive/ubuntu bionic main' >> /etc/apt/sources.list \
 && (apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys 8AA7AF1F1091A5FD \
 || apt-key adv --keyserver hkp://p80.pool.sks-keyservers.net:80 --recv-keys 8AA7AF1F1091A5FD) \
 && (apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys 308C15A29AD198E9 \
 || apt-key adv --keyserver hkp://p80.pool.sks-keyservers.net:80 --recv-keys 308C15A29AD198E9) \
 && apt-get update \
 && apt-get install -y -q \
    build-essential \
    golang-1.11-go \
    git \
    libssl-dev \
    libzmq3-dev \
    openssl \
    python3-grpcio-tools \
    python3-sawtooth-cli \
    python3-sawtooth-sdk \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /app
ENV GOPATH=/go:/go/src/github.com/hyperledger/sawtooth-sdk-go:/app
ENV PATH=$PATH::/go/bin:/usr/lib/go-1.11/bin:/app

RUN go get -u \
    github.com/golang/protobuf/proto \
    github.com/golang/protobuf/protoc-gen-go \
    github.com/pebbe/zmq4 \
    github.com/satori/go.uuid \
    github.com/btcsuite/btcd/btcec \
    github.com/jessevdk/go-flags \
    github.com/golang/mock/gomock \
    github.com/golang/mock/mockgen \
    golang.org/x/crypto/ssh \
    gopkg.in/yaml.v2 \
    github.com/hyperledger/sawtooth-sdk-go

WORKDIR /go/src/github.com/hyperledger/sawtooth-sdk-go
RUN go generate 

EXPOSE 3000

WORKDIR /app
COPY ./goclient/*.go ./
RUN go build -o cookiejar
