# nagome-polly

nagome-polly is the plugin of nagome to read aloud comments by nico nico live streaming by using [AWS polly](https://aws.amazon.com/polly/)

# Supported Platforms

nagome-polly supports Windows, Linux, and OS X. nagome-polly on Windows is using Direct Show library. On other OS is using SoX.

# Requirement

nagome-polly needs AWS account. Please sign up for AWS.

Please install SoX if you are using Linux or OS X.

# Install

## Install SoX

Skip this topic if you are using Windows

When OS X. you can install it by Homebrew

```
$ brew install sox
```

When Ubuntu 16.04. you can install it by apt

```
$ sudo apt-get install sox
```

## Build

You simply do below

```
go build
```

## Setup nagome-polly.yml

nagome-polly.yml is a setting file. nagome-polly needs access key and session key for AWS Polly.

Create these keys according to https://docs.aws.amazon.com/general/latest/gr/managing-aws-access-keys.html

And then, Write these keys in the file.

```
access_key : "<ACCESS KEY>"
secret_key : "<SECRET KEY>"
```

## Copy in plugin directory

Copy files are listed below to `<Nagome application directoy>/plugin/nagome-polly` directory

+ nagome-polly(.exe)
+ nagome-polly.yml
+ plugin.yml

