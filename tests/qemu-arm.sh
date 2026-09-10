#!/usr/bin/env bash
# Run the Go integration suite against the NDK-built shared libraries.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TEST_ROOT="${TEST_ROOT:-/tmp/opuspurego/qemu}"
export LIBOPUS_DIST="${LIBOPUS_DIST:-/tmp/opuspurego/dist/libopus}"
ARM64_CC="${ARM64_CC:-/home/yj/toolchains/arm-linux-musleabi-armv8a/bin/aarch64-linux-musl-gcc}"
ARM_CC="${ARM_CC:-/home/yj/toolchains/arm-linux-musleabi-armv7a/bin/arm-linux-musleabi-gcc}"
cd "$ROOT"
for arch in arm64 arm; do
    if [[ "$arch" == arm64 ]]; then
        cc="$ARM64_CC"; loader=ld-musl-aarch64.so.1; runner=qemu-aarch64
    else
        cc="$ARM_CC"; loader=ld-musl-arm.so.1; runner=qemu-arm
    fi
    sysroot="$TEST_ROOT/$arch"
    mkdir -p "$sysroot/lib"
    # Toolchain loaders may point to absolute /lib/libc.so. Relative links
    # in the temporary sysroot prevent QEMU from selecting the host libc.
    cp "$("$cc" -print-file-name=libc.so)" "$sysroot/lib/libc.so"
    ln -sfn libc.so "$sysroot/lib/$loader"
    CGO_ENABLED=1 GOOS=linux GOARCH="$arch" GOARM=7 CC="$cc" \
        go test -c -o "$TEST_ROOT/test-$arch"
    "$runner" -L "$sysroot" "$TEST_ROOT/test-$arch" -test.v
done
