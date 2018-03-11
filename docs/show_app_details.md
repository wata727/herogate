# Show app details

```
$ herogate info
=== young-eyrie-24091
Containers:       web: 1
                  worker: 1
Web URL:          http://young-eyrie-24091-123456789.us-east-1.elb.amazonaws.com
Git URL:          ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/young-eyrie-24091
Status:           CREATE_COMPLETE
Region:           us-east-1
Platform Version: 1.0
```

Also, you can specify app with `-app` options.

```
$ herogate info -a young-eyrie-24091
```

## Internal

The `herogate info` command maps to the DescribeStacks API in CloudFormation. The Git URL and the Web URL are set as output of the stack. Container definition is obtained from the latest task definition.
