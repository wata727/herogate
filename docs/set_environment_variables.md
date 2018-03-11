# Set environment variables

```
$ herogate config:set RAILS_ENV=production RACK_ENV=production
Setting RAILS_ENV, RACK_ENV and restarting ⬢ young-eyrie-24091... done
```

Also, you can specify app with `-app` options.

```
$ herogate config:set -a young-eyrie-24091 RAILS_ENV=production RACK_ENV=production
```

Since the environment variable is passed to the application without being encrypted, you should not set secrets.

## Internal

The `herogate config:set` command maps to the UpdateStack API in CloudFormation. Update the task definition with the input environment variable and update the stack.
