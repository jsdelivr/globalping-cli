rm -rf mocks/mock_*.go

bin/mockgen -source api/client.go -destination mocks/api/mock_client.go -package api
bin/mockgen -source api/probe/probe.go -destination mocks/api/mock_probe.go -package api
bin/mockgen -source view/viewer.go -destination mocks/view/mock_viewer.go -package view
bin/mockgen -source utils/utils.go -destination mocks/utils/mock_utils.go -package utils
bin/mockgen -destination mocks/globalping/mock_globalping.go -package lib github.com/jsdelivr/globalping-go Client
