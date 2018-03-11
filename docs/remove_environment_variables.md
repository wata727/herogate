# Remove environment variables

```
$ herogate config:unset RAILS_ENV RACK_ENV
Unsetting RAILS_ENV, RACK_ENV and restarting ⬢ young-eyrie-24091... done
```

Also, you can specify app with `-app` options.

```
$ herogate config:unset -a young-eyrie-24091 RAILS_ENV RACK_ENV
```

## Internal

The `herogate config:unset` command maps to the UpdateStack API in CloudFormation. Delete the input environment variable from the task definition and update the stack.
