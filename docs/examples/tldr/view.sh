#!/bin/bash

sunbeam query .command | xargs tldr --raw | sunbeam detail
