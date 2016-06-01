var webpack = require('webpack');
var path = require('path');
var ExtractTextPlugin = require("extract-text-webpack-plugin");

module.exports = {
    entry: './app/jsx/index.jsx',
    output: {
        path: "./public",
        filename: 'js/bundle.js'
    },
    module: {
        loaders: [{
            test: /\.scss$/,
            loader: ExtractTextPlugin.extract("style", "css!sass")
        },{
            test: /\.css$/,
            loader: ExtractTextPlugin.extract("style", "css")
        },{
            test: /\.jsx?/,
            include: path.resolve(__dirname, 'app'),
            loader: 'babel'
        }]
    },
    plugins: [
        new ExtractTextPlugin("css/style.css", {
            allChunks: true
        }),
        new webpack.DefinePlugin({
            'process.env': {
                'NODE_ENV': JSON.stringify('production')
            }
        })
    ]
};
