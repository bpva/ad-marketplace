export const MOCK_TG_USER = {
  id: Number(import.meta.env.VITE_MOCK_TG_ID) || 77777777,
  first_name: import.meta.env.VITE_MOCK_TG_FIRST_NAME || "Dev",
  last_name: import.meta.env.VITE_MOCK_TG_LAST_NAME || "User",
  username: import.meta.env.VITE_MOCK_TG_USERNAME || "devuser",
};
