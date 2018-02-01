// see http://vuejs-templates.github.io/webpack for documentation.
var path = require('path')

module.exports = {
  index: path.resolve(__dirname, '../dist/index.html'),
  assetsRoot: path.resolve(__dirname, '../dist'),
  assetsSubDirectory: 'static',
  assetsPublicPath: '{{ .BaseURL }}/',
  build: {
    env: {
      NODE_ENV: '"production"'
    },
    productionSourceMap: true,
    // Run the build command with an extra argument to
    // View the bundle analyzer report after build finishes:
    // `npm run build --report`
    // Set to `true` or `false` to always turn it on or off
    bundleAnalyzerReport: process.env.npm_config_report
  },
  dev: {
    env: {
      NODE_ENV: '"development"'
    },
    produceSourceMap: true
  }
}
