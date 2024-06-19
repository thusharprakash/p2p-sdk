rm -rf p2p/out/p2p.xcframework
go get golang.org/x/mobile/cmd/gomobile
go mod download
go run golang.org/x/mobile/cmd/gomobile init
GO111MODULE=on
cd p2p/
mkdir out 
mkdir -vp ../MobileApp/ios/MobileApp/Frameworks
go run golang.org/x/mobile/cmd/gomobile bind -v -tags=netgo -ldflags='-s -w' -target=ios -o out/p2p.xcframework 
echo -e "\033[1;32mGomobile bind completed successfully. Please copy the p2p folder to ios and clear xcode build and run again. \033[0m"
#cp -r out/p2p.xcframework ../MobileApp/ios/MobileApp/Frameworks/p2p.xcframework/
