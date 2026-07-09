# Makefile — 大乐透助手构建自动化
#
# 使用 gogio（gioui.org/cmd/gogio）构建 Android APK。
# 桌面端直接使用 go run/build。

# ─── 项目配置 ────────────────────────────────────────────
APP_NAME      := 大乐透助手
APP_ID        := com.lottery.app
ICON_SRC      := ui/icon/app_icon.svg
ICON_PNG      := android/ic_launcher/ic_launcher.png
ICON_PNG_DIR  := android/ic_launcher

# ─── Android SDK / NDK ──────────────────────────────────
ANDROID_HOME    ?= $(HOME)/Android/Sdk
ANDROID_NDK_HOME ?= $(ANDROID_HOME)/ndk/23.1.7779620
ADB             ?= $(ANDROID_HOME)/platform-tools/adb

# ─── 构建时配置 ──────────────────────────────────────
# API_BASE_URL 可通过 ldflags 覆盖，默认由 client 包硬编码兜底
API_BASE_URL  ?= ""

# ─── 版本信息 ──────────────────────────────────────────
VERSION_MAJOR ?= 1
VERSION_MINOR ?= 0
VERSION_PATCH ?= 0
VERSION_CODE  ?= 10000
VERSION       := $(VERSION_MAJOR).$(VERSION_MINOR).$(VERSION_PATCH).$(VERSION_CODE)

# ─── 构建产物 ──────────────────────────────────────────
DEBUG_APK     := lottery-debug.apk
RELEASE_APK   := lottery-release.apk
RELEASE_KEY   := lottery-release.keystore

# ─── 环境检测 ──────────────────────────────────────────
GOGIO := $(shell PATH="$$PATH:$$HOME/go/bin" which gogio 2>/dev/null || echo "")

.PHONY: all run build apk install release icon test lint fmt clean help

# ─── 默认目标 ──────────────────────────────────────────
all: help

# ─── 桌面开发 ──────────────────────────────────────────
run:  ## 桌面端运行（UI 快速调试）
	go run .

build:  ## 桌面端编译验证
	go build .

# ─── 图标生成（apk 前置依赖） ──────────────────────────
$(ICON_PNG): $(ICON_SRC)
	python3 internal/tools/gen_icons.py "$(ICON_SRC)" "$(ICON_PNG_DIR)"

icon: $(ICON_PNG)  ## 从 SVG 生成 Android 各分辨率 PNG 图标
	@ls -lh $(ICON_PNG_DIR)/*/ic_launcher.png

# ─── Android 构建 ──────────────────────────────────────
apk: check-gogio check-android $(ICON_PNG)  ## 构建 Android 调试 APK
	export PATH="$$PATH:$$HOME/go/bin" ANDROID_HOME="$(ANDROID_HOME)" ANDROID_NDK_HOME="$(ANDROID_NDK_HOME)" \
	  && gogio -target android -arch arm64 -appid "$(APP_ID)" -icon "$(ICON_PNG)" -minsdk 21  -o "$(DEBUG_APK)" .
	@echo "✅ $(DEBUG_APK) 构建完成"
	@ls -lh "$(DEBUG_APK)"

install: apk  ## 构建并安装 Android APK 到连接的设备
	@if [ -x "$(ADB)" ]; then \
	  echo "📱 安装 $(DEBUG_APK) 到设备..."; \
	  $(ADB) install -r "$(DEBUG_APK)" && echo "✅ 安装成功"; \
	else \
	  echo "⚠️ adb 未找到，请手动安装: adb install -r $(DEBUG_APK)"; \
	fi

release: check-gogio check-android check-key $(ICON_PNG)  ## 构建已签名的发布 APK
	@if [ -z "$(GOGIO_SIGNPASS)" ]; then \
	  echo "❌ 请设置 GOGIO_SIGNPASS 环境变量"; \
	  exit 1; \
	fi
	export PATH="$$PATH:$$HOME/go/bin" ANDROID_HOME="$(ANDROID_HOME)" ANDROID_NDK_HOME="$(ANDROID_NDK_HOME)" \
	  && gogio -target android -arch arm64 -appid "$(APP_ID)" -icon "$(ICON_PNG)"  \
	    -version "$(VERSION)" -minsdk 21 -targetsdk 33 \
	    -signkey "$(RELEASE_KEY)" -signpass "$(GOGIO_SIGNPASS)" -o "$(RELEASE_APK)" .
	@echo "✅ $(RELEASE_APK) 构建完成（已签名）"

# ─── 代码质量 ──────────────────────────────────────────
test:  ## 运行全部测试
	go test -count=1 -race -coverprofile=coverage.out ./...
	@echo ""
	@go tool cover -func=coverage.out | grep total

lint:  ## Go 静态分析
	go vet ./...

fmt:  ## 检查 Go 代码格式
	@echo "未格式化的文件:"
	@gofmt -l .

# ─── 清理 ──────────────────────────────────────────────
clean:  ## 清理构建产物
	rm -f "$(DEBUG_APK)" "$(RELEASE_APK)" coverage.out
	rm -rf "$(ICON_PNG_DIR)"
	@echo "✅ 清理完成"

# ─── 环境检查 ──────────────────────────────────────────
check-gogio:
	@if [ -z "$(GOGIO)" ]; then \
	  echo "❌ gogio 未安装，执行: go install gioui.org/cmd/gogio@latest"; \
	  exit 1; \
	fi

check-android:
	@if [ ! -d "$(ANDROID_HOME)" ]; then \
	  echo "❌ ANDROID_HOME 路径不存在: $(ANDROID_HOME)"; \
	  exit 1; \
	fi
	@if [ ! -d "$(ANDROID_NDK_HOME)" ]; then \
	  echo "⚠️  ANDROID_NDK_HOME 路径不存在: $(ANDROID_NDK_HOME)"; \
	fi

check-key:
	@if [ ! -f "$(RELEASE_KEY)" ]; then \
	  echo "⚠️  签名密钥不存在: $(RELEASE_KEY)"; \
	  echo "   生成命令: keytool -genkey -v -keystore $(RELEASE_KEY) -alias lottery -keyalg RSA -keysize 2048 -validity 10000"; \
	fi

# ─── 帮助 ──────────────────────────────────────────────
help:  ## 显示帮助信息
	@echo "═══════════════════════════════════════"
	@echo "  大乐透助手 — 构建快捷命令"
	@echo "═══════════════════════════════════════"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "── 当前配置 ──"
	@echo "  APP_ID       = $(APP_ID)"
	@echo "  VERSION      = $(VERSION)"
	@echo "  ICON_SRC     = $(ICON_SRC)"
	@echo "  ICON_PNG     = $(ICON_PNG)"
	@echo "  ANDROID_HOME = $(ANDROID_HOME)"
	@echo ""
	@echo "── 环境变量 ──"
	@echo "  GOGIO_SIGNPASS  发布签名密码（仅 release 需要）"
