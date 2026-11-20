#!/bin/bash
# Builds dist/Karnize.app (Apple Silicon + Intel) and dist/Karnize-macos.zip.
# Signed ad-hoc: free, and enough for install.sh, which downloads without the quarantine flag.
set -euo pipefail
cd "$(dirname "$0")/.."

VERSION="${VERSION:-$(git describe --tags --always 2>/dev/null || echo dev)}"
VERSION="${VERSION#v}"
APP=dist/Karnize.app
rm -rf "$APP" dist/Karnize-macos.zip build/mac
mkdir -p "$APP/Contents/MacOS" "$APP/Contents/Resources" build/mac

for arch in arm64 amd64; do
  carch=$([ $arch = amd64 ] && echo x86_64 || echo arm64)
  # The panel window (WebKit) has C++ parts, so both compilers get the target arch.
  CC="clang -arch $carch" CXX="clang++ -arch $carch" CGO_CXXFLAGS="-Wno-deprecated-literal-operator" \
  CGO_ENABLED=1 GOOS=darwin GOARCH=$arch MACOSX_DEPLOYMENT_TARGET=11.0 \
    go build -trimpath -ldflags "-s -w -X main.version=$VERSION" -o build/mac/karnize-$arch ./cmd/karnize
done
lipo -create build/mac/karnize-arm64 build/mac/karnize-amd64 -output "$APP/Contents/MacOS/karnize"

go run ./tools/mkicons build/icons
iconutil -c icns build/icons/Karnize.iconset -o "$APP/Contents/Resources/Karnize.icns"

cat > "$APP/Contents/Info.plist" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleName</key><string>Karnize</string>
  <key>CFBundleDisplayName</key><string>Karnize</string>
  <key>CFBundleIdentifier</key><string>io.github.criptorochy.karnize</string>
  <key>CFBundleExecutable</key><string>karnize</string>
  <key>CFBundleIconFile</key><string>Karnize</string>
  <key>CFBundlePackageType</key><string>APPL</string>
  <key>CFBundleShortVersionString</key><string>$VERSION</string>
  <key>CFBundleVersion</key><string>$VERSION</string>
  <key>LSMinimumSystemVersion</key><string>11.0</string>
  <key>LSUIElement</key><true/>
  <key>NSHumanReadableCopyright</key><string>© 2026 Efimov Dev (Kordon), CriptoRochy (Karnize). GPL-3.0</string>
</dict>
</plist>
EOF

codesign --force --deep --sign - "$APP"
(cd dist && ditto -c -k --keepParent Karnize.app Karnize-macos.zip)
echo "built $APP ($VERSION)"
