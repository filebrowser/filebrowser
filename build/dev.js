process.env.NODE_ENV = 'development'

var rm = require('rimraf')
var path = require('path')
var chalk = require('chalk')
var webpack = require('webpack')
var config = require('./config')
var webpackConfig = require('./webpack.dev.conf')
var fs = require('fs')

if (fs.existsSync('./rice-box.go')) {
  fs.unlinkSync('./rice-box.go')
}

if (fs.existsSync('./plugins/rice-box.go')) {
  fs.unlinkSync('./plugins/rice-box.go')
}

rm(path.join(config.assetsRoot, config.assetsSubDirectory), err => {
  if (err) throw err
  webpack(webpackConfig, function (err, stats) {
    if (err) throw err
    process.stdout.write(stats.toString({
      colors: true,
      modules: false,
      children: false,
      chunks: false,
      chunkModules: false
    }) + '\n\n')

    console.log(chalk.cyan('  Build complete.\n'))
  })
})
