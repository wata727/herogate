# Retrieve logs

```
$ herogate logs
```

Valid options are as follows:

|name|shorthand|description|
|:-|:-|:-|
|--app|-a|Specify an app|
|--num|-n|Number of lines to display (default: 100)|
|--ps|-p|Process to limit filter by|
|--source|-s|Log source to limit filter by|
|--tail|-t|Continually stream logs|

## Internal

The `herogate logs` command maps to the GetLogEvents and the DescribeServices API.
