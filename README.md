# watch-localhost

## Environment variable
| KEY  | VALUE |
| ---- | ----  |
| CHECK_URL | チェックするURL |
| CHECK_TIMEOUT | リクエストタイウアウト秒数 |
| CHECK_INTERVAL | チェックのインターバル秒数 |
| RETRY_COUNT | 最大リトライ回数|
| WAIT_AFTER_STOP | STOP_COMMAND実行後のウェイト秒数 |
| WAIT_AFTER_RESTART | START_COMMAND実行後のウェイト秒数 |
| STOP_COMMAND | 最大リトライ回数までチェックエラーの場合に実行するコマンド |
| START_COMMAND | STOP_COMMAND 実行後に実行するコマンド

### eg.
```
export CHECK_URL='http://localhost/'
export CHECK_TIMEOUT=3
export CHECK_INTERVAL=3
export RETRY_COUNT=3
export WAIT_AFTER_STOP=3
export WAIT_AFTER_RESTART=120
export STOP_COMMAND='killall -s KILL httpd'
export START_COMMAND='systemctl start httpd'
```
