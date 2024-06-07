rm -rf p2p/out/p2p.xcframework
go get golang.org/x/mobile/cmd/gomobile
go mod download
go run golang.org/x/mobile/cmd/gomobile init
GO111MODULE=on
cd p2p/
mkdir out
go run golang.org/x/mobile/cmd/gomobile bind -v -tags=netgo -ldflags='-s -w' -target=ios -o out/p2p.xcframework