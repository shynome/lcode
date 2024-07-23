testserver:
	go run -ldflags="-X 'main.Version=test'" .
build:
	go build -ldflags="-X 'main.Version=$$(git describe --tags --always --dirty)' -s -w" -o lcode .
build-with-upx: build
	upx lcode
