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

`-w` option enables highlighting based on log level. Default regex to detect log level is `(?i)\\b((?P<warning>warn|warning)|(?P<error>error))\\b`. Custom regex can be provided with `-l` option. 

Custom regex must contain named capture groups with names `warning` and `error`. If any of these groups are not empty for any given line of code then that line will be highlighted with yellow for warning or red for errors. If both groups match, then line will be highlighted as error.

Your terminal must support 24-bit colors for log level highlighting to work. 

### Highlighting parts of the log message

`-c` option allows to pass a regular expression to highlight certain parts of the log messages. If expression matches the log line, then each capturing group of the expression will be displayed in a distinct color.

Example of the regexp to highlight time, then something in square brackets then sequence of non-whitespace characters and then sequence of alphanumeric characters and '.'

    `(\d{2}:\d{2}:\d{2}.\d{3})\s+\[(.*)\]\s+(\S+)\s+([a-zA-Z0-9_.]+).*`

### Filtering the log

`-f` option allows to pass regular expression to filter log lines. If the log line matches the expression then it will be displayed and the matching part will be highlighted.

If the expression is prefixed with exclamation mark that it will work as `grep -v`, i.e. all the lines that match the expression **will be discarded**.

`-f Exception` will match all the lines containing the sequence "Exception"

`-f !Exception` will match all the lines that do not contain the sequence "Exception"

