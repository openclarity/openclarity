const ModuleScopePlugin = require('react-dev-utils/ModuleScopePlugin');

module.exports = function override(config, env) {

    let loaders = config.resolve
    loaders.fallback = {
        "fs": require.resolve("fs-extra"),
        "assert": require.resolve("assert/"),
        "buffer": require.resolve("buffer/"),
        "constants": require.resolve("constants-browserify"),
        "path": require.resolve("path-browserify"),
        "stream": require.resolve("stream-browserify"),
        "util": require.resolve("util/"),
        'process': require.resolve('process/browser'),
    }

    loaders.plugins = loaders.plugins.filter(plugin => !(plugin instanceof ModuleScopePlugin));

    config.module.rules[1].oneOf.splice(0, 0, {
        test: /\.ya?ml$/,
        use: 'yaml-loader'
    })

    return config;
}
