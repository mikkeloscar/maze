# maze
[![Build Status](https://travis-ci.org/mikkeloscar/maze.svg?branch=master)](https://travis-ci.org/mikkeloscar/maze)

Automated build system for Archlinux packages.

## TODO

 * [ ] UI
 * [ ] Sign packages
 * [ ] Package sources
    * [x] AUR
    * [ ] Local packages

### Running in local shell

```
$ env $(cat env.conf | xargs) ./maze -d
```

## License

See [LICENSE](LICENSE).
