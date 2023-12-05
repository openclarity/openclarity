import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import viteTsconfigPaths from 'vite-tsconfig-paths';

const PROXY_TARGET = 'http://localhost:8080';

export default defineConfig({
    base: '',
    plugins: [react(), viteTsconfigPaths()],
    define: {
        'process.env': process.env,
    },
    resolve: {
        alias: {
            utils: '/src/utils',
            process: 'process/browser',
        }
    },
    server: {    
        open: true,
        port: 3000,
        proxy: {
            '/api': {
                target: PROXY_TARGET,
                changeOrigin: true,
                secure: false,
            },
            '/ui/api': {
                target: PROXY_TARGET,
                changeOrigin: true,
                secure: false,
            },
        },
    },
});
