/**
 * Runtime configuration sourced from Expo public env vars.
 * Only variables prefixed with EXPO_PUBLIC_ are inlined into the app bundle.
 */
const API_URL = process.env.EXPO_PUBLIC_API_URL ?? 'http://localhost:3000/api/v1';

export const env = {
  apiUrl: API_URL,
} as const;
