# Logger

## Synopsis

Logger is listening for all messages on nats, it will encode the sensible data for each message and will send it to the created logger.

The default logger is a basic logger, which is mainly sending data to a log file. Logger will create this default basic logger based on the environment variable *ERNEST_LOG_FILE*, in case it's not defined it wont create a default listener.

You can create / remove new loggers by sending nats requests, for example:
```
# New basic logger
$ nats-pub logger.set `{"type":"basic","logfile":"/tmp/ernest.log"}`

# Ovrride basic logger
$ nats-pub logger.set `{"type":"basic","logfile":"/tmp/ernest-2.log"}`

# Delete basic logger
$ nats-pub logger.del `{"type":"basic"}`
```

```
# New logstash logger
$ nats-pub logger.set `{"type":"logstash","hostname":"http://my-new-logstash.com/","port":2234,"timeout":1}`

# Ovrride logstash logger
$ nats-pub logger.set `{"type":"logstash","hostname":"http://my-logstash.com/","port":2234,"timeout":1}`

# Delete logstash logger
$ nats-pub logger.del `{"type":"logstash"}`
```

Additionally an endpoint is exposed in order to query the active loggers


## Build status

* Master: [![CircleCI](https://circleci.com/gh/ernestio/logger/tree/master.svg?style=svg)](https://circleci.com/gh/ernestio/logger/tree/master)
* Develop: [![CircleCI](https://circleci.com/gh/ernestio/logger/tree/develop.svg?style=svg)](https://circleci.com/gh/ernestio/logger/tree/develop)


## Installing

```
$ make deps
$ make install
```

## Tests

Running the tests:
```
make test
```

## Contributing

Please read through our
[contributing guidelines](CONTRIBUTING.md).
Included are directions for opening issues, coding standards, and notes on
development.

Moreover, if your pull request contains patches or features, you must include
relevant unit tests.

## Versioning

For transparency into our release cycle and in striving to maintain backward
compatibility, this project is maintained under [the Semantic Versioning guidelines](http://semver.org/).

## Copyright and License

Code and documentation copyright since 2015 r3labs.io authors.

Code released under
[the Mozilla Public License Version 2.0](LICENSE).
