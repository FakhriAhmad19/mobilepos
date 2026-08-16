import { Platform } from 'react-native';
import * as SecureStore from 'expo-secure-store';

/**
 * Secure persistence for the auth token. On native it uses the device
 * keychain/keystore via expo-secure-store so the JWT is never written to plain
 * AsyncStorage. expo-secure-store has no web implementation, so on web (dev/demo
 * in a browser) it falls back to localStorage.
 */
const TOKEN_KEY = 'kasirku.auth.token';

const isWeb = Platform.OS === 'web';

export async function getToken(): Promise<string | null> {
  if (isWeb) {
    return globalThis.localStorage?.getItem(TOKEN_KEY) ?? null;
  }
  return SecureStore.getItemAsync(TOKEN_KEY);
}

export async function setToken(token: string): Promise<void> {
  if (isWeb) {
    globalThis.localStorage?.setItem(TOKEN_KEY, token);
    return;
  }
  await SecureStore.setItemAsync(TOKEN_KEY, token);
}

export async function clearToken(): Promise<void> {
  if (isWeb) {
    globalThis.localStorage?.removeItem(TOKEN_KEY);
    return;
  }
  await SecureStore.deleteItemAsync(TOKEN_KEY);
}
