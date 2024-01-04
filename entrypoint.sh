#!/bin/bash
set -e

owner=$(stat -c '%u:%g' /workspace/src)

fix_owner() {
    chown -R $owner /workspace/src
}
trap fix_owner EXIT

pushd /workspace/src
if [ ! -d "cmake_build" ]; then
    tar -xzf /cmake_build_cache.tar.gz
fi

make gpu-install
