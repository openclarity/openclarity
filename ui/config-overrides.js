const ModuleScopePlugin = require('react-dev-utils/ModuleScopePlugin');
const webpack = require('webpack');

module.exports = function override(config, env) {
    // Explanation: https://web3auth.io/community/t/webpack-5-polyfills-issue-in-react/3126/2
    let loaders = config.resolve
    loaders.fallback = {
        "assert": require.resolve("assert/"),
        "buffer": require.resolve("buffer/"),
        "constants": require.resolve("constants-browserify"),
        "fs": require.resolve("fs-extra"),
        "http": require.resolve("stream-http"),
        "https": require.resolve("https-browserify"),
        "path": require.resolve("path-browserify"),
        "stream": require.resolve("stream-browserify"),
        "url": require.resolve("url/"),
        "util": require.resolve("util/"),
        'process': require.resolve('process/browser'),
    }

    loaders.plugins = loaders.plugins.filter(plugin => !(plugin instanceof ModuleScopePlugin));
    config.plugins = (config.plugins || []).concat([
        new webpack.ProvidePlugin({
            process: "process/browser",
            Buffer: ["buffer", "Buffer"],
        }),
    ]);
    config.ignoreWarnings = [/Failed to parse source map/];
    config.module.rules.push({
        test: /\.(js|mjs|jsx)$/,
        enforce: "pre",
        loader: require.resolve("source-map-loader"),
        resolve: {
            fullySpecified: false,
        },
    });

    // to import a yaml file, we have to provide the yaml-loader before file-loader catches it 
    config.module.rules[1].oneOf.splice(0, 0, {
        test: /\.ya?ml$/,
        use: 'yaml-loader'
    })

    return config;
}
