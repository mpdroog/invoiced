// https://github.com/webpack/react-starter
var webpack = require("webpack");
var ExtractTextPlugin = require("extract-text-webpack-plugin");
//var StaticSiteGeneratorPlugin = require('static-site-generator-webpack-plugin');

module.exports = {
	entry: "./app.jsx",
	output: {
		path: "../static/assets",
		filename: "app.js",
		libraryTarget: 'umd'
	},
	module: {
		loaders: [
			{
				test: /\.jsx?$/,
				exclude: /(node_modules|bower_components)/,
				loader: 'babel', // 'babel-loader' is also a legal name to reference
				query: {presets: ['es2015']}
			},
			{ test: /\.json$/, loader: "json-loader" },
			{ test: /\.css$/, loader: ExtractTextPlugin.extract("css-loader") },
            { test: /\.jpe?g$|\.gif$|\.png$|\.svg$|\.woff$|\.ttf$|\.wav$|\.mp3$/, loader: "file" }
		],
		/*preLoaders: [
			{
				test: /\.jsx$/,
				loader: "jsxhint-loader"
			}
		]*/
	},

	plugins: [
		new ExtractTextPlugin("app.css", { allChunks: true }),

		//new StaticSiteGeneratorPlugin('members.js', [], []),
		/*new webpack.optimize.UglifyJsPlugin({
				compressor: {
					warnings: false
				}
		}),*/
		new webpack.optimize.DedupePlugin()
	]
};
