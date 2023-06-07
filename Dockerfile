ARG BASE_IMAGE=docker.io/library/ubuntu:22.04
####################################################################################################
# Builder image
# Initial stage which pulls prepares build dependencies and CLI tooling we need for our final image
# Also used as the image in CI jobs so needs all dependencies
####################################################################################################
FROM docker.io/library/golang:1.18 AS builder

RUN echo 'deb http://deb.debian.org/debian buster-backports main' >> /etc/apt/sources.list

RUN apt-get update && apt-get install --no-install-recommends -y \
    openssh-server \
    nginx \
    unzip \
    fcgiwrap \
    git \
    git-lfs \
    make \
    wget \
    gcc \
    sudo \
    zip && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

WORKDIR /tmp

####################################################################################################
# Base - used as the base for both the release and dev images
####################################################################################################
FROM $BASE_IMAGE AS selfservice-base

USER root

ENV SELFSERVICE_USER_ID=999
ENV DEBIAN_FRONTEND=noninteractive

RUN groupadd -g $SELFSERVICE_USER_ID selfservice && \
    useradd -r -u $SELFSERVICE_USER_ID -g selfservice selfservice && \
    mkdir -p /home/selfservice && \
    chown selfservice:0 /home/selfservice && \
    chmod g=u /home/selfservice && \
    apt-get update && \
    apt-get dist-upgrade -y && \
    apt-get install -y \
    git git-lfs tini gpg tzdata && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY entrypoint.sh /usr/local/bin/entrypoint.sh

ENV USER=selfservice

USER $SELFSERVICE_USER_ID
WORKDIR /home/selfservice

####################################################################################################
# UI stage
####################################################################################################
FROM --platform=$BUILDPLATFORM docker.io/library/node:12.18.4 AS selfservice-ui

WORKDIR /src
COPY ["ui/package.json", "ui/package-lock.json", "./"]

RUN npm install && npm cache clean --force

COPY ["ui/", "."]

ARG TARGETARCH
RUN HOST_ARCH=$TARGETARCH NODE_ENV='production' NODE_OPTIONS=--max_old_space_size=8192 npm run build

####################################################################################################
# Build stage which performs the actual build of binaries
####################################################################################################
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.18 AS selfservice-build

WORKDIR /go/src/github.com/infor-design/selfservice

COPY go.* ./
RUN go mod download

# Perform the build
COPY . .
COPY --from=selfservice-ui /src/build /go/src/github.com/infor-design/selfservice/ui/build
ARG TARGETOS
ARG TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH make selfservice-all

####################################################################################################
# Final image
####################################################################################################
FROM selfservice-base
COPY --from=selfservice-build /go/src/github.com/infor-design/selfservice/dist/selfservice* /usr/local/bin/

USER root
RUN ln -s /usr/local/bin/selfservice /usr/local/bin/selfservice-server && \
    ln -s /usr/local/bin/selfservice /usr/local/bin/selfservice-reposerver && \
    ln -s /usr/local/bin/selfservice /usr/local/bin/selfservice-wsserver

USER $SELFSERVICE_USER_ID
