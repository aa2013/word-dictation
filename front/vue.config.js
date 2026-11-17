const port = process.env.DEV_PORT || 8080

module.exports = {
    transpileDependencies: [
        'vuetify'
    ],
    publicPath: './',
    devServer: {
        port,
        // 开发时把 /api 与 /covers 代理到本地 Go 后端
        proxy: {
            '/api': {
                target: 'http://localhost:8099',
                changeOrigin: true
            },
            '/covers': {
                target: 'http://localhost:8099',
                changeOrigin: true
            }
        }
    }
}
