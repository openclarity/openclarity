import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import viteTsconfigPaths from 'vite-tsconfig-paths';
import { esbuildPluginBrowserslist } from 'esbuild-plugin-browserslist';
import { browserslist } from './package.json';

const PROXY_TARGET = 'http://localhost:8080';

export default defineConfig(({ mode }) => {
    return {
        base: '/',
        plugins: [
            react(),
            viteTsconfigPaths(),
            esbuildPluginBrowserslist(mode === 'production' ? browserslist.production : browserslist.development),
        ],
        build: {
            outDir: 'build',
        },
        resolve: {
            alias: {
                utils: '/src/utils',
            }
        },
        server: {    
            open: true,
            port: 3000,
            proxy: {
                '/api': {
                    target: PROXY_TARGET,
                },
                '/ui/api': {
                    target: PROXY_TARGET,
                },
            },
        },
        test: {
            globals: true,
            environment: 'happy-dom'
        }
    }
});
