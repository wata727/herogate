# List your containers

```
$ herogate ps
=== web (1): bundle exec rails server

=== worker (1): bundle exec rake jobs:work

```

Also, you can specify app with `-app` options.

```
$ herogate ps -a young-eyrie-24091
```

## Internal

The `herogate apps` command maps to the DescribeTaskDefinition API in ECS. Display a list of container definitions.
