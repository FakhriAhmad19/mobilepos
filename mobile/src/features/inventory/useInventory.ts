import { useInfiniteQuery } from '@tanstack/react-query';

import { fetchInventory, InventoryPage } from './inventoryApi';

/** Paginated stock levels with search + low-stock filter (PRD §8.9, §8.24). */
export function useInventory(search: string, lowStock: boolean) {
  return useInfiniteQuery({
    queryKey: ['inventory', search, lowStock],
    initialPageParam: 1,
    queryFn: ({ pageParam }) =>
      fetchInventory({ page: pageParam as number, search, lowStock }),
    getNextPageParam: (last: InventoryPage) =>
      last.page < last.lastPage ? last.page + 1 : undefined,
  });
}
