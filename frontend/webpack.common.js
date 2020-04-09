const path = require('path');
const MiniCssExtractPlugin = require("mini-css-extract-plugin");
const HtmlWebpackPlugin = require('html-webpack-plugin');
const { CleanWebpackPlugin } = require('clean-webpack-plugin');
const ForkTsCheckerWebpackPlugin = require('fork-ts-checker-webpack-plugin');

var webpack = require("webpack");

module.exports = {
    context: __dirname,
    entry: {
        "controlpanel": path.resolve(__dirname, "src", "controlpanel.ts"),
        "semantic": path.resolve(__dirname, "semantic", "dist", "semantic.min.js"),
        "templates": path.resolve(__dirname, "src", "templates.ts"),
        "sidemenu": path.resolve(__dirname, "src", "sidemenu.ts")
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
        new webpack.ProvidePlugin({
            $: 'jquery',
            jQuery: 'jquery'
        }),
        new HtmlWebpackPlugin({
            filename: 'controlpanel.html',
            chunks: ['controlpanel', 'semantic'],
            template: path.resolve(__dirname, 'src', 'controlpanel.html')
        }),
        new HtmlWebpackPlugin({
            filename: 'templates.html',
            chunks: ['templates', 'semantic'],
            template: path.resolve(__dirname, 'src', 'templates.html')
        }),
        new HtmlWebpackPlugin({
            filename: 'sidemenu.html',
            chunks: ['sidemenu', 'semantic'],
            template: path.resolve(__dirname, 'src', 'sidemenu.html')
        })
        // new ForkTsCheckerWebpackPlugin()
    ]
};