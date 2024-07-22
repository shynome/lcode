testserver:
	go run -ldflags="-X 'main.VERSION=test'" .
build:
	go build -ldflags="-X 'main.VERSION=$$(git describe --tags --always --dirty | cut -c2-)' -s -w" -o lcode .
build-with-upx: build
	upx lcode
