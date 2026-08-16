import axios, { AxiosError, AxiosInstance } from 'axios';

import { env } from '../config/env';
import { ApiError } from '../types/api';
import { getToken } from './tokenStore';

/**
 * Shared axios instance. Attaches the bearer token on every request and
 * normalises backend error envelopes into a rejected ApiError.
 */
export const api: AxiosInstance = axios.create({
  baseURL: env.apiUrl,
  timeout: 15000,
  headers: { Accept: 'application/json' },
});

// Callback invoked when the API returns 401 so the app can log the user out.
let onUnauthorized: (() => void) | null = null;
export function setUnauthorizedHandler(handler: () => void): void {
  onUnauthorized = handler;
}

api.interceptors.request.use(async (config) => {
  const token = await getToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

api.interceptors.response.use(
  (response) => response,
  (error: AxiosError<ApiError>) => {
    if (error.response?.status === 401) {
      onUnauthorized?.();
    }

    const normalised: ApiError = error.response?.data ?? {
      success: false,
      message: error.message || 'Network error',
    };
    return Promise.reject(normalised);
  },
);
