const LiveReloadPlugin = require('webpack-livereload-plugin');
const path = require('path');
const MiniCssExtractPlugin = require("mini-css-extract-plugin");
const HtmlWebpackPlugin = require('html-webpack-plugin');
const { CleanWebpackPlugin } = require('clean-webpack-plugin');

module.exports = {
    context: __dirname,
    entry: {
        "main": path.resolve(__dirname, "src", "main.ts"),
        "xterm": path.resolve(__dirname, "node_modules", "xterm", "lib", "xterm.js")
    },
    output: {
        filename: '[name].bundle.js',
        path: path.resolve(__dirname, 'dist'),
        globalObject: 'this'
    },
    resolve: {
        modules: [
            path.resolve(__dirname, "src"),
            'node_modules',
            path.resolve(__dirname, "node_modules"),
        ],
        extensions: ['.tsx', '.ts', '.js']
    },
    module: {
        rules: [
            {
                test: /\.css$/,
                use: ['style-loader', MiniCssExtractPlugin.loader, 'css-loader']
            },
            {
                test: /\.tsx?$/,
                exclude: /node_modules/,
                use: [
                    {
                        loader: 'ts-loader',
                        options: {
                            transpileOnly: true,
                            // experimentalWatchApi: true,
                        },
                    },
                ],
            },
            {
                test: /\.(png|woff|woff2|eot|ttf|svg)$/,
                loader: 'url-loader?limit=100000'
            }
        ]
    },
    target: 'web',
    node: {
        fs: 'empty',
        child_process: 'empty',
        net: 'empty',
        crypto: 'empty'
    },
    // optimization: {
    //     splitChunks: {
    //         chunks: 'async'
    //     }
    // },
    plugins: [
        new CleanWebpackPlugin(),
        new MiniCssExtractPlugin({
            filename: "style.[contenthash].css"
        }),
        new HtmlWebpackPlugin({
            filename: 'index.html',
            chunks: ['main', 'xterm'],
            template: path.resolve(__dirname, 'src', 'index.html')
        }),
        new LiveReloadPlugin({ appendScriptTag: true })
    ],
    mode: 'development',
    devtool: 'inline-source-map',
    devServer: {
        contentBase: './dist',
        writeToDisk: true
    }
};