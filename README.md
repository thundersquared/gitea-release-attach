# Gitea Release Attach

GRA was built to facilitate uploading artifacts to Gitea Releases upon CI builds.

## Installation

Download the executable binary and run it, or compile it yourself with:

```shell script
go install
go build
```

## Usage

You can simply launch the binary with `-h` arg to view required params.

Gitea username and password can be either passed as params, or set as environment variables.

Params:
```shell script
./gitea-release-attach -u username -p password ...
``` 

Environment:
```shell script
export GITEA_USER=username
export GITEA_PASS=password
./gitea-release-attach ...
``` 

These are some other params this tool supports:

| Short | Param | Required | Purpose |
| ----- | ----- | -------- | ------- |
| -u | --user | yes [1] | Set username for login |
| -p | --pass | yes [1] | Set password for login |
| -r | --repo | yes | Set target repository |
| -t | --tag | yes | Set target repository tag |
| -d | --delete | no | Clean existing attachments on Target Release |
| -f | --attachment | yes | File to be uploaded [2] |
 
[1] Required if not set as Environment Variable.
[2] Can be specified multiple times, for uploading multiple files to the same release. 
