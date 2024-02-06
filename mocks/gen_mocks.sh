rm -rf mocks/mock_*.go

bin/mockgen -source globalping/globalping.go -destination mocks/mock_globalping.go -package mocks
