# Makefile for gomobile_ipfs_0
# 参考了原版 gomobile-ipfs 项目的构建流程

# 项目目录设置
MAKEFILE_DIR = $(shell pwd)
GO_DIR = $(MAKEFILE_DIR)/go
ANDROID_DIR = $(MAKEFILE_DIR)/android
IOS_DIR = $(MAKEFILE_DIR)/ios

# 构建目录
BUILD_DIR = $(MAKEFILE_DIR)/build
ANDROID_BUILD_DIR = $(BUILD_DIR)/android
IOS_BUILD_DIR = $(BUILD_DIR)/ios

# Go绑定包路径
CORE_PACKAGE = github.com/marssuren/gomobile_ipfs_0/go/bind/core

# Android相关设置
ANDROID_MINIMUM_VERSION = 21
ANDROID_BUILD_DIR_INT = $(ANDROID_BUILD_DIR)/intermediates
ANDROID_BUILD_DIR_INT_CORE = $(ANDROID_BUILD_DIR_INT)/core
ANDROID_GOMOBILE_CACHE = "$(ANDROID_BUILD_DIR_INT_CORE)/.gomobile-cache"
ANDROID_CORE = $(ANDROID_BUILD_DIR_INT_CORE)/ipfs.aar

# iOS相关设置
IOS_BUILD_DIR_INT = $(IOS_BUILD_DIR)/intermediates
IOS_BUILD_DIR_INT_CORE = $(IOS_BUILD_DIR_INT)/core
IOS_GOMOBILE_CACHE = "$(IOS_BUILD_DIR_INT_CORE)/.gomobile-cache"
IOS_CORE = $(IOS_BUILD_DIR_INT_CORE)/Ipfs.xcframework

# 主要构建目标
.PHONY: all build_core build_core.android build_core.ios clean clean.android clean.ios 

all: build_core

# 核心库构建
build_core: build_core.android build_core.ios

# Android核心库构建
build_core.android: $(ANDROID_CORE)

$(ANDROID_CORE): $(ANDROID_BUILD_DIR_INT_CORE)
	@echo '------------------------------------'
	@echo '   Android Core: Gomobile binding   '
	@echo '------------------------------------'
	# 下载Go依赖
	cd $(GO_DIR) && go mod download
	# 初始化GoMobile
	cd $(GO_DIR) && go run golang.org/x/mobile/cmd/gomobile init
	# 创建缓存目录
	mkdir -p $(ANDROID_GOMOBILE_CACHE) android/libs
	# 运行GoMobile绑定命令，生成AAR
	cd $(GO_DIR) && go run golang.org/x/mobile/cmd/gomobile bind \
		-o $(ANDROID_CORE) \
		-v \
		-cache $(ANDROID_GOMOBILE_CACHE) \
		-target=android \
		-androidapi $(ANDROID_MINIMUM_VERSION) \
		-javapkg=org.ipfs.gomobile \
		$(CORE_PACKAGE)
	@echo 'Done!'

$(ANDROID_BUILD_DIR_INT_CORE):
	mkdir -p $(ANDROID_BUILD_DIR_INT_CORE)

# iOS核心库构建
build_core.ios: $(IOS_CORE)

$(IOS_CORE): $(IOS_BUILD_DIR_INT_CORE)
	@echo '------------------------------------'
	@echo '     iOS Core: Gomobile binding     '
	@echo '------------------------------------'
	# 下载Go依赖
	cd $(GO_DIR) && go mod download
	# 安装gobind工具
	cd $(GO_DIR) && go install golang.org/x/mobile/cmd/gobind
	# 初始化GoMobile
	cd $(GO_DIR) && go run golang.org/x/mobile/cmd/gomobile init
	# 创建目录
	mkdir -p $(IOS_GOMOBILE_CACHE) ios/Frameworks
	# 运行GoMobile绑定命令，生成XCFramework
	cd $(GO_DIR) && go run golang.org/x/mobile/cmd/gomobile bind \
			-o $(IOS_CORE) \
			-tags 'nowatchdog' \
			-cache $(IOS_GOMOBILE_CACHE) \
			-target=ios \
			$(CORE_PACKAGE)
	@echo 'Done!'

$(IOS_BUILD_DIR_INT_CORE):
	@mkdir -p $(IOS_BUILD_DIR_INT_CORE)

# 清理构建产物
clean: clean.android clean.ios

# 清理Android构建产物
clean.android:
	@echo '------------------------------------'
	@echo '  Android Core: removing build dir  '
	@echo '------------------------------------'
	rm -rf $(ANDROID_BUILD_DIR)
	# gomobile cache
	rm -rf $(ANDROID_GOMOBILE_CACHE)
	@echo 'Done!'

# 清理iOS构建产物
clean.ios:
	@echo '------------------------------------'
	@echo '    iOS Core: removing build dir    '
	@echo '------------------------------------'
	rm -rf $(IOS_BUILD_DIR)
	# gomobile cache
	rm -rf $(IOS_GOMOBILE_CACHE)
	@echo 'Done!' 