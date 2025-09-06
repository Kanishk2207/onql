# syntax=docker/dockerfile:1
FROM ubuntu:22.04

# avoid interactive prompts
ENV DEBIAN_FRONTEND=noninteractive
# enable Go modules
ENV GO111MODULE=on

# install runtimes & JDK, then clean up
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      python3 python3-pip \
      curl ca-certificates \
      nodejs npm \
      golang-go \
      default-jdk \
 && rm -rf /var/lib/apt/lists/*

# force-create a python -> python3 link, and node -> nodejs (ignore errors if it already exists)
RUN ln -sf /usr/bin/python3 /usr/bin/python \
 && ln -sf /usr/bin/nodejs /usr/bin/node || true

# set working dir
WORKDIR /workspace

# default entrypoint
ENTRYPOINT ["bash"]
