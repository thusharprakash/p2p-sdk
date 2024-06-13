function uninstall {
  for device in $(adb devices | grep -v List | awk '{print $1}')
  do
    echo "Uninstalling from device $device"
    adb -s $device uninstall com.mobileapp || true
  done
}

if [[ $1 == "fresh" ]]; then
  uninstall
fi

rm -rf p2p/out/android
rm -rf ./MobileApp/android/app/libs/p2p.aar
go get golang.org/x/mobile/cmd/gomobile
go mod download
go run golang.org/x/mobile/cmd/gomobile init
GO111MODULE=on
cd p2p/
mkdir -p out/android
go run golang.org/x/mobile/cmd/gomobile bind -v -target=android -o out/android/p2p.aar  -androidapi 24
mkdir -p ../MobileApp/android/app/libs
cp out/android/p2p.aar ../MobileApp/android/app/libs/
cd ..
cd MobileApp/android
yarn android