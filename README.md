# Introduction

## What is sunbeam?

Sunbeam is a command-line launcher, inspired by [fzf](https://github.com/junegunn/fzf), [raycast](https://raycast.com) and [gum](https://github.com/charmbracelet/gum).

It allows you to build interactives UIs from simple scripts.

![demo gif](./docs/examples/demo/demo.gif)

## Why sunbeam?

I love TUIs, but I spent way to much time writing them.

I used a lot of application launchers, but all of them had some limitations.

Sunbeam supports:

- Running on all platforms (Windows, Linux, MacOS)
- Commands written in any language, as long as they can output JSON
- Generating powerful UIs composed of a succession of pages

## Running the examples in this repository

- clone this repository: `git clone https://github.com/pomdtr/sunbeam.git && cd sunbeam`
- run sunbeam from the root of the repository: `go install && sunbeam`
