# nintendo-switch-checker :heavy_check_mark:

Can't you buy Nintendo Switch because the stocks always sold out soon?
nintendo-switch-checker can help you!

nintendo-switch-checker crawls bunch of online shopping sites and notify you when it's available.

### Installation

```
$ go get -u github.com/haya14busa/nintendo-switch-checker/cmd/nintendo-switch-checker
```

### Usage
You can use slack or LINE for notification service.

See `$ nintendo-switch-checker -h` for detail.

## Run on Google App Engine

1. Edit server/app.yaml, Set SLACK_API_TOKEN or LINE_NOTIFY_TOKEN
2. Deploy!

```bash
$ goapp deploy
or
$ gcloud app deploy server/app.yaml server/cron.yaml
```

fyi: https://cloud.google.com/appengine/docs/standard/go/quickstart


### :tada: :sparkles: Achivements :sparkles: :tada:

![image](https://user-images.githubusercontent.com/3797062/27329079-3bc59072-55ef-11e7-990e-fe2c77a22ce7.png)
https://twitter.com/haya14busa/status/876315003074224129

## :bird: Author
haya14busa (https://github.com/haya14busa)
