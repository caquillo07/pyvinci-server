BINARY=pyvinci
GOARCH=amd64

dev-reload:
	air -c .air.conf

dev:
	go run . server --config config.yaml --dev-log

migrate:
	go run . migrate --config config.yaml --dev-log

linux:
	GOOS=linux GOARCH=${GOARCH} go build -o ${BINARY}-linux-${GOARCH} .

darwin:
	GOOS=darwin GOARCH=${GOARCH} go build -o ${BINARY}-darwin-${GOARCH} .