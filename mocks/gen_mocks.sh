rm -rf mocks/mock_*.go

bin/mockgen -source globalping/client.go -destination mocks/mock_client.go -package mocks
bin/mockgen -source globalping/probe/probe.go -destination mocks/mock_probe.go -package mocks
bin/mockgen -source view/viewer.go -destination mocks/mock_viewer.go -package mocks
bin/mockgen -source utils/utils.go -destination mocks/mock_utils.go -package mocks
