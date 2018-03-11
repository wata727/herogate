# Destroy the app

```
$ herogate destroy
▸    WARNING: This will delete ⬢ young-eyrie-24091
▸    To proceed, type young-eyrie-24091 or re-run this command with
▸    --confirm young-eyrie-24091

> young-eyrie-24091
Destroying... done
```

```
$ herogate destory --confirm young-eyrie-24091
Destroying... done
```

Also, you can specify app with `-app` options.

```
$ herogate destroy -a young-eyrie-24091
```

## Internal

The `herogate destroy` command maps to the DeleteStack API in CloudFormation. However, in order to delete S3 bucket and ECR repository, it executes DeleteBucket and DeleteRepository API before that.
