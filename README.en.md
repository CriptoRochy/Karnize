<p align="center"><img src="docs/img/icon.png" width="112" alt="Karnize"></p>

<h1 align="center">Karnize</h1>

<p align="center"><b>A kill switch for apps.</b> When your VPN drops or your network or IP changes, Karnize freezes the apps you chose at once, and closes them if an allowed address doesn't come back.</p>

<p align="center"><a href="README.md">Русский</a></p>

> Karnize is a free fork of [Kordon 1.2.2](https://github.com/ef1m03/kordon/tree/v1.2.2) (© Efimov Dev), licensed under GPL-3.0. It is not Kordon and not an official Kordon build. What changed: [CHANGELOG](CHANGELOG.md).

## How it works

- Every half second, and right away on OS network events, Karnize checks the network. On any change the apps are frozen (Windows: `NtSuspendProcess`, macOS/Linux: `SIGSTOP`) and continue exactly where they were once an allowed IP is confirmed.
- The public IP is asked from several services at once and trusted only when they agree. IPv6, the country, the DNS resolver and trusted networks (by the router's MAC) are checked too.
- **Route to the services themselves.** Claude, ChatGPT, Codex and Perplexity are judged by the address their own servers see, which is what matters with split routing.
- If no allowed IP comes back within 20 seconds, the apps are closed, politely first. Starting one meanwhile closes it right away.
- A supervisor process restarts Karnize if it crashes; on exit everything is un-frozen.

## Install

Windows (PowerShell, no admin rights):

```powershell
irm https://raw.githubusercontent.com/CriptoRochy/Karnize/main/scripts/install.ps1 | iex
```

macOS and Linux:

```bash
curl -fsSL https://raw.githubusercontent.com/CriptoRochy/Karnize/main/scripts/install.sh | sh
```

The installers take the files from the [latest release](https://github.com/CriptoRochy/Karnize/releases/latest) and check their SHA256.

## Settings

`config.yaml` in `%APPDATA%\Karnize`, `~/Library/Application Support/Karnize` or `~/.config/karnize`. It is commented and reloaded on the fly. New in Karnize: `route_check` and a per-app `route:` for custom apps.

## Build from source

Go 1.26+:

```bash
git clone https://github.com/CriptoRochy/Karnize && cd Karnize
go test ./...
CGO_ENABLED=0 GOOS=windows go build -trimpath -ldflags "-H windowsgui" -o karnize.exe ./cmd/karnize
```

## License

[GPL-3.0](LICENSE). Based on Kordon 1.2.2 © 2026 Efimov Dev, see [NOTICE](NOTICE).
