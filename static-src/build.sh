#!/bin/bash
time node_modules/webpack/bin/webpack.js --config=webpack.dll.js
time node_modules/webpack/bin/webpack.js --progress -p
