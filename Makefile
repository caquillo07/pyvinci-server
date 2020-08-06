
dev-reload:
	air -c .air.conf

dev:
	go run . server --config config.yaml --dev-log

migrate:
	go run . migrate --config config.yaml --dev-log