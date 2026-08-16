import { useQuery } from '@tanstack/react-query';

import { api } from '../../api/client';
import { ApiSuccess, HealthData } from '../../types/api';

/**
 * Polls the backend health endpoint (PRD §9.1) — used on the dashboard to
 * prove end-to-end connectivity from the app through to MySQL.
 */
export function useHealth() {
  return useQuery({
    queryKey: ['health'],
    queryFn: async (): Promise<HealthData> => {
      const { data } = await api.get<ApiSuccess<HealthData>>('/health');
      return data.data;
    },
    refetchInterval: 15_000,
  });
}
