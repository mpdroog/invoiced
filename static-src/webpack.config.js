var webpack = require("webpack");
var extract = require("extract-text-webpack-plugin");

module.exports = {
    cache: true,
    entry: "./app.tsx",
    output: {
        path: "../static/assets",
        filename: "app.js"
    },

    // Enable sourcemaps for debugging webpack's output.
    devtool: "eval",

    resolve: {
        modulesDirectories: [
            "node_modules"
        ],
        // Add '.ts' and '.tsx' as resolvable extensions.
        extensions: ["", ".webpack.js", ".web.js", ".ts", ".tsx", ".js"]
    },

    module: {
        loaders: [
            // All files with jsx extension get sent to babel
            {
                test: /\.jsx?$/,
                loader: 'babel', // 'babel-loader' is also a legal name to reference
                include: [
                    "./pages" //important for performance!
                ],
                query: {presets: ['es2015'], cacheDirectory: true}
            },

            // All files with a '.ts' or '.tsx' extension will be handled by 'ts-loader'.
            { test: /\.tsx?$/, loader: "ts-loader" },
            // Create CSS from all included styling files.
            { test: /\.css$/, loader: extract.extract("css-loader") }
        ],

        preLoaders: [
            // All output '.js' files will have any sourcemaps re-processed by 'source-map-loader'.
            //{ test: /\.js$/, loader: "source-map-loader" }
        ]
    },

    // Minify the output!
    plugins: [
        new webpack.DllReferencePlugin({
            context: ".",
            manifest: require("./vendor.json"),
        }),
        new extract("app.css", { allChunks: true })
        //new webpack.optimize.UglifyJsPlugin()
    ]
};