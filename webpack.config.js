const path = require("path");
const { WebpackManifestPlugin } = require("webpack-manifest-plugin");

module.exports = {
    entry: {
        main: "./frontend/main.tsx",
        oobe: "./frontend/oobe.tsx",
    },
    mode: process.env.NODE_ENV,
    devtool: "source-map",
    module: {
        rules: [
            {
                test: /\.tsx?$/,
                use: "ts-loader",
                exclude: /node_modules/,
            },
        ],
    },
    resolve: {
        extensions: [".tsx", ".ts", ".js"],
        alias: {
            "react": "preact/compat",
            "react-dom": "preact/compat",
        },
    },
    output: {
        filename: "[name].[contenthash].js",
        path: path.resolve(__dirname, "public", "bundles"),
    },
    plugins: [
        new WebpackManifestPlugin({}),
    ],
};
