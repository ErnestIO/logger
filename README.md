# Logger

## Synopsis

This microservice will send all configured nats messages to logstash, giving additional visibility between events inside of ernest.

## Build status

* Master: [![CircleCI Master](https://circleci.com/gh/ErnestIO/logger/tree/master.svg?style=svg&circle-token=cad48d128f889cbf40f7143a5882313668989ce6)](https://circleci.com/gh/ErnestIO/logger/tree/master)
* Develop: [![CircleCI Develop](https://circleci.com/gh/ErnestIO/logger/tree/develop.svg?style=svg&circle-token=cad48d128f889cbf40f7143a5882313668989ce6)](https://circleci.com/gh/ErnestIO/logger/tree/develop)


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
