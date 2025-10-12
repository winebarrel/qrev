# qrev

[![test](https://github.com/winebarrel/qrev/actions/workflows/test.yml/badge.svg)](https://github.com/winebarrel/qrev/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/winebarrel/qrev)](https://goreportcard.com/report/github.com/winebarrel/qrev)

qrev is a SQL execution history management tool.

![](https://github.com/user-attachments/assets/288d0ef9-a0cf-437a-95c2-2f6c43e7c449)

## Usage

```
Usage: qrev --dsn=STRING <command> [flags]

Flags:
  -h, --help             Show context-sensitive help.
  -d, --dsn=STRING       DSN for the database to connect to ($QREV_DSN).
      --timeout=3m       Transaction timeout duration ($QREV_TIMEOUT).
      --[no-]iam-auth    Use RDS IAM authentication ($QREV_IAM_AUTH).
  -C, --[no-]color       Colorize output ($QREV_COLOR).
      --version

Commands:
  apply --dsn=STRING [<path>] [flags]
    TODO

  init --dsn=STRING [flags]
    TODO

  mark --dsn=STRING <status> <name> [flags]
    TODO

  plan --dsn=STRING [<path>] [flags]
    TODO

  status --dsn=STRING [<status-or-filename>] [flags]
    TODO

Run "qrev <command> --help" for more information on a command.
```

```
$ echo 'SELECT 1' > 001.sql
$ echo 'SELECT now()' > 002.sql
$ echo 'SELECT CURRENT_DATE' > 003.sql
$ export QREV_DSN='file:test.db'

$ qrev init
qrev_history table has been created

$ qrev status
No SQL history

$ qrev plan
001.sql SELECT 1
002.sql SELECT now()
003.sql SELECT CURRENT_DATE

$ qrev apply
done 001.sql SELECT 1
fail 002.sql SELECT now()
│ SQL logic error: no such function: now (1)
qrev: error: SQL fails

$ qrev status --show-error
12 Oct 15:40 done a0a22c9 001.sql
12 Oct 15:40 fail df4776a 002.sql
│ SQL logic error: no such function: now (1)

$ qrev plan --if-modified
003.sql SELECT CURRENT_DATE

$ echo 'SELECT CURRENT_TIMESTAMP' > 002.sql
$ qrev plan --if-modified
002.sql* SELECT CURRENT_TIMESTAMP
003.sql SELECT CURRENT_DATE

$ qrev apply --if-modified
done 002.sql SELECT CURRENT_TIMESTAMP
done 003.sql SELECT CURRENT_DATE

$ qrev apply --if-modified
No SQL file to run

$ qrev status
12 Oct 15:40 done a0a22c9 001.sql
12 Oct 15:40 done 45fb14e 002.sql
12 Oct 15:40 done e8d881b 003.sql
```
