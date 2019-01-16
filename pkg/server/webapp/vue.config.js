module.exports = {
  baseUrl: process.env.NODE_ENV === 'production'
          ? '/webapp/'
          : '/',
  devServer: {
    proxy: {
      '/api': {
        target: 'http://localhost:9876',
        changeOrigin: true
      }
    }
  }
};