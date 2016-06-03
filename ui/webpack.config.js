var webpack = require('webpack');
var path = require('path');
var ExtractTextPlugin = require("extract-text-webpack-plugin");
var autoprefixer = require('autoprefixer');

module.exports = {
    devtool: 'cheap-module-source-map',
    entry: './app/js',
    output: {
        path: path.join(__dirname, 'public'),
        filename: 'bundle.js'
    },
    module: {
        loaders: [{
            test: /\.scss$/,
            loader: ExtractTextPlugin.extract("style", "css!postcss-loader!sass")
        },{
            test: /\.less$/,
            loader: ExtractTextPlugin.extract("style", "css!postcss-loader!less")
        },{
            test: /\.css$/,
            loader: ExtractTextPlugin.extract("style", "css!postcss-loader")
        },{
            test: /\.js?/,
            include: path.resolve(__dirname, 'app'),
            loader: 'babel'
        },{
            test: /\.eot$/,
            loader: 'file?name=[name].[ext]',
        },{
            test: /\.(svg|ttf|woff|woff2)$/,
            loader: 'url-loader?limit=10000&name=[name].[ext]'
        }]
    },
    postcss: () => [autoprefixer],
    plugins: [
        new webpack.optimize.OccurrenceOrderPlugin(),
        new ExtractTextPlugin("style.css", {
            allChunks: true
        }),
        new webpack.DefinePlugin({
            'process.env': {
                'NODE_ENV': JSON.stringify('production')
            }
        })
    ]
};
