prepare:
	git submodule update --init --recursive; \
	go mod download; \
	pushd web/piflix-web; \
	git pull origin main; \
	popd

run-mac:
	go run --tags "libsqlite3 darwin" main.go

build-mac:
	tag=$$(git describe --tags --abbrev=0); \
	rev=$$(git rev-parse --short HEAD); \
	ver="$${tag}-$${rev}"; \
	go build --tags "libsqlite3 darwin" -ldflags "-X main.version=$${ver}"

build-linux:
	tag=$$(git describe --tags --abbrev=0); \
	rev=$$(git rev-parse --short HEAD); \
	ver="$${tag}-$${rev}"; \
	go build --tags="libsqlite3 linux" -ldflags "-X main.version=$${ver}"

package-mac: build-web build-mac

package-linux: build-web build-linux

build-web:
	pushd web/piflix-web; \
	npm install; \
	npm run build; \
	popd