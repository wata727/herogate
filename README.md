# Herogate

[![GitHub release](https://img.shields.io/github/release/wata727/herogate.svg)](https://github.com/wata727/herogate/releases/latest)
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENSE)

Heroku + AWS Fargate = Herogate 🚀 Deploy and manage containerized applications like Heroku on AWS.

## Overview

Herogate wraps management services on AWS and provides a Heroku like interface. All updates are done via CloudFormation, including targeting CodePipleline, CodeBuild, AWS Fargate, CodeCommit etc.

<p align="center">
  <img src="https://user-images.githubusercontent.com/9624059/37250952-dd506590-254a-11e8-92dd-552705ff4ab7.png" />
</p>

When pushing the new source code to CodeCommit, the Pipeline executes and a new image is built on CodeBuild. Finally, CloudFormation updates Fargate's service with the created image.

For details, you can see internal section in the [documentation](docs).

## Installation

Currently, you need to build from the source code when installing.

```
$ go get -d github.com/wata727/herogate
$ cd $GOPATH/src/github.com/wata727/herogate
$ make install
```

## Production Ready?

No. This is a highly experimental project. It should not be used in a production environment.

Currently, we don't provide a migration path from the old version. This means that you cannot bump up version without downtime.

## Quick Start

### 1. Create an app

You can create an app on AWS by the `create` command:

```
$ herogate create your-first-app
Creating app... 0%
```

This process takes about 5 minutes. After that, the remote repository is automatically added locally as `herogate`.

```
$ git remote -v
herogate ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/your-first-app (fetch)
herogate ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/your-first-app (push)
```

### 2. Open the app

You can see the created app via browser.

```
$ herogate open
```

Congratulations! 🎉 Your first app is now available.

### 3. Create your `Procfile`

You can run arbitrary containers by creating [Procfile](https://devcenter.heroku.com/articles/procfile).

```
$ cat Procfile
web: bundle exec rails server
worker: bundle exec rake jobs:work
```

### 4. Deploy new app

You can easily deploy new app with `git push`.

```
$ git push herogate master
```

Deployment logs can be seen with `herogate logs`. Unlike Heroku, `git push` is completed soon.

```
$ herogate logs
```

## Usage

Please check the [documentation](docs) for details.

## Developing

This project requires Go 1.9 or higher. You can build and install with `make install`.

```
$ make install
```

## Author

[Kazuma Watanabe](https://github.com/wata727)
