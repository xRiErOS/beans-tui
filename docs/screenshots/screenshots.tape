# Regenerate: vhs docs/screenshots/screenshots.tape
# Runs against an isolated $HOME + a throwaway COPY of the repo (not this
# checkout) so the Lobby/Repo-Picker never leaks real local paths or other
# registered projects into the screenshots. See git history of this file /
# the README-publish session for the setup commands (mkdir fake HOME with a
# single-repo config.yaml, cp -R the repo to /tmp, seed a few loose backlog
# beans with `beans create` in the COPY, run vhs, then rm -rf the copy).
Output docs/screenshots/_session.gif

Set Shell "zsh"
Set FontSize 15
Set Width 1280
Set Height 800
Set Theme "Catppuccin Macchiato"
Set Padding 20
Set TypingSpeed 50ms

Env HOME "/tmp/beans-tui-fakehome"

Type "cd /tmp/beans-tui-demo"
Enter
Sleep 300ms

Type "./bin/bt"
Enter
Sleep 1s
Screenshot docs/screenshots/browse.png

Type "p"
Sleep 800ms
Screenshot docs/screenshots/lobby.png

Escape
Sleep 500ms

Type "b"
Sleep 1s
Screenshot docs/screenshots/backlog.png

Type "b"
Sleep 500ms

Type "K"
Sleep 500ms
Screenshot docs/screenshots/command-center.png

Escape
Sleep 300ms

Down 2
Sleep 200ms
Tab
Sleep 300ms
Type "v"
Sleep 500ms
Screenshot docs/screenshots/detail-fullscreen.png

Type "q"
Sleep 200ms
