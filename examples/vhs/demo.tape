Output ../../assets/demo.gif

Set Shell zsh
Set FontSize 28
Set Width 1400
Set Height 800
Set TypingSpeed 100ms

Sleep 1s
Type "go run ./main.go examples/go-cli-cobra ./demo-out"
Enter
Sleep 2s

# Project scaffolding intro appears here

# 1) Project name (change from default)
Type "stellar-cli"
Enter
Sleep 1.0s

# 2) Go module name (change from default)
Type "github.com/yaroslavk/stellar-cli"
Enter
Sleep 1.0s

# 3) Author name (change from default)
Type "Yaroslav K"
Enter
Sleep 1.0s

# 4) Author email (change from default)
Type "yaroslav@example.com"
Enter
Sleep 1.0s

# 5) Project description (change from default)
Type "Fast CLI starter with Cobra + Viper"
Enter
Sleep 1.0s

# 6) License (change to Apache-2.0)
Down
Sleep 0.5s
Enter
Sleep 1.0s

# 7) Include configuration file support? (toggle to No)
Left
Sleep 0.5s
Enter
Sleep 1.0s

# 8) Include version command? (toggle to No)
Left
Sleep 0.5s
Enter
Sleep 1.0s

# 9) Go version (change to 1.22)
Type "1.22"
Enter

# Allow time for pre/post hooks and progress to complete
Sleep 12s

# Let the success outro display
Sleep 4s
