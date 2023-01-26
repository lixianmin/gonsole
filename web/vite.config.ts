import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'
import {resolve} from 'path'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import {ElementPlusResolver} from 'unplugin-vue-components/resolvers'
import vueJsx from '@vitejs/plugin-vue-jsx'

// https://vitejs.dev/config/
export default defineConfig({
    plugins: [
        vue(),
        vueJsx({}),
        AutoImport({
            resolvers: [ElementPlusResolver()],
        }),
        Components({
            resolvers: [ElementPlusResolver()],
        }),
    ],
    build: {
        rollupOptions: {
            input: {
                main: resolve(__dirname, 'console.html'),
                nested: resolve(__dirname, 'index.html')
            }
        }
    },
    resolve: {
        alias: [
            {   // 支持import时使用@代替代码根目录，另外还需要tsconfig.json中配合
                find: '@', replacement: resolve(__dirname, '/src')
            }
        ]
    }
})
