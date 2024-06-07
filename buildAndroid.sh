rm -rf p2p/out/android
go get golang.org/x/mobile/cmd/gomobile
go mod download
go run golang.org/x/mobile/cmd/gomobile init
GO111MODULE=on
cd p2p/
mkdir out/android
go run golang.org/x/mobile/cmd/gomobile bind -v -target=android -o out/android/p2p.aar  -androidapi 24