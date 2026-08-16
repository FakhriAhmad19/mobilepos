import { useInfiniteQuery } from '@tanstack/react-query';

import { fetchOrders, OrderPage } from './ordersApi';

/** Paginated transaction history (PRD §8.21). */
export function useOrders() {
  return useInfiniteQuery({
    queryKey: ['orders'],
    initialPageParam: 1,
    queryFn: ({ pageParam }) => fetchOrders(pageParam as number),
    getNextPageParam: (last: OrderPage) =>
      last.page < last.lastPage ? last.page + 1 : undefined,
  });
}
