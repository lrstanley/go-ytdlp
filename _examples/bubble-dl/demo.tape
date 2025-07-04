# go run github.com/charmbracelet/vhs@latest demo.tape
Output demo.gif
Require go

Set Theme "Dracula+"
Set FontSize 18
Set Width 1600
Set Height 450
Set Margin 20
Set MarginFill "#674EFF"
Set BorderRadius 10
Env PS1 "$ "

Type "go run *.go"
Sleep 500ms
Enter
Sleep 18s
