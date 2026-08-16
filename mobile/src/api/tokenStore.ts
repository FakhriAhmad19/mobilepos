import * as SecureStore from 'expo-secure-store';

/**
 * Secure persistence for the auth token. Uses the device keychain/keystore
 * via expo-secure-store so the JWT is never written to plain AsyncStorage.
 */
const TOKEN_KEY = 'kasirku.auth.token';

export async function getToken(): Promise<string | null> {
  return SecureStore.getItemAsync(TOKEN_KEY);
}

export async function setToken(token: string): Promise<void> {
  await SecureStore.setItemAsync(TOKEN_KEY, token);
}

export async function clearToken(): Promise<void> {
  await SecureStore.deleteItemAsync(TOKEN_KEY);
}
