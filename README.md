Public test server: m45sci.xyz 7777 or 7778 (SSL/TLS secure) 

Old Changelog:
April 25th - May 2nd 2024 Changelog (129 commits):
Threaded networking, ANSI color with styles, telnet commands, mud client detection, unicode support, character map negotiation, character map translation, accounts, characters, fast-relog (logout), reconnect to 'already playing' characters, passphrase generation, passphrase scoring, threaded input buffering with flooding-kick, basic new character name checks, accounts/character names/passphrases support unicode, high security password hashing, support for TLS/SSL secure connections, account/character unique fingerprint IDs, basic say/who/quit commands.

May 2th - 4th (35 commits):
Login menu idle kick and in-game idle kick, rewrote ANSI color system, basic help file system (see help ansi), fixed some relog/reconnect bugs,  (no link) to who command to show player lost connection, fixed a minor race condition with link loss, new password hashing (encrypt) was rewritten to be asynchronous (on it's own cpu thread) and limited (queue, one at a time). This prevents a small lag hitch for the rest of the mud when
a new account is created or logs in.

May 5th - 6th (30 commits):
connection order shuffling , better partial command matching, basic help, basic look command, basic go <exit> command, changed time units to short form. 1h23m3s. Show connected and idle time in who. added uptime to who. added telnet options command. added say character limit.  help 'commands'.  afk kick, different for login menu, character menu and playing. 

May 7-11th:
Started areas/rooms structs, one default system area/room, player level, pset command, new UUID system added
(quicker, better), with unit test, player.toRoom and fromRoom, commands: dig, asave, new directrions, with colors (nw ne se sw), auto links rooms at boo, async password checking, players to look command, commands now have player levels, pset <level> command,  revDir() command for auto reversing directions (used by dig, to create exits to-from rooms), sendToRoom to send text to players in a specific room (such as say), nsew quick move aliases, player movement notices (name arrives from west). Started OLC command. Dig and pointer relink added, some code to protect the mud from connection flooding, output text is now buffered, charmap translated and has ansi color applied at the end of each pulse. Some code cleanup. Asave added. 

May 12-15th:
Fixed multiple bugs in ANSI color/style and improved end size, charmap and ANSI color translation are now multi-threaded, started work on chat channels and world objects, improved the "go" command, aliases to walk nw ne sw se. up and down, commands are now sorted by level and name, warn on login with many failed attempts, automatically detect HTTP requests send a redirect and block, fixed a bug with quit, adjusted idle times for different states. 

May 17-18th:
Chat channels, tells, who and look to reconnect,  player location is now saved/restored on save/quit, fixed TLS/SSL, channel command, toggle channels on/off, tell command., character autosave, player save fully async, asave fully asyn,  pload function and offline tells. 

May 19-21th:
Disable command implemented, improved WHO command, improved command help, telnet command settings added,  character list command, logins announced if haven't logged in/out in 30 min, emote command, custom marshaler for UUID, and the room hash map. 

May 24th - 27th:
Blocked host list and command to add, delete or clear the list. critlog only send to imps. increased password hash complexity, decreased password requirements, allow account and character creation without typing 'new', support telnet '\r' line ending, show SSL port in greeting. near-instant command responses., emoji shortcuts like :smile:. improved coninfo command, 'del' character support, 'boom' command,  FIGlet fonts, refactored emoji translation, TextEmoji config option, 'bug' command for recording bug/typo reports. 

May 30th - June 2nd:
NAWS support, for mud/terminal client window size detection, stats command, shows time/load of the mud loop, basic ban/unban commands, options menu (change pass, reroll), command and settings file for wizlock, newlock, force <target/all> command. OLC Improvements, nocolor option. 

June 3rd - 5th :
Added panic command & panic recovery, log and error dump. Adjusted pulse system. improved stat command. improved input spam protection. Added OLC edit history. Added int values to config system, added manual terminal width setting, added shutdown command, show if player link lost or regained. Added OLC hybrid mode. Support mono termtypes. Fixed bug with players not correctly being removed if disconnected and they hit afk timer. 

June 6th- 9th :
256 color support, replaced \r\n with NEWLINE in source code, hand-created a 256->16 color conversion. Wrote a prototype for copyover, but scrapped it. MTTS support, telnet command improve. Added automatic channel enable and nochannel disable when chatting. Moved all Bitmask bit-shifts to Bitmask functions for simplicity and so the flags can be small decimal numbers when used outside a bitmask. This shrinks config values in pfiles. Who now shows hidden players for staff with (hidden). Added mud-stats.json, to keep track of record number of players online and the total number of logins to the world. 
