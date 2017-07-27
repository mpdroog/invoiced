// Webpack DLL to build one big dependency file
var path = require("path");
var webpack = require("webpack");

module.exports = {
    entry: {
        vendor: [
            "react", "react-dom", "react-router",
            "axios", "moment", "big.js",
            "react-datepicker"
        ]
    },
    output: {
        path: __dirname + "/../static/assets",
        filename: "[name].dll.js",
        library: "[name]"
    },
    plugins: [
        new webpack.DllPlugin({
            path: "[name].json",
            name: "[name]"
        }),
        new webpack.optimize.UglifyJsPlugin()
    ]
};