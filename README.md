<div align="center">


# AstralRbxTracker


**Posts a Discord notification when specific Roblox games update**

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Language](https://img.shields.io/badge/Go-logoColor=white)


</div>

## Usage

Download from: [Releases](../../releases)

| OS                    | File                     | Run it                                                        |
|-----------------------|--------------------------|---------------------------------------------------------------|
| Windows               | `AstralRbxTracker.exe`   | double click, or `.\AstralRbxTracker.exe`                     |
| Linux                 | `AstralRbxTracker-linux` | `chmod +x AstralRbxTracker-linux && ./AstralRbxTracker-linux` |
| macOS (Apple Silicon) | `AstralRbxTracker-macos` | `chmod +x AstralRbxTracker-macos && ./AstralRbxTracker-macos` |

* Create a Discord webhook and copy it
* Run `AstralRbxTracker`, paste the webhook when asked
* Paste a place ID when asked, add more with `AstralRbxTracker add <place_id>`
* Keep it running

Checks **https://games.roblox.com/v1/games** every 2 minutes

### Adding games

The place ID is the number in the game's link: `roblox.com/games/`**`13822889`**`/Lumber-Tycoon-2`

```bash
AstralRbxTracker add 13822889
AstralRbxTracker list
AstralRbxTracker disable 13822889
AstralRbxTracker enable 13822889
AstralRbxTracker remove 13822889
```

Swap `AstralRbxTracker` for your file from the table above, like `./AstralRbxTracker-linux add 13822889`

`add` finds the universe ID and saves it to `games.json` for you. The name comes from Roblox

To edit `games.json` by hand instead, get the universe ID from **https://apis.roblox.com/universes/v1/places/13822889/universe** (swap in your place ID):

```json
{ "games": [ { "place_id": 13822889, "universe_id": 2471084, "enabled": true } ] }
```

### Running it on GitHub instead

* Fork this repo
* Go to **Settings > Secrets and variables > Actions**, **Repository secrets** tab (not Environment secrets), **New repository secret**: name it `DISCORD_WEBHOOK_URL`, paste your webhook
* Go to the **Actions** tab and click **I understand my workflows, go ahead and enable them**
* [check-version.yml](.github/workflows/check-version.yml)

## Build

needs [Go 1.23](https://go.dev/doc/install)+

```bash
git clone https://github.com/<your-fork>/AstralRbxTracker.git
cd AstralRbxTracker
```

set `DISCORD_WEBHOOK_URL=` in `.env`, then:

```bash
go run ./cmd/AstralRbxTracker
```

### Compiling

```bash
go build -trimpath -ldflags "-s -w" -o dist/ ./cmd/AstralRbxTracker
```

Outputs `dist/AstralRbxTracker` (`.exe` on Windows) for your OS. Pushing a tag like `v1.0.0` builds all three and publishes them automatically via [release.yml](.github/workflows/release.yml)

## License

[MIT](LICENSE) © [YourPOV](https://github.com/yourpovv)
