# Create new app

```
$ herogate create your-first-app
Creating app... done, ⬢ your-first-app
http://your-first-app-123456789.us-east-1.elb.amazonaws.com | ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/your-first-app
```

If you omit the app's name, it generates Heroku-like memorable random name.

```
$ herogate create
Creating app... done, ⬢ young-eyrie-24091
http://young-eyrie-24091-123456789.us-east-1.elb.amazonaws.com | ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/young-eyrie-24091
```

Currently, Herogate only supports `us-east-1` region because of AWS Fargate support.

## Internal

The `herogate create` command maps to the CreateStack API in CloudFormation. This [template](../api/assets/platform.yaml) is used when creating a stack.
