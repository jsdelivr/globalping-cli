rm -rf mocks/mock_*.go

bin/mockgen -source client/measurements_fetcher.go -destination mocks/mock_measurements_fetcher.go -package mocks
