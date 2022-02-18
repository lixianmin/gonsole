import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'
import {resolve} from 'path'

// https://vitejs.dev/config/
export default defineConfig({
    plugins: [vue()],
    // build: {
    //     rollupOptions: {
    //         input: {
    //             main: resolve(__dirname, 'console.html'),
    //             nested: resolve(__dirname, 'index.html')
    //         }
    //     }
    // }
})
