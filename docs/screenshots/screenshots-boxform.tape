# Regenerate: vhs docs/screenshots/screenshots-boxform.tape
# Companion to screenshots.tape for the ONE opt-in Box-Form screenshot
# (BT_BOXFORM=1, README §Experimental). Kept as its own tape/process rather
# than a relaunch inside screenshots.tape — a mid-tape quit+relaunch proved
# flaky (the second `bt` never came up). Same isolated $HOME + throwaway repo
# COPY setup as screenshots.tape (see its header for the setup commands).
Output docs/screenshots/_session-boxform.gif

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

Type "BT_BOXFORM=1 ./bin/bt"
Enter
Sleep 1500ms

# Move the tree cursor onto a leaf task so the Detail pane shows a fully
# populated Box-Form (Title / Status·Type·Priority / Parent·Tags / Body …).
Down 2
Sleep 400ms
Screenshot docs/screenshots/boxform.png

Type "q"
Sleep 200ms
