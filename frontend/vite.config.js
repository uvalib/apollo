/*global process */

import { fileURLToPath, URL } from 'url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
   define: {
      // enable hydration mismatch details in production build
      __VUE_PROD_HYDRATION_MISMATCH_DETAILS__: 'true'
   },
   plugins: [vue()],
   resolve: {
      alias: {
         '@': fileURLToPath(new URL('./src', import.meta.url))
      }
   },
   server: { // this is used in dev mode only
      port: 8080,
      proxy: {
         '/api': {
            target: process.env.APOLLO_API,  //export APOLLO_API=http://localhost:8085
            changeOrigin: true
         },
         '/authenticate': {
            target: process.env.APOLLO_API,
            changeOrigin: true
         },
         '/config': {
            target: process.env.APOLLO_API,
            changeOrigin: true
         },
         '/healthcheck': {
            target: process.env.APOLLO_API,
            changeOrigin: true
         },
         '/version': {
            target: process.env.APOLLO_API,
            changeOrigin: true
         },
      }
   },
   css: {
      preprocessorOptions : {
          scss: {
              api: "modern-compiler",
          },
      }
   },
})


