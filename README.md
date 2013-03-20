# Visor [![Build Status][1]][2]

Visor is a library which provides an abstraction over a global process state on top of [doozerd][3].

[1]: https://secure.travis-ci.org/soundcloud/visor.png
[2]: http://travis-ci.org/soundcloud/visor
[3]: https://github.com/ha/doozerd

## Installing

Install [Go 1][4], either [from source][5] or [with a prepackaged binary][6].
Then,

```bash
$ go get github.com/soundcloud/visor
```

[4]: http://golang.org
[5]: http://golang.org/doc/install/source
[6]: http://golang.org/doc/install

## Documentation

See [the gopkgdoc page](http://gopkgdoc.appspot.com/github.com/soundcloud/visor) for up-to-the-minute documentation and usage.

## Contributing

Pull requests are very much welcomed.  Create your pull request on a non-master branch, make sure a test or example is included that covers your change and your commits represent coherent changes that include a reason for the change.

To run the integration tests, make sure you have Doozerd reachable under the [DefaultUri][7] and run `go test`. TravisCI will also run the integration tests.

[7]: https://github.com/soundcloud/visor/blob/master/visor.go#L46

## Credits

* [Alexis Sellier][8]
* [Alexander Simmerl][9]
* [Daniel Bornkessel][10]
* [François Wurmus][11]
* [Matt T. Proud][12]
* [Tomás Senart][13]
* [Julius Volz][14]
* [Patrick Ellis][15]
* [Lars Gierth][16]
* [Tobias Schmidt][17]

[8]: https://github.com/cloudhead
[9]: https://github.com/xla
[10]: https://github.com/kesselborn
[11]: https://github.com/fronx
[12]: https://github.com/matttproud-soundcloud
[13]: https://github.com/tsenart
[14]: https://github.com/juliusv
[15]: https://github.com/pje
[16]: https://github.com/lgierth
[17]: https://github.com/grobie

## License

BSD 2-Clause, see [LICENSE][18] for more details.

[18]: https://github.com/soundcloud/cotterpin/blob/master/LICENSE
