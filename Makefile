.PHONY: build build-all clean install test release docker

# 版本信息
VERSION := 1.0.0
BINARY := llm-detector
BUILD_DIR := ./build
DIST_DIR := ./dist

# 默认构建当前平台
build:
	@echo "Building $(BINARY) for current platform..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "-s -w -X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY) cmd/detector/main.go
	@echo "✓ Build complete: $(BUILD_DIR)/$(BINARY)"

# 构建所有平台
build-all:
	@echo "Building for all platforms..."
	@mkdir -p $(DIST_DIR)
	
	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X main.Version=$(VERSION)" \
		-o $(DIST_DIR)/$(BINARY)-linux-amd64 cmd/detector/main.go
	@echo "✓ Linux AMD64"
	
	# Linux ARM64
	GOOS=linux GOARCH=arm64 go build -ldflags "-s -w -X main.Version=$(VERSION)" \
		-o $(DIST_DIR)/$(BINARY)-linux-arm64 cmd/detector/main.go
	@echo "✓ Linux ARM64"
	
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w -X main.Version=$(VERSION)" \
		-o $(DIST_DIR)/$(BINARY)-darwin-amd64 cmd/detector/main.go
	@echo "✓ macOS AMD64"
	
	# macOS ARM64 (M1/M2)
	GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w -X main.Version=$(VERSION)" \
		-o $(DIST_DIR)/$(BINARY)-darwin-arm64 cmd/detector/main.go
	@echo "✓ macOS ARM64"
	
	# Windows AMD64
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -X main.Version=$(VERSION)" \
		-o $(DIST_DIR)/$(BINARY)-windows-amd64.exe cmd/detector/main.go
	@echo "✓ Windows AMD64"
	
	@echo "All builds complete in $(DIST_DIR)/"

# 创建发布包
release: build-all
	@echo "Creating release packages..."
	@mkdir -p $(DIST_DIR)/packages
	
	# 复制指纹数据
	cp -r pkg/fingerprints/data $(DIST_DIR)/fingerprints
	
	# 创建各平台压缩包
	cd $(DIST_DIR) && \
	for file in $(BINARY)-linux-* $(BINARY)-darwin-* $(BINARY)-windows-*; do \
		if [ -f "$$file" ]; then \
			mkdir -p "tmp_$$file" && \
			if echo "$$file" | grep -q "windows"; then \
				cp "$$file" "tmp_$$file/$(BINARY).exe"; \
			else \
				cp "$$file" "tmp_$$file/$(BINARY)"; \
			fi && \
			cp -r fingerprints "tmp_$$file/" && \
			cp ../../README.md "tmp_$$file/" 2>/dev/null || true && \
			if echo "$$file" | grep -q "windows"; then \
				cd "tmp_$$file" && zip -r "../packages/$$file.zip" . && cd ../..; \
			else \
				tar -czf "packages/$$file.tar.gz" -C "tmp_$$file" .; \
			fi && \
			rm -rf "tmp_$$file"; \
		fi \
	done
	
	@echo "✓ Release packages created in $(DIST_DIR)/packages/"

# Docker构建
docker:
	docker build -t $(BINARY):$(VERSION) -t $(BINARY):latest .

# 安装到系统
install: build
	@echo "Installing $(BINARY) to /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(BINARY) /usr/local/bin/
	@sudo mkdir -p /etc/$(BINARY)
	@sudo cp -r pkg/fingerprints/data /etc/$(BINARY)/fingerprints
	@echo "✓ Installed to /usr/local/bin/$(BINARY)"

# 卸载
uninstall:
	@echo "Uninstalling $(BINARY)..."
	@sudo rm -f /usr/local/bin/$(BINARY)
	@sudo rm -rf /etc/$(BINARY)
	@echo "✓ Uninstalled"

# 测试
test:
	go test -v ./...

# 清理
clean:
	@rm -rf $(BUILD_DIR) $(DIST_DIR)
	@echo "✓ Cleaned build artifacts"

# 运行本地测试
run: build
	$(BUILD_DIR)/$(BINARY) -t localhost:11434 -v

# 帮助
help:
	@echo "LLM Detector Build System"
	@echo ""
	@echo "Usage:"
	@echo "  make build       - Build for current platform"
	@echo "  make build-all   - Build for all platforms"
	@echo "  make release     - Create release packages"
	@echo "  make install     - Install to /usr/local/bin"
	@echo "  make docker      - Build Docker image"
	@echo "  make test        - Run tests"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make help        - Show this help"
