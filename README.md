# opuspurego

通过 purego 动态调用 libopus，沿用 ambepurego 的库与编解码器生命周期设计。
绑定本身不使用 CGo，运行时需要动态库。处理原始 Opus 数据包，不处理 Ogg 容器。
`opus/` 通过 Git 子模块引用官方源码，当前使用的版本是 `v1.6.1`。

## 构建与优化开关

```sh
git submodule update --init --recursive
./build-libopus.sh list
./build-libopus.sh                     # 全部目标，默认启用硬件优化
./build-libopus.sh linux-x64 linux-armv7a
./build-libopus.sh --disable-simd linux-x64  # 关闭上游 SIMD 专用实现
./build-libopus.sh --enable-simd linux-x64   # 显式恢复默认设置
```

默认使用定点内核、Release 优化等级，保留浮点 PCM 接口。定点也有硬件优化：

| 架构 | 默认优化 |
| --- | --- |
| x64/x86，包括 Windows | SSE、SSE2、SSE4.1、AVX2/FMA 专用路径，上游运行时 CPU 检测选择；不全局强制 AVX2 |
| ARM64 | 架构基线 NEON |
| ARMv7 | NEON，产物要求 CPU 支持 NEON |
| WASM | 上游通用实现，未启用 WASM SIMD |

采用上游 CMake 集成的 intrinsics 实现，编译后生成向量指令；不声称包含源码树
中的所有独立汇编文件。ARMv7 使用 NDK 默认的 NEON 目标，避免引入基于
`/proc/cpuinfo` 的检测依赖。DRED、OSCE 关闭，浮点输入由定点内核量化。

`--disable-simd` 对应 `OPUS_DISABLE_INTRINSICS=ON`，关闭 Opus 的专用 SIMD
实现，不禁用编译器自动向量化，也不改变 NDK 的最低 CPU 要求。同一构建目录
可以切换开关；默认输出位置相同，后一次构建覆盖前一次产物。对比两种构建时
应分别设置 `BUILD_ROOT` 和 `DIST_ROOT`。

需要 CMake、Ninja。Linux/Android 使用 NDK，Windows 使用 MinGW GCC，WASM
使用 Emscripten。默认 NDK 为 `/home/yj/Android/Sdk/ndk/28.2.13676358`，API 21。
可通过 `BUILD_ROOT`、`DIST_ROOT`、`OPUS_SRC_DIR`、`NDK_ROOT`、`NDK_API`、
`JOBS` 覆盖路径和并发数，编译器配置见 `--help`。

## 产物与兼容性

生成文件放在 `/tmp/opuspurego`，发行目录默认为 `/tmp/opuspurego/dist/libopus`。

| 目标 | 目标子目录中的文件 |
| --- | --- |
| linux-x64、linux-x86、linux-arm64、linux-armv7a | libopus.so |
| mingw-x64、mingw-x86 | libopus.dll、libopus.dll.a |
| wasm | libopus.js、libopus.wasm |

同架构 Linux/Android 共用一份 `.so`。所需数学函数和编译器辅助函数静态链接，
无 `NEEDED` 和带版本导入，支持 16KB 页对齐，只导出公开 API。脚本对导入符号、
动态依赖及对齐做强制检查。ARMv7 还依赖 NDK 数学库使用的 libc 栈保护符号。

Linux x64/x86 原生运行测试；ARM64/ARMv7 使用 QEMU 加 musl 运行测试。
Android 真机尚未验证，无动态依赖不等于兼容所有 libc 版本。Windows 产物
使用标准 CRT、静态 libgcc，当前 MinGW 依赖 UCRT，通常由 Windows 10 及
更新版本提供；Windows 测试使用 Wine。WASM 通过 Node 测试，不能通过 Go
动态库接口加载。

头文件在 `include/opus/`，许可证为 `COPYING.opus`、`NOTICE.ndk` 和
`NOTICE.ndk-toolchain`，分发时保留适用声明。

## Go 用法

```go
package main

import (
    "log"
    "github.com/yangjuncode/opuspurego"
)

func main() {
    lib, err := opuspurego.OpenDist("/tmp/opuspurego/dist/libopus")
    if err != nil { log.Fatal(err) }
    defer lib.Close()
    enc, err := lib.NewEncoder(48000, 1, opuspurego.ApplicationVoIP)
    if err != nil { log.Fatal(err) }
    defer enc.Close()
    dec, err := lib.NewDecoder(48000, 1)
    if err != nil { log.Fatal(err) }
    defer dec.Close()
    if err := enc.SetBitrate(24000); err != nil { log.Fatal(err) }
    pcm := make([]int16, 960) // 48 kHz、单声道、20 ms
    packet := make([]byte, 4000)
    n, err := enc.Encode(packet, pcm)
    if err != nil { log.Fatal(err) }
    output := make([]int16, 960)
    samples, err := dec.Decode(output, packet[:n], false)
    if err != nil { log.Fatal(err) }
    log.Printf("数据包 %d 字节，每声道 %d 个采样", n, samples)
}
```

`Open(path)` 指定库路径。`OpenDefault()` 在 Linux/BSD 使用 `libopus.so.0`，
Android 使用 `libopus.so`，Apple 使用 `libopus.dylib`，Windows 使用
`libopus.dll`。`OpenDist` 将 Android 映射到相应 `linux-*` 目录。

采样率支持 8000、12000、16000、24000、48000 Hz，声道数为 1 或 2。
双声道 PCM 交错排列。编码帧长支持 2.5、5、10、20、40、60 ms，根据输入长度
推导。编码返回字节数，解码返回每声道采样数，有效输出为
`dst[:samples*channels]`。任意接收包的输出缓冲区应容纳每声道最多 120 ms。

`DecodeLost(dst)` 或 `Decode(dst,nil,false)` 进行 PLC，补偿长度由输出长度
决定。`Decode(dst,nextPacket,true)` 用下一包恢复上一丢失帧，输出长度应准确
对应丢失帧；随后还要正常解码下一包。没有 FEC 时回退到 PLC。
`EncodeFloat`、`DecodeFloat` 使用标称 [-1,1] 的 float32 PCM。

支持码率、复杂度、VBR、DTX、带内 FEC、预期丢包率、前瞻长度和重置。
`PacketSamples` 查询包时长；暂未包装多流、重组包与实验控制接口。
`NativeError` 保留原生状态码，可用 `errors.As` 提取，参数错误支持
`errors.Is(err,ErrInvalidArgument)`。

先关闭编解码器，再关闭库；仍有对象存活时返回 `ErrLibraryInUse`。关闭幂等，
对象方法及关闭互斥，不要复制对象。应显式关闭，finalizer 只是兜底。
并发安全不保证音频帧的业务顺序。purego 固定为 `v0.11.0`，其平台限制
仍适用：Android/iOS 需要 CGo，FreeBSD/NetBSD 禁用 CGo 需要上游指定编译参数。

## 测试

```sh
CGO_ENABLED=0 LIBOPUS_DIST=/tmp/opuspurego/dist/libopus go test ./...
CGO_ENABLED=0 GOARCH=386 LIBOPUS_DIST=/tmp/opuspurego/dist/libopus go test ./...
LIBOPUS_DIST=/tmp/opuspurego/dist/libopus go test -race ./...
go vet ./...
bash tests/qemu-arm.sh
node tests/wasm-smoke.mjs
```

QEMU 脚本使用 `/home/yj/toolchains` 下 musl 工具链，可通过 `ARM64_CC`、
`ARM_CC` 替换，通过 `TEST_ROOT` 指定临时目录。测试程序启用 CGo 链接 musl，
加载现有 NDK 共享库，覆盖所有采样率、单/双声道、整数/浮点、PLC/FEC、
控制参数、非法缓冲区和并发关闭。

未设置 `LIBOPUS_DIST` 或 `LIBOPUS_PATH` 时尝试系统库，无法加载会跳过；
设置任一变量后，加载失败会使测试失败。
