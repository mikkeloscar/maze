# maze
[![Build Status](https://travis-ci.org/mikkeloscar/maze.svg?branch=master)](https://travis-ci.org/mikkeloscar/maze)

Automated build system for Archlinux packages.

## TODO

 * [ ] UI
 * [ ] Signed packages
 * [ ] License
 * [ ] Package sources
    * [x] AUR
    * [ ] Local packages
    * [ ] ABS

### Running in local shell

```
$ env $(cat env.conf | xargs) ./maze -d
```
