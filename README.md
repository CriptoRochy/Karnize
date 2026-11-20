<p align="center"><img src="docs/img/icon.png" width="112" alt="Karnize"></p>

<h1 align="center">Karnize</h1>

<p align="center"><b>Kill switch для программ.</b> Упал VPN, сменилась сеть или IP - Karnize мгновенно ставит выбранные программы на паузу, а если разрешённый адрес не вернулся, закрывает их.</p>

<p align="center"><a href="README.en.md">English</a></p>

> Karnize - свободный форк [Kordon 1.2.2](https://github.com/ef1m03/kordon/tree/v1.2.2) (© Efimov Dev) под лицензией GPL-3.0. Это не Kordon и не его официальная сборка. Что изменено - в [CHANGELOG](CHANGELOG.md).

## Как это работает

```
Старт ──первая проверка──▶ Можно ◀──────── разрешённый IP ────────┐
                             │                                    │
       сменилась сеть / чужой IP / нет сети / утечка IPv6 или DNS │
                             ▼                                    │
                   Пауза (программы заморожены) ── 20 с ──▶ Закрыто
```

- Каждые полсекунды и сразу по событию системы Karnize сверяет сеть. При любом изменении программы замораживаются (Windows: `NtSuspendProcess`, macOS/Linux: `SIGSTOP`) и продолжают ровно с того же места, как только подтвердится разрешённый IP.
- Внешний IP узнаётся у нескольких сервисов сразу, верить можно, только если они согласны. Отдельно проверяются IPv6, страна, DNS-резолвер и доверенные сети (по MAC роутера).
- **Маршрут до самих сервисов.** Claude, ChatGPT, Codex и Perplexity проверяются по адресу, который видят их собственные серверы. Это важно при раздельной маршрутизации, когда часть трафика идёт мимо VPN.
- Если разрешённый IP не вернулся за 20 секунд, программы закрываются: сначала вежливо, потом принудительно. Новый запуск в это время сразу закрывается.
- Отдельный процесс-сторож перезапускает Karnize при падении. При выходе всё размораживается.

Под защиту можно взять: claude, chatgpt, codex, cursor, windsurf, copilot, gemini, perplexity, telegram, discord, slack, zoom и любые свои программы.

## Установка

Windows (PowerShell, без прав администратора):

```powershell
irm https://raw.githubusercontent.com/CriptoRochy/Karnize/main/scripts/install.ps1 | iex
```

macOS и Linux:

```bash
curl -fsSL https://raw.githubusercontent.com/CriptoRochy/Karnize/main/scripts/install.sh | sh
```

Установщики берут файлы из [последнего релиза](https://github.com/CriptoRochy/Karnize/releases/latest) и сверяют SHA256. Сборки не подписаны платными сертификатами, поэтому SmartScreen или Gatekeeper могут предупредить.

## Настройки

`%APPDATA%\Karnize\config.yaml` (Windows), `~/Library/Application Support/Karnize/config.yaml` (macOS), `~/.config/karnize/config.yaml` (Linux). Файл с комментариями, изменения подхватываются на лету.

```yaml
apps: [claude, cursor]
custom_apps:
  - name: My App
    match: ["myapp.exe"]
    route: ["https://my-service.example/cdn-cgi/trace"]  # необязательно
allow:
  ips: ["64.137.54.0/24"]
  countries: [NL, DE]
app_rules:
  cursor:
    countries: [NL, DE, US]
grace: 20s
route_check: true
dns_check: warn   # warn | block | off
```

Команды:

```
karnize                 запустить (значок у часов / в строке меню)
karnize status          IP, страна, маршруты до сервисов и состояние
karnize allow [IP|CC]   разрешить текущий IP, адрес или подсеть, страну
karnize config          путь к config.yaml
karnize autostart on|off
karnize panel           открыть панель
karnize headless        без значка (серверы)
karnize update          обновиться до последнего релиза
```

## Сборка из исходников

Нужен Go 1.26+.

```bash
git clone https://github.com/CriptoRochy/Karnize && cd Karnize
go test ./...
CGO_ENABLED=0 GOOS=windows go build -trimpath -ldflags "-H windowsgui" -o karnize.exe ./cmd/karnize
./scripts/build-mac.sh        # macOS, на Mac
```

## Лицензия

[GPL-3.0](LICENSE). Основано на Kordon 1.2.2 © 2026 Efimov Dev, см. [NOTICE](NOTICE).
