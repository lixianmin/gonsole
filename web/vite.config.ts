import {defineConfig} from 'vite';
import solidPlugin from 'vite-plugin-solid';
import {resolve} from 'path'

const root = resolve(__dirname, "src");

export default defineConfig({
    plugins: [solidPlugin()],
    resolve: {
        alias: {
            "@src": root,
        }
    },
    server: {
        port: 3000,
        https: {
            key: resolve(__dirname, '../res/ssl/localhost.key'),
            cert: resolve(__dirname, '../res/ssl/localhost.crt'),
        }
    },
    css: {  // 阻止vite在编译时把css选择器的名字改掉
        modules: false
    },
    build: {
        target: 'esnext',
        rollupOptions: {
            input: {
                main: resolve(__dirname, 'console.html'),
                nested: resolve(__dirname, 'index.html')
            }
        }
    },
});