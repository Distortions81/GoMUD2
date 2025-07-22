# Changelog

### April 25 – May 2, 2024 (129 commits)
- Threaded networking and input buffering with flood protection.
- ANSI color with styles and Unicode support.
- Telnet command parsing and MUD client detection.
- Character map negotiation and translation.
- Account and character system with:
  - Unicode support in names and passphrases.
  - Secure password hashing and passphrase scoring.
  - Fast relog and reconnect to existing characters.
  - Unique fingerprint IDs for account/character.
- TLS/SSL secure connection support.
- Basic commands: `say`, `who`, `quit`.

### May 2 – 4, 2024 (35 commits)
- Idle kick at login menu and in-game.
- Rewrote ANSI color system.
- Basic help file system (`help ansi`).
- Improved reconnect reliability, added reconnect note to `who`.
- Fixed race condition on connection loss.
- Password hashing now asynchronous (separate thread with queue) to avoid hitching during login or account creation.

### May 5 – 6, 2024 (30 commits)
- Randomized connection processing order.
- Improved partial command matching.
- Added `look`, `go <exit>` commands.
- Abbreviated time formatting (e.g. `1h23m3s`).
- `who` now shows connected and idle time; added uptime display.
- Added `telnet options` command.
- Command: `help commands`.
- Say character limit.
- AFK kick logic varies by context (login, menu, in-game).

### May 7 – 11, 2024
- Introduced area/room structs and basic world system.
- Added `pset`, player levels, and room transfer logic.
- Commands: `dig`, `asave`, directional aliases (`nw`, `ne`, `sw`, `se`), `nsew`, and `go`.
- Async password verification.
- Player movement messages (`<name> arrives from <direction>`).
- Started OLC (Online Creation) system.
- Output buffering: character map, ANSI applied at pulse.
- Connection flood protection.
- Code cleanup.

### May 12 – 15, 2024
- ANSI size fixes and multi-threaded color/char translation.
- Began chat channels and world objects system.
- Improved `go` command and added `up`/`down` aliases.
- Commands sorted by level and name.
- Login warnings after repeated failures.
- HTTP detection and blocking with redirect.
- Fixed `quit` bug; adjusted idle timers.

### May 17 – 18, 2024
- Implemented chat channels and `tell` command.
- Improved reconnect: `who`, `look`, and chat restore after relog.
- Character autosave and fully async save system (`asave`).
- Introduced `pload` command and offline tells.
- TLS/SSL fixes.

### May 19 – 21, 2024
- Implemented `disable` command.
- Improved `who` display and command help.
- Added telnet command settings.
- `characters` command to list saved characters.
- Reconnection messages if a player hasn’t logged in for 30+ minutes.
- Added `emote` command.
- Custom UUID marshaling and room hash map.

### May 24 – 27, 2024
- Blocked host list: add/delete/clear support.
- `critlog` now only sends to imps.
- Increased password hash complexity, simplified password requirements.
- Streamlined account/character creation (`new` no longer required).
- Telnet `\r` support; SSL port shown in greeting.
- Near-instant command responses.
- Emoji shortcuts (e.g. `:smile:`), FIGlet fonts, `boom` and `del` commands.
- Improved `coninfo`, emoji translation refactored.
- Config option: `TextEmoji`.
- New `bug` command for reporting bugs/typos.

### May 30 – June 2, 2024
- NAWS (terminal size detection) support.
- `stats` command: loop time/load monitoring.
- Ban/unban commands.
- Options menu: change password, reroll.
- Admin commands: `wizlock`, `newlock`, `force <target/all>`.
- OLC improvements; `nocolor` option.

### June 3 – 5, 2024
- Added `panic` command and recovery system (logs, error dumps).
- Pulse system tuning; improved `stat` command.
- Better input spam protection.
- OLC edit history tracking.
- Config system now supports integer values.
- Manual terminal width setting.
- Added `shutdown` command.
- Improved handling of lost/recovered player links.
- Added hybrid OLC mode.
- Mono terminal type support.
- Fixed bug with player removal after AFK + disconnect.

### June 6 – 9, 2024
- Full 256-color ANSI support.
- Standardized line endings (`NEWLINE` used throughout code).
- Created manual 256-to-16 color fallback.
- Scrapped copyover prototype.
- MTTS (MUD Terminal Type Standard) support.
- Improved telnet and channel commands.
- Automatic channel enable and `nochannel` disable toggling.
- Refactored bitmask logic to use functions (allows smaller config values).
- Staff `who` now shows hidden players (`(hidden)`).
- Added `mud-stats.json`:
  - Tracks record online players.
  - Tracks total number of logins.
