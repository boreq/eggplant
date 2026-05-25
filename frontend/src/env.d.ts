/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VUE_APP_API_PREFIX: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
