# 08-mobile-dev-deps.md — 移动端开发环境依赖

## 1. 环境概述

本 APP 基于 Go + gioui 构建移动端应用。gioui 通过 gomobile 编译为 Android/iOS 原生应用，需要以下依赖环境。

## 2. Android 开发环境

### 2.1 核心依赖

| 组件 | 版本要求 | 用途 |
|------|---------|------|
| Android SDK | API 24+ (Android 7.0) | Android 编译目标 |
| Android NDK | 25+ | C 交叉编译（gioui 依赖） |
| gomobile | 最新 | Go → Android .apk 编译 |
| Java | 1.8.0_492 | Android SDK 工具链依赖 |

### 2.2 环境变量

```bash
# Android SDK 路径（项目实际路径 ~/Android）
export ANDROID_HOME=$HOME/Android/Sdk
export ANDROID_NDK_HOME=$ANDROID_HOME/ndk/25.2.9519653
export PATH=$PATH:$ANDROID_HOME/tools:$ANDROID_HOME/platform-tools
```

### 2.3 安装 gomobile

```bash
go install golang.org/x/mobile/cmd/gomobile@latest
gomobile init
```

## 3. 编译与运行

### 3.1 编译 Android APK

```bash
# 在项目根目录执行
gomobile build -target=android -o lottery.apk .
```

### 3.2 调试运行

```bash
# 安装到已连接的设备/模拟器
adb install -r lottery.apk

# 查看日志
adb logcat | grep "lottery"
```

### 3.3 支持的编译目标

| 目标 | 命令 | 说明 |
|------|------|------|
| Android ARM64 | `-target=android/arm64` | 主流手机架构 |
| Android ARMv7a | `-target=android/arm` | 老旧设备兼容 |
| Android x86_64 | `-target=android/amd64` | 模拟器 |

## 4. 本地开发验证

### 4.1 桌面端快速验证

gioui 支持在桌面端（Linux/Windows/macOS）直接运行做 UI 调试，无需每次编译到手机：

```bash
go run .
```

### 4.2 交叉编译检查

```bash
# 语法检查
GOOS=android GOARCH=arm64 go build ./...

# 完整构建
gomobile build -target=android/arm64 -o lottery.apk .
```

## 5. 常见问题

| 问题 | 排查方式 |
|------|---------|
| `gomobile` 命令未找到 | 确认 `$GOPATH/bin` 在 PATH 中 |
| `NDK not found` | 确认 `ANDROID_NDK_HOME` 指向正确的 NDK 路径 |
| `javac not found` | 确认 `$JAVA_HOME` 已设置且指向 JDK 8 |
| 编译时报 `*.syso` 错误 | 确认 NDK 版本 ≥ 25 |
| 模拟器无法安装 APK | 确认 API level ≥ 24 |

## 6. 实际路径

```
Android SDK:  ~/Android/Sdk
NDK:         ~/Android/Sdk/ndk/
Java:        $JAVA_HOME (jdk 1.8.0_492)
```
