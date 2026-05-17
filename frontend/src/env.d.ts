/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VUE_APP_API_PREFIX: string;
  readonly VUE_APP_WS_PREFIX: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
