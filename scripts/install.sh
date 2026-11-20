#!/bin/sh
# Karnize installer for macOS and Linux.
#
#   curl -fsSL https://raw.githubusercontent.com/CriptoRochy/Karnize/main/scripts/install.sh | sh
#   curl -fsSL https://raw.githubusercontent.com/CriptoRochy/Karnize/main/scripts/install.sh | sh -s uninstall
#
# On macOS the app is downloaded with curl, so it carries no quarantine flag and
# Gatekeeper doesn't block it although it isn't notarized (that costs $99/year).
set -eu

REPO="CriptoRochy/Karnize"
BASE="https://github.com/$REPO/releases/latest/download"
ACTION="${1:-install}"

say() { printf '\033[1m%s\033[0m\n' "$*"; }

# verify FILE NAME: compares FILE with the NAME line of the release's SHA256SUMS.txt,
# so a file damaged or swapped on the way is never installed.
verify() {
  curl -fsSL "$BASE/SHA256SUMS.txt" -o "$1.sums"
  want="$(grep "  $2\$" "$1.sums" | cut -d' ' -f1)"
  if command -v sha256sum >/dev/null 2>&1; then got="$(sha256sum "$1" | cut -d' ' -f1)"
  else got="$(shasum -a 256 "$1" | cut -d' ' -f1)"; fi
  if [ -z "$want" ] || [ "$want" != "$got" ]; then
    echo "Контрольная сумма $2 не совпала - файл повреждён или подменён, ничего не установлено." >&2
    exit 1
  fi
}

stop_running() {
  # Quitting through the menu would un-freeze apps; on update Karnize does it at its next start.
  pkill -x karnize 2>/dev/null || true
  sleep 1
}

mac() {
  dest="/Applications"
  [ -w "$dest" ] || dest="$HOME/Applications"
  if [ "$ACTION" = "uninstall" ]; then
    stop_running
    rm -rf "$dest/Karnize.app" "$HOME/Applications/Karnize.app"
    rm -f "$HOME/Library/LaunchAgents/io.github.criptorochy.karnize.plist"
    say "Karnize удалён. Настройки остались в ~/Library/Application Support/Karnize - удалите папку, если не нужны."
    return
  fi
  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' EXIT
  say "Скачиваю Karnize…"
  curl -fsSL "$BASE/Karnize-macos.zip" -o "$tmp/Karnize.zip"
  verify "$tmp/Karnize.zip" Karnize-macos.zip
  stop_running
  mkdir -p "$dest"
  rm -rf "$dest/Karnize.app"
  ditto -x -k "$tmp/Karnize.zip" "$dest"
  xattr -dr com.apple.quarantine "$dest/Karnize.app" 2>/dev/null || true
  open "$dest/Karnize.app"
  say "Готово: Karnize в $dest и уже в строке меню (зелёная точка рядом с часами)."
}

linux() {
  bin="$HOME/.local/bin"
  if [ "$ACTION" = "uninstall" ]; then
    stop_running
    rm -f "$bin/karnize" "$HOME/.config/autostart/karnize.desktop"
    say "Karnize удалён. Настройки остались в ~/.config/karnize."
    return
  fi
  case "$(uname -m)" in
    x86_64 | amd64) arch=amd64 ;;
    aarch64 | arm64) arch=arm64 ;;
    *) echo "Неподдерживаемая архитектура: $(uname -m)" >&2; exit 1 ;;
  esac
  mkdir -p "$bin"
  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' EXIT
  say "Скачиваю Karnize…"
  curl -fsSL "$BASE/karnize-linux-$arch.tar.gz" -o "$tmp/karnize.tar.gz"
  verify "$tmp/karnize.tar.gz" "karnize-linux-$arch.tar.gz"
  stop_running
  tar -xzf "$tmp/karnize.tar.gz" -C "$bin" karnize
  chmod +x "$bin/karnize"
  "$bin/karnize" autostart on
  nohup "$bin/karnize" >/dev/null 2>&1 &
  say "Готово: $bin/karnize запущен и стартует при входе."
}

case "$(uname -s)" in
  Darwin) mac ;;
  Linux) linux ;;
  *) echo "Для Windows (PowerShell): irm https://raw.githubusercontent.com/$REPO/main/scripts/install.ps1 | iex" >&2; exit 1 ;;
esac
