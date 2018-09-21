// NOTES on this are found here:
//    https://cli.vuejs.org/config/#devserver
//    https://github.com/chimurai/http-proxy-middleware#proxycontext-config
module.exports = {
  devServer: {
    proxy: {
      '/api': {
        target: process.env.APOLLO_API, // or 'http://localhost:8085',
        changeOrigin: true,
        logLevel: 'debug'
      },
      '/authenticate': {
        target: process.env.APOLLO_API, // or 'http://localhost:8085',
        changeOrigin: true,
        logLevel: 'debug'
      }
    }
  }
}
