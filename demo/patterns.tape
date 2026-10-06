Output patterns.gif

Set FontSize 16
Set Width 1000
Set Height 580
Set Padding 20
Set TypingSpeed 40ms
Set Theme '{"background": "#1e1e2e", "foreground": "#cdd6f4", "cursor": "#f5e0dc", "palette": ["#45475a", "#f38ba8", "#a6e3a1", "#f9e2af", "#89b4fa", "#f5c2e7", "#94e2d5", "#bac2de", "#585b70", "#f38ba8", "#a6e3a1", "#f9e2af", "#89b4fa", "#f5c2e7", "#94e2d5", "#a6adc8"]}'

# Show a package.json with several dependencies
Type "cat fixture-patterns.json"
Enter
Sleep 800ms

# Only upgrade packages matching a pattern
Type "upd -n -f fixture-patterns.json 'react*'"
Enter
Sleep 5s
