# Display environment variables

```
$ herogate config
=== young-eyrie-24091 Config Vars
RAILS_ENV: production
RACK_ENV:  production
```

Also, you can specify app with `-app` options.

```
$ herogate config -a young-eyrie-24091
```

If you want to get a single environment variable, use `config:get` command instead.

```
$ herogate config:get RAILS_ENV
production
```

## Internal

The `herogate config` command maps to the DescribeTaskDefinition API in ECS. Display the obtained container definition environment variable list.
