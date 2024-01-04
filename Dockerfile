# Copyright (C) 2019-2022 Zilliz. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance
# with the License. You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software distributed under the License
# is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
# or implied. See the License for the specific language governing permissions and limitations under the License.

ARG CUDA_VERSION=12.1.0
ARG GO_VERSION=1.20
ARG CONAN_VERSION=1.61.0
ARG CMAKE_VERSION=3.27.4
ARG RUST_VERSION=1.73
ARG TARGET_ARCH=x86_64

FROM golang:${GO_VERSION} AS golang_image
FROM nvidia/cuda:${CUDA_VERSION}-devel-ubuntu20.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update \
 && apt-get install -qy \
    wget \
    curl \
    ca-certificates \
    gnupg2\
    g++ \
    gcc \
    gfortran \
    git \
    make \
    ccache \
    libssl-dev \
    zlib1g-dev \
    zip \
    unzip \
    clang-format-10 \
    clang-tidy-10 \
    lcov \
    libtool \
    m4 \
    autoconf \
    automake \
    python3 \
    python3-pip \
    pkg-config \
    uuid-dev \
    libaio-dev \
    libgoogle-perftools-dev

ENV GOROOT=/usr/local/go
ENV GOPATH=/go
COPY --from=golang_image $GOROOT $GOROOT
COPY --from=golang_image $GOPATH $GOPATH

ENV GO111MODULE=on
ENV GOOS=linux
ENV GOARCH=${TARGET_ARCH}
ENV PATH="$GOPATH/bin:$GOROOT/bin:${PATH}"

RUN mkdir /tmp/downloads
WORKDIR /tmp/downloads
ARG CMAKE_VERSION
RUN curl -OL \
    https://github.com/Kitware/CMake/releases/download/v${CMAKE_VERSION}/cmake-${CMAKE_VERSION}.tar.gz \
 && tar -xzf cmake-${CMAKE_VERSION}.tar.gz \
 && cd cmake-${CMAKE_VERSION} \
 && ./bootstrap -- -DCMAKE_BUILD_TYPE:STRING=Release \
 && make -j$(nproc) \
 && make install \
 && rm -rf /tmp/downloads \
 && mkdir /workspace \
 && cd /workspace

WORKDIR /workspace

ARG RUST_VERSION
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs \
 |  sh -s -- --default-toolchain=${RUST_VERSION} -y

ENV PATH="/root/.cargo/bin:${PATH}"

ARG CONAN_VERSION
RUN pip3 install conan==$CONAN_VERSION

RUN mkdir /workspace/conan
ENV CONAN_USER_HOME=/workspace/conan
RUN conan remote add default-conan-local \
    https://milvus01.jfrog.io/artifactory/api/conan/default-conan-local

COPY ./ /workspace/src
WORKDIR /workspace/src

RUN make clean && make gpu-install

RUN tar -czf /cmake_build_cache.tar.gz /workspace/src/cmake_build
COPY ./entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
