import { useQuery } from '@tanstack/react-query';

import { fetchDashboard } from './dashboardApi';

/** Dashboard KPIs + charts, refreshed periodically (PRD §8.23). */
export function useDashboard() {
  return useQuery({
    queryKey: ['dashboard'],
    queryFn: fetchDashboard,
    refetchInterval: 30_000,
  });
}
