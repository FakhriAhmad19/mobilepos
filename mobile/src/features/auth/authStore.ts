import { create } from 'zustand';

import { clearToken, getToken, setToken } from '../../api/tokenStore';
import { User } from '../../types/api';
import { fetchMe } from './authApi';

interface AuthState {
  user: User | null;
  token: string | null;
  /** True until the persisted token has been read/validated on app launch. */
  initializing: boolean;
  bootstrap: () => Promise<void>;
  signIn: (token: string, user: User) => Promise<void>;
  signOut: () => Promise<void>;
}

/**
 * Global auth state. The token is mirrored into SecureStore so sessions
 * survive app restarts; `bootstrap` rehydrates it and validates it against
 * the backend (`/auth/me`) before treating the user as logged in.
 */
export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  token: null,
  initializing: true,

  bootstrap: async () => {
    const token = await getToken();
    if (!token) {
      set({ token: null, user: null, initializing: false });
      return;
    }

    try {
      const user = await fetchMe();
      set({ token, user, initializing: false });
    } catch {
      // Token missing/expired/invalid — start clean.
      await clearToken();
      set({ token: null, user: null, initializing: false });
    }
  },

  signIn: async (token, user) => {
    await setToken(token);
    set({ token, user });
  },

  signOut: async () => {
    await clearToken();
    set({ token: null, user: null });
  },
}));
