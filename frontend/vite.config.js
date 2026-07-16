import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'node:path';

// Vite config — sostituisce craco.config.js (issue #25).
// - Plugin React + Fast Refresh
// - Alias `@` → src/ (allineato a jsconfig.json)
// - Build outDir 'build' per compatibilità con il Dockerfile esistente
//   (COPY --from=builder /app/build /usr/share/nginx/html)
// - Dev server proxy `/api` → backend locale (dev workflow `yarn dev`)
// - Env var prefisso `VITE_` (CRA usava `REACT_APP_`)
// Il progetto eredita da CRA tante pagine `.js` con JSX inline. Vite usa
// esbuild che di default tratta `.js` come puro JavaScript: senza override
// la prima `<Component />` solleva "Unexpected token <". Estendiamo il
// loader esbuild a `.js` (sia in dev sia in build) finché non rinominiamo
// in batch a `.jsx` (#27).
// react() del plugin Vite copre .jsx/.tsx. Per i .js legacy CRA con JSX inline
// e per i .ts puri (dopo #27), usiamo un esbuild plugin custom che assegna il
// loader corretto file-per-file: .js/.jsx → 'jsx', .ts/.tsx → 'tsx'.
const tsJsxLoader = {
  name: 'ts-jsx-loader',
  enforce: 'pre',
  async transform(code, id) {
    if (!id.includes('/src/')) return null;
    const m = id.match(/\.([jt]sx?)(?:\?|$)/);
    if (!m) return null;
    const ext = m[1];
    const loader =
      ext === 'ts' ? 'ts' :
      ext === 'tsx' ? 'tsx' :
      'jsx';
    const { transform } = await import('esbuild');
    const result = await transform(code, {
      loader,
      sourcemap: true,
      sourcefile: id,
      target: 'es2022',
      // FIX runtime "React is not defined": senza `jsx: 'automatic'`
      // esbuild usa il classic transform che genera `React.createElement`,
      // pero' nelle nostre pagine top-level (OrdersPage, App.js, ecc.)
      // importiamo solo gli hook named (`import { useState }`) e non
      // `import React from 'react'`. Il runtime automatic inietta
      // `import { jsx } from 'react/jsx-runtime'` ed evita la
      // dipendenza dal default export di React.
      jsx: 'automatic',
    });
    return { code: result.code, map: result.map };
  },
};

export default defineConfig({
  plugins: [tsJsxLoader, react({ include: /\.(jsx?|tsx?)$/ })],
  optimizeDeps: {
    esbuildOptions: { loader: { '.js': 'jsx' } },
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
    },
  },
  build: {
    outDir: 'build',
    sourcemap: true,
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8000',
        changeOrigin: true,
      },
    },
  },
});
