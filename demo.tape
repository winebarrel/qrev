# https://github.com/charmbracelet/vhs
Output demo.gif

Require echo

Set Shell "bash"
Set FontSize 14
Set Width 700
Set Height 500
Set Padding 20

Type "echo 'SELECT 1' > 001.sql" Sleep 100ms Enter
Type "echo 'SELECT now()' > 002.sql" Sleep 100ms Enter
Type "echo 'SELECT CURRENT_DATE' > 003.sql" Sleep 100ms Enter
Type "export QREV_DSN='file:test.db'" Sleep 100ms Enter Enter

Sleep 2s

Type "qrev init" Sleep 500ms Enter Enter

Sleep 4s

Type "qrev status" Sleep 500ms Enter Enter

Sleep 4s

Type "qrev plan" Sleep 500ms Enter Enter

Sleep 4s

Type "qrev apply" Sleep 500ms Enter Enter

Sleep 4s

Type "qrev status --show-error" Sleep 500ms Enter Enter

Sleep 4s

Type "qrev plan --if-modified" Sleep 500ms Enter Enter

Sleep 4s

Type "echo 'SELECT CURRENT_TIMESTAMP' > 002.sql" Sleep 100ms Enter Enter

Sleep 4s

Type "qrev plan --if-modified" Sleep 500ms Enter Enter

Sleep 4s

Type "qrev apply --if-modified" Sleep 500ms Enter Enter

Sleep 4s

Type "qrev apply --if-modified" Sleep 500ms Enter Enter

Sleep 4s

Type "qrev status" Sleep 500ms Enter Enter

Sleep 10s
