import {defineConfig} from 'vite';
import solidPlugin from 'vite-plugin-solid';
import {resolve} from 'path'

export default defineConfig({
    plugins: [solidPlugin()],
    server: {
        port: 3000,
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