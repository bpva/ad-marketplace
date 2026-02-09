interface ImportMetaEnv {
  readonly [key: string]: any;
  readonly VITE_API_URL: string;
  readonly VITE_ENV: string;
  readonly VITE_BOT_TOKEN: string;
  readonly VITE_MOCK_TG_ID: string;
  readonly VITE_MOCK_TG_FIRST_NAME: string;
  readonly VITE_MOCK_TG_LAST_NAME: string;
  readonly VITE_MOCK_TG_USERNAME: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
  glob<T = unknown>(pattern: string): Record<string, () => Promise<T>>;
  glob<T = unknown>(patterns: string[]): Record<string, () => Promise<T>>;
}
