# 09-packaging-deployment.md — 打包与发布

## 1. 概述

本文档描述大乐透 APP 从开发构建到正式发布的全流程。构建使用 gioui 官方工具 **`gogio`**（`gioui.org/cmd/gogio`），它封装了 Android 打包的完整流程（编译、资源处理、APK 组装、签名）。

**目标产出物**：

| 产出物 | 用途 | 触发时机 |
|--------|------|---------|
| `go run .` | 桌面端 UI 快速调试 | 本地开发 |
| `lottery-debug.apk` | Android 调试安装包（自动签名） | `make apk` |
| `lottery-release.apk` | 正式发布包（手动签名） | 版本发布 |
| `lottery-release.aab` | Google Play 发布包 | 版本发布 |

## 2. 开发构建

### 2.1 桌面端调试

gioui 支持在 Linux 桌面端直接运行，无需编译到 Android，UI 效果与移动端基本一致：

```bash
# 项目根目录
cd /opt/codes/lottery

# 运行桌面端
go run .
```

桌面端运行特点：
- 窗口标题：`大乐透助手`
- 默认窗口尺寸：`400×700`（模拟手机竖屏比例）
- 键盘快捷键可用（方便 UI 调试）

### 2.2 Android 构建工具：gogio

`gogio` 是 gioui 官方提供的 Android/iOS 构建工具，一键完成所有打包工作：

| 特性 | 说明 |
|------|------|
| 安装 | `go install gioui.org/cmd/gogio@latest` |
| 构建 | `gogio -target android -appid com.lottery.app -o app.apk .` |
| 图标 | `-icon` 指定 PNG 或 SVG 作为应用图标 |
| 版本 | `-version` 设置版本号（格式：`major.minor.patch.versioncode`）|
| 架构 | `-arch` 指定架构（默认全架构）|
| 签名 | `-signkey` + `-signpass` 指定签名密钥 |
| minSdk | `-minsdk` 设置最低 SDK 版本（默认 16）|
| targetSdk | `-targetsdk` 设置目标 SDK 版本 |

### 2.3 构建调试 APK

```bash
# ARM64 单架构（推荐开发调试，包体更小）
gogio -target android -arch arm64 \
  -appid com.lottery.app \
  -icon android/ic_launcher/mipmap-xxxhdpi/ic_launcher.png \
  -o lottery-debug.apk .

# 全架构（兼容所有设备）
gogio -target android \
  -appid com.lottery.app \
  -icon android/ic_launcher/mipmap-xxxhdpi/ic_launcher.png \
  -o lottery-debug.apk .
```

> 首次运行需安装 gogio：`go install gioui.org/cmd/gogio@latest`
>
> `gogio` 自动处理以下环节：
> - 交叉编译 Go 代码为 Android 原生库
> - 生成 AndroidManifest.xml（含 INTERNET 权限）
> - 根据 `-icon` 指定的 PNG 生成各分辨率图标（含自适应图标）
> - 生成 debug 签名并打包 APK
> - 无需手动配置 AndroidManifest 或资源目录
>
> **注意**：`-icon` 参数要求 **PNG** 格式，不支持 SVG。
> 项目中 `make icon` 可从 SVG 源文件生成 PNG；`make apk` 会自动触发此步骤。

### 2.4 安装到设备

```bash
# 连接设备后安装
adb install -r lottery-debug.apk

# 查看运行日志
adb logcat -s "lottery" --format=tag

# 过滤特定日志级别
adb logcat -s "lottery" *:E   # 仅 Error
adb logcat -s "lottery" *:W   # Warn 及以上
```

## 3. gogio 构建参数详解

### 3.1 必选参数

| 参数 | 说明 | 示例 |
|------|------|------|
| `-target` | 目标平台：`android` / `ios` / `js` / `macos` / `windows` | `-target android` |
| `-appid` | Android package name 或 iOS bundle ID | `-appid com.lottery.app` |
| `<package>` | 要构建的 Go 包路径（`.` 表示当前目录） | `.` |

### 3.2 可选参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-o` | 输出文件路径 | 自动生成 |
| `-arch` | 逗号分隔的 GOARCH 列表 | 所有支持的架构 |
| `-icon` | 应用图标（PNG 或 SVG 路径） | `appicon.png`（若存在） |
| `-version` | 版本号，格式 `major.minor.patch.versioncode` | 无 |
| `-minsdk` | 最低 Android API 级别 | 16 |
| `-targetsdk` | 目标 Android API 级别 | 编译 SDK 版本 |
| `-buildmode` | 构建模式：`exe`（APK）或 `archive`（AAR） | `exe` |
| `-signkey` | 签名密钥库路径 | debug 签名 |
| `-signpass` | 密钥库密码 | 环境变量 `GOGIO_SIGNPASS` |
| `-ldflags` | 传递给 Go 编译器的链接参数 | — |
| `-tags` | 传递给 Go 编译器的构建标签 | — |
| `-work` | 打印工作目录路径，不自动删除 | 关闭 |
| `-x` | 打印所有外部命令 | 关闭 |

### 3.3 参数示例

```bash
# 完整参数示例
gogio -target android \
  -appid com.lottery.app \
  -arch arm64 \
  -icon android/ic_launcher/mipmap-xxxhdpi/ic_launcher.png \
  -version 1.0.0.1 \
  -minsdk 21 \
  -targetsdk 33 \
  -o lottery-release.apk \
  .
```

## 4. 版本管理

### 4.1 版本号体系

采用 **三位语义化版本 + Android versionCode**，整体格式：

```
主版本.次版本.补丁.versionCode
示例：1.2.3.10203
```

| 位 | 说明 | 递增加1的时机 |
|---|------|-------------|
| 主版本 | 架构重构、不兼容 UI 变更 | 1.x.x.x → 2.0.0.20000 |
| 次版本 | 功能新增、UI 调整、向后兼容 | 1.1.0.10100 → 1.2.0.10200 |
| 补丁 | Bug 修复、性能优化 | 1.1.0.10100 → 1.1.1.10101 |
| versionCode | 递增整数，仅用于 Android 内部版本比较 | 每次构建 +1 |

### 4.2 versionCode 编码规则

```
公式：主版本 × 10000 + 次版本 × 100 + 补丁
示例：
  1.0.0 → versionCode 10000
  1.2.3 → versionCode 10203
  2.0.0 → versionCode 20000
```

### 4.3 gogio 版本注入

在 gogio 中通过 `-version` 参数直接指定：

```bash
gogio -target android \
  -appid com.lottery.app \
  -version 1.2.3.10203 \
  -o lottery-release.apk .
```

> `gogio` 自动将 versionName 设为 `1.2.3`、versionCode 设为 `10203`。

## 5. 发布构建

### 5.1 签名准备

Android 应用正式发布需要签名。生成签名密钥：

```bash
# 生成密钥库（仅一次）
keytool -genkey -v \
    -keystore lottery-release.keystore \
    -alias lottery \
    -keyalg RSA \
    -keysize 2048 \
    -validity 10000

# 注意：keystore 文件应妥善保管，不得提交到 Git 仓库
```

### 5.2 构建发布 APK（已签名）

```bash
export GOGIO_SIGNPASS="你的密钥密码"

gogio -target android \
  -appid com.lottery.app \
  -arch arm64 \
  -icon android/ic_launcher/mipmap-xxxhdpi/ic_launcher.png \
  -version 1.0.0.10000 \
  -minsdk 21 \
  -targetsdk 33 \
  -signkey lottery-release.keystore \
  -signpass "$GOGIO_SIGNPASS" \
  -o lottery-release.apk \
  .
```

> `gogio` 一次性完成构建 + 签名 + 对齐，无需单独的 jarsigner 和 zipalign 步骤。

### 5.3 构建 AAB（Google Play 格式）

当前版本 `gogio` 直接输出 APK。如需 AAB（Android App Bundle），有以下方案：

```bash
# 方案一：使用 bundletool 将 APK 转换为 AAB
# 需安装 Android SDK build-tools
java -jar bundletool.jar build-bundle \
  --modules=lottery-release.apk \
  --output=lottery-release.aab

# 方案二：直接上传已签名的 APK 到 Google Play
# Google Play Console 仍接受 APK 格式
```

### 5.4 密钥管理安全规范

| 要求 | 说明 |
|------|------|
| 密钥库不提交 Git | 在 `.gitignore` 中排除 `*.keystore` 和 `*.jks` |
| 密钥密码不写代码 | 使用环境变量 `GOGIO_SIGNPASS` 注入 |
| 密钥备份 | 丢失密钥将无法更新已发布的应用 |
| 密钥轮换 | 如需换密钥，使用 Android 密钥轮换机制 |

## 6. 构建时 API 配置注入

### 6.1 概述

数据源 API 地址通过 `-ldflags` 在构建时注入，不硬编码。同一份代码可针对不同数据源构建。

原理：`internal/config/config.go` 中声明全局变量，构建时通过 `-X` 标志覆盖。

### 6.2 可用配置变量

| 变量路径 | 用途 | 默认值 |
|---------|------|--------|
| `github.com/user/lottery/internal/config.APIBaseURL` | 数据源 API 基础 URL | `https://www.cwl.gov.cn` |
| `github.com/user/lottery/internal/config.DataSource` | 数据源类型 | `cwl` |

> `DataSource` 取值：`cwl`（福彩官网接口）、`mock`（内置模拟数据，无网络时可用）

### 6.3 桌面端注入

```bash
# 使用在线数据（需 API 可用）
go run -ldflags="-X 'github.com/user/lottery/internal/config.APIBaseURL=https://api.example.com'" .

# 使用模拟数据（离线可用）
go run -ldflags="-X 'github.com/user/lottery/internal/config.DataSource=mock'" .
```

### 6.4 Makefile 配置（推荐）

在项目根目录执行：

```bash
# 使用默认 API（福彩官网）
make run

# 使用自定义 API 地址
API_BASE_URL=https://api.example.com make run

# 使用模拟数据（离线构建）
DATA_SOURCE=mock make run

# 构建 APK 时注入自定义 API
API_BASE_URL=https://api.example.com make apk
```

### 6.5 APK 构建时注入

桌面端和 Android 构建均支持 `-ldflags`：

```bash
# Android APK 构建 + 自定义 API 地址
gogio -target android -arch arm64 \
  -appid com.lottery.app \
  -icon android/ic_launcher/ic_launcher.png \
  -ldflags="-X 'github.com/user/lottery/internal/config.APIBaseURL=https://api.example.com' \
            -X 'github.com/user/lottery/internal/config.DataSource=mock'" \
  -o lottery-debug.apk .
```

> 注：通过 Makefile 的 `API_BASE_URL` 和 `DATA_SOURCE` 变量可自动注入，无需手写 ldflags 字符串。

## 7. 构建优化

### 6.1 APK 瘦身

| 优化项 | 措施 | 预期效果 |
|--------|------|---------|
| 架构裁剪 | 仅构建 `arm64`（`-arch arm64`） | APK 减小约 40% |
| 编译优化 | gogio 默认已启用 `-trimpath` | 减小二进制体积 |

```bash
# 生产构建命令（gogio 默认含 -trimpath）
gogio -target android \
  -appid com.lottery.app \
  -arch arm64 \
  -icon android/ic_launcher/mipmap-xxxhdpi/ic_launcher.png \
  -o lottery-release.apk \
  .
```

### 6.2 安全性说明

- Go 编译为原生机器码，不需要 ProGuard/R8 混淆
- `gogio` 输出的 APK 已包含必要的 Android 权限声明（INTERNET + ACCESS_NETWORK_STATE）
- 应用默认不请求位置、存储、相机等敏感权限

## 7. Makefile 自动化

项目中提供了 `Makefile` 将常用操作自动化（详见项目根目录 `Makefile`）。

| 命令 | 作用 |
|------|------|
| `make run` | 桌面端运行 |
| `make build` | 桌面端编译 |
| `make apk` | 构建 Android 调试 APK |
| `make install` | 构建并安装到连接的设备 |
| `make release` | 构建已签名的发布 APK |
| `make icon` | 从 SVG 生成各分辨率 PNG |
| `make test` | 运行全部测试 |
| `make lint` | 静态分析 |
| `make fmt` | 格式化代码 |
| `make clean` | 清理构建产物 |

## 8. CI/CD 集成

### 8.1 CI 构建脚本

```yaml
# .github/workflows/build.yml
name: Build

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.26.1'

      - name: Install gogio
        run: go install gioui.org/cmd/gogio@latest

      - name: Build Android APK
        run: |
          gogio -target android \
            -arch arm64 \
            -appid com.lottery.app \
            -icon android/ic_launcher/mipmap-xxxhdpi/ic_launcher.png \
            -o lottery-debug.apk \
            .

      - name: Upload APK artifact
        uses: actions/upload-artifact@v4
        with:
          name: lottery-apk
          path: lottery-debug.apk
```

### 8.2 版本标签与发布

```bash
# 创建版本标签
git tag -a v1.0.0 -m "v1.0.0 首次发布"

# 推送标签（触发 CI 发布构建）
git push origin v1.0.0
```

## 9. 环境准备

### 9.1 依赖安装

```bash
# 1. Go 工具链（已安装 go1.26.1）
go version

# 2. gogio 构建工具
go install gioui.org/cmd/gogio@latest

# 3. 验证 gogio 可用
gogio -help

# 4. 确认 Android SDK 路径
echo $ANDROID_HOME        # 应指向 ~/Android/Sdk
echo $ANDROID_NDK_HOME    # 应指向 NDK 目录

# 5. adb 工具（可选，用于连接真机）
# 位于 $ANDROID_HOME/platform-tools/adb
```

### 9.2 环境变量

```bash
# 建议写入 ~/.bashrc 或 ~/.profile
export ANDROID_HOME=$HOME/Android/Sdk
export ANDROID_NDK_HOME=$ANDROID_HOME/ndk/23.1.7779620
export PATH=$PATH:$ANDROID_HOME/platform-tools:$HOME/go/bin
```

## 10. .gitignore 规则

确保构建产物和敏感文件不提交到 Git：

```gitignore
# Android 构建产物
*.apk
*.aab

# 签名信息
*.keystore
*.jks

# IDE
.idea/
*.iml
.vscode/

# Go
vendor/

# 操作系统
.DS_Store
Thumbs.db
```

## 11. 快速参考

### 调试 → 跑 → 看

```bash
# 桌面调试（秒级）
make run

# Android 构建（约 2-3 分钟）
make apk

# 构建+安装（需连接设备）
make install

# 看日志
adb logcat -s "lottery" --format=tag
```

### 发布前检查清单

- [ ] versionCode 已递增（`-version` 参数）
- [ ] 测试已全量通过（`make test`）
- [ ] 桌面端 UI 验证正常
- [ ] 真机安装验证 UI 正常
- [ ] 所有网络请求正常（cwl.gov.cn 可达）
- [ ] 签名密钥可用
- [ ] 图标已确认（`-icon android/ic_launcher/mipmap-xxxhdpi/ic_launcher.png`）
- [ ] 隐私政策已更新（如首次发布）

## 12. 附录：实际构建输出示例

```
$ gogio -target android \
  -arch arm64 \
  -appid com.lottery.app \
  -icon android/ic_launcher/mipmap-xxxhdpi/ic_launcher.png \
  -o lottery-debug.apk .

# 无输出 = 构建成功
# 产出 lottery-debug.apk（约 15MB）
```

APK 内部结构确认：

```
unzip -l lottery-debug.apk
  res/mipmap-mdpi-v4/ic_launcher_adaptive.png   → 自适应图标
  res/mipmap-hdpi-v4/ic_launcher.png             → 桌面图标
  res/mipmap-xhdpi-v4/ic_launcher.png
  res/mipmap-xxhdpi-v4/ic_launcher.png
  res/mipmap-xxxhdpi-v4/ic_launcher.png          → 高清图标
  AndroidManifest.xml                            → 自动生成（含 INTERNET 权限）
  classes.dex                                     → Android 运行时
  lib/arm64-v8a/libgojni.so                       → Go 原生代码
```
