var webpack = require('webpack');
var path = require('path');
var ExtractTextPlugin = require("extract-text-webpack-plugin");

module.exports = {
    devtool: 'cheap-module-source-map',
    entry: './app/js',
    output: {
        path: path.join(__dirname, 'public'),
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
            test: /\.js?/,
            include: path.resolve(__dirname, 'app'),
            loader: 'babel'
        }]
    },
    plugins: [
        new webpack.optimize.OccurrenceOrderPlugin(),
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
