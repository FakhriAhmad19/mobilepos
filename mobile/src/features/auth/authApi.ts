import { api } from '../../api/client';
import { ApiSuccess, User } from '../../types/api';

interface LoginResponse {
  token: string;
  token_type: string;
  user: User;
}

/** POST /auth/login — exchanges credentials for a JWT (PRD §8.1). */
export async function login(email: string, password: string): Promise<LoginResponse> {
  const { data } = await api.post<ApiSuccess<LoginResponse>>('/auth/login', {
    email,
    password,
  });
  return data.data;
}

/** GET /auth/me — returns the authenticated user; used to validate a stored token. */
export async function fetchMe(): Promise<User> {
  const { data } = await api.get<ApiSuccess<User>>('/auth/me');
  return data.data;
}

/** POST /auth/logout — invalidates the current token server-side. */
export async function logout(): Promise<void> {
  await api.post('/auth/logout');
}
