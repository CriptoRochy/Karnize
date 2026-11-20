**Установка и обновление**

Уже стоит Karnize - обновитесь из меню (пункт «Обновить до …») или командой `karnize update`. Первая установка:

macOS и Linux:
```bash
curl -fsSL https://raw.githubusercontent.com/CriptoRochy/Karnize/main/scripts/install.sh | sh
```

Windows (PowerShell):
```powershell
irm https://raw.githubusercontent.com/CriptoRochy/Karnize/main/scripts/install.ps1 | iex
```

Подробнее: [README](https://github.com/CriptoRochy/Karnize#readme). Файлы ниже - для ручной установки.

| Файл | Для чего |
|---|---|
| `Karnize-macos.zip` | macOS 11+, Apple Silicon и Intel |
| `karnize-windows-amd64.exe` | Windows 10/11 x64 |
| `karnize-windows-arm64.exe` | Windows на ARM |
| `karnize-linux-*.tar.gz` | Linux x64 / ARM64 |
| `SHA256SUMS.txt` | контрольные суммы |

Karnize - свободный форк [Kordon 1.2.2](https://github.com/ef1m03/kordon/tree/v1.2.2) (© Efimov Dev) под лицензией GPL-3.0.
