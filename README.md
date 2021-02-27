# cwltail

`tail -f` for cloudwatch logs and with custom highlighting.

Supports tailing all streams for cloudwatch log group with following capabilities:

 - Highlight parts of log message based on regular expression
 - Highlight warning and error log messages, pattern to detect log level can be customized
 - Filter log lines by matching or not matching regular expression
 - Including short (last 6 character) name of log stream in the log message
 - Include Cloudwatch event timestamp (either time or full timestamp) in the log message
 - Use AWS profile name for credentials

## Usage

### Highlighting based on log level

`-w` option enables highlighting based on log level. Default regex to detect log level is `(?i)\s+warn|warning|error\s+`. Custom regex can be provided with `-l` option. 

If matching expression equals