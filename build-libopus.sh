#!/usr/bin/env bash
# Build the pinned opus submodule. All generated files live outside the repo.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SRC_DIR="${OPUS_SRC_DIR:-${SCRIPT_DIR}/opus}"
BUILD_ROOT="${BUILD_ROOT:-/tmp/opuspurego/build}"
DIST_ROOT="${DIST_ROOT:-/tmp/opuspurego/dist/libopus}"
NDK_ROOT="${NDK_ROOT:-/home/yj/Android/Sdk/ndk/28.2.13676358}"
NDK_TOOLCHAIN="${NDK_ROOT}/toolchains/llvm/prebuilt/linux-x86_64/bin"
NDK_API="${NDK_API:-21}"
JOBS="${JOBS:-$(getconf _NPROCESSORS_ONLN)}"
DISABLE_SIMD=OFF
TARGETS=(linux-x64 linux-x86 linux-arm64 linux-armv7a mingw-x64 mingw-x86 wasm)
COMMON=(-DCMAKE_BUILD_TYPE=Release -DBUILD_SHARED_LIBS=OFF
    -DOPUS_BUILD_SHARED_LIBRARY=OFF -DOPUS_BUILD_TESTING=OFF
    -DOPUS_BUILD_PROGRAMS=OFF -DOPUS_FIXED_POINT=ON
    -DOPUS_DISABLE_INTRINSICS=OFF -DOPUS_STACK_PROTECTOR=OFF
    -DOPUS_FORTIFY_SOURCE=OFF -DOPUS_HARDENING=OFF
    -DOPUS_DRED=OFF -DOPUS_OSCE=OFF -DCMAKE_POSITION_INDEPENDENT_CODE=ON)

check_elf() {
    local so="$1" symbol target="$2"
    if readelf -d "$so" | grep -q '(NEEDED)'; then
        echo "ERROR: unexpected platform dependency: $so" >&2
        return 1
    fi
    if readelf --dyn-syms --wide "$so" | grep -q '@'; then
        echo "ERROR: versioned imports in $so" >&2
        return 1
    fi
    # Only the common libc subset is allowed. Compiler helpers must be linked.
    while read -r symbol; do
        case "$symbol" in
            malloc|calloc|realloc|free|memcpy|memmove|memset|memcmp|strlen|abort) ;;
            __stack_chk_fail|__stack_chk_guard)
                # NDK's ARM libm uses the ARM libc stack protector ABI.
                [[ "$target" == linux-armv7a ]] || return 1 ;;
            *) echo "ERROR: nonportable import $symbol in $so" >&2; return 1 ;;
        esac
    done < <("${NDK_TOOLCHAIN}/llvm-nm" -D --undefined-only "$so" | awk '{print $NF}')
    readelf -lW "$so" | awk '/ LOAD / { n++; if ($NF != "0x4000" && $NF != "0x10000") bad=1 } END { exit (!n || bad) }'
}

build_target() {
    local target="$1" build="${BUILD_ROOT}/$1" out="${DIST_ROOT}/$1"
    local cc triple processor strip builtin mathlib libtriple
    local arch_options=()
    mkdir -p "$build" "$out"
    case "$target" in
        linux-*)
            case "$target" in
                linux-x64) triple=x86_64-linux-android; processor=x86_64 ;;
                linux-x86) triple=i686-linux-android; processor=i686 ;;
                linux-arm64) triple=aarch64-linux-android; processor=aarch64 ;;
                linux-armv7a) triple=armv7a-linux-androideabi; processor=armv7-a ;;
            esac
            cc="${NDK_TOOLCHAIN}/${triple}${NDK_API}-clang"
            case "$target" in
                linux-arm*) arch_options=(-DOPUS_MAY_HAVE_NEON=OFF -DOPUS_PRESUME_NEON=ON) ;;
                *) arch_options=(-DOPUS_X86_MAY_HAVE_SSE=ON -DOPUS_X86_MAY_HAVE_SSE2=ON
                    -DOPUS_X86_MAY_HAVE_SSE4_1=ON -DOPUS_X86_MAY_HAVE_AVX2=ON
                    -DOPUS_X86_PRESUME_SSE4_1=OFF -DOPUS_X86_PRESUME_AVX2=OFF) ;;
            esac
            [[ -x "$cc" ]] || { echo "Missing NDK compiler: $cc" >&2; return 1; }
            cp "$NDK_ROOT/NOTICE" "$DIST_ROOT/NOTICE.ndk"
            cp "$NDK_ROOT/NOTICE.toolchain" "$DIST_ROOT/NOTICE.ndk-toolchain"
            cmake -S "$SRC_DIR" -B "$build" -G Ninja "${COMMON[@]}" "${arch_options[@]}" \
                -DCMAKE_SYSTEM_NAME=Generic -DCMAKE_SYSTEM_PROCESSOR="$processor" \
                -DCMAKE_C_COMPILER="$cc" -DCMAKE_TRY_COMPILE_TARGET_TYPE=STATIC_LIBRARY \
                -DCMAKE_C_FLAGS='-fPIC -fvisibility=hidden -fno-stack-protector -U_FORTIFY_SOURCE -D_FORTIFY_SOURCE=0'
            cmake --build "$build" --parallel "$JOBS"
            builtin="$("$cc" --rtlib=compiler-rt -print-libgcc-file-name)"
            libtriple="$triple"
            if [[ "$target" == linux-armv7a ]]; then libtriple=arm-linux-androideabi; fi
            mathlib="${NDK_TOOLCHAIN}/../sysroot/usr/lib/${libtriple}/libm.a"
            "$cc" -shared -nostdlib -Wl,-soname,libopus.so -Wl,-Bsymbolic \
                "-Wl,--version-script=${SCRIPT_DIR}/libopus.ver" \
                -Wl,-z,max-page-size=16384 "-Wl,--exclude-libs,$(basename "$builtin"):libm.a" \
                -Wl,--whole-archive "$build/libopus.a" -Wl,--no-whole-archive \
                "$mathlib" "$builtin" -o "$out/libopus.so"
            "${NDK_TOOLCHAIN}/llvm-strip" --strip-all "$out/libopus.so"
            check_elf "$out/libopus.so" "$target"
            ;;
        mingw-*)
            if [[ "$target" == mingw-x64 ]]; then
                cc="${MINGW_X64_CC:-x86_64-w64-mingw32-gcc}"
            else
                cc="${MINGW_X86_CC:-i686-w64-mingw32-gcc}"
            fi
            command -v "$cc" >/dev/null
            cmake -S "$SRC_DIR" -B "$build" -G Ninja "${COMMON[@]}" \
                -DCMAKE_SYSTEM_NAME=Windows -DCMAKE_C_COMPILER="$cc" \
                -DBUILD_SHARED_LIBS=ON -DOPUS_BUILD_SHARED_LIBRARY=ON \
                -DCMAKE_SHARED_LINKER_FLAGS=-static-libgcc
            cmake --build "$build" --parallel "$JOBS"
            cp "$build/libopus.dll" "$out/libopus.dll"
            cp "$build/libopus.dll.a" "$out/libopus.dll.a"
            strip="${cc%gcc}strip"
            "$strip" --strip-all "$out/libopus.dll"
            ;;
        wasm)
            local emcc="${EMCC:-emcc}" emcmake="${EMCMAKE:-emcmake}"
            "$emcmake" cmake -S "$SRC_DIR" -B "$build" -G Ninja "${COMMON[@]}"
            cmake --build "$build" --parallel "$JOBS"
            "$emcc" "$build/libopus.a" -Os --no-entry -s MODULARIZE=1 \
                -s EXPORT_ES6=1 -s EXPORT_NAME=createLibopus -s ALLOW_MEMORY_GROWTH=1 \
                -s EXPORTED_FUNCTIONS='["_opus_encoder_create","_opus_encoder_destroy","_opus_encode","_opus_encode_float","_opus_encoder_ctl","_opus_decoder_create","_opus_decoder_destroy","_opus_decode","_opus_decode_float","_opus_decoder_ctl","_opus_packet_get_nb_samples","_opus_get_version_string","_opus_strerror","_malloc","_free"]' \
                -s EXPORTED_RUNTIME_METHODS='["ccall","cwrap","getValue","setValue","UTF8ToString"]' \
                -o "$out/libopus.js"
            ;;
    esac
    echo "Built $target -> $out"
}

main() {
    local positional=() arg
    for arg in "$@"; do
        case "$arg" in
            --disable-simd) DISABLE_SIMD=ON ;;
            --enable-simd) DISABLE_SIMD=OFF ;;
            *) positional+=("$arg") ;;
        esac
    done
    set -- "${positional[@]}"
    if [[ "${1:-}" == list ]]; then printf '%s\n' "${TARGETS[@]}"; return; fi
    if [[ "${1:-}" == --help || "${1:-}" == -h ]]; then
        echo "用法: $0 [--enable-simd | --disable-simd] [list | 目标 ...]"
        echo "默认构建全部目标，启用 SIMD；--disable-simd 关闭上游专用 SIMD 实现。"
        echo "Overrides: OPUS_SRC_DIR BUILD_ROOT DIST_ROOT NDK_ROOT NDK_API JOBS MINGW_X64_CC MINGW_X86_CC EMCC EMCMAKE"
        return
    fi
    local selected=("$@") target
    if [[ ${#selected[@]} == 0 ]]; then selected=("${TARGETS[@]}"); fi
    for target in "${selected[@]}"; do
        [[ " ${TARGETS[*]} " == *" $target "* ]] || { echo "Unknown target: $target" >&2; return 1; }
    done
    [[ -f "$SRC_DIR/CMakeLists.txt" ]] || { echo 'Run: git submodule update --init --recursive' >&2; return 1; }
    command -v cmake >/dev/null
    command -v ninja >/dev/null
    COMMON+=("-DOPUS_DISABLE_INTRINSICS=$DISABLE_SIMD")
    mkdir -p "$DIST_ROOT/include/opus"
    cp "$SRC_DIR"/include/*.h "$DIST_ROOT/include/opus/"
    cp "$SRC_DIR/COPYING" "$DIST_ROOT/COPYING.opus"
    for target in "${selected[@]}"; do build_target "$target"; done
}
main "$@"
