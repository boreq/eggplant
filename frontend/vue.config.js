var webpack = require('webpack');

module.exports = {
    css: {
        loaderOptions: {
            sass: {
                prependData: `
                    @import "@/scss/variables.scss";
                `
            }
        }
    },
    productionSourceMap: false,
    configureWebpack: {
        plugins: [
            new webpack.IgnorePlugin(/^\.\/locale$/, /moment$/)
        ]
    }
};
