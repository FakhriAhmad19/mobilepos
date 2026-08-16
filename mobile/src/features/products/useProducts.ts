import { useInfiniteQuery } from '@tanstack/react-query';

import { fetchProducts, ProductPage } from './productsApi';

/**
 * Infinite (paginated) product list with optional search (PRD §8.16).
 * Fetches the next page when the list reaches its end.
 */
export function useProducts(search: string) {
  return useInfiniteQuery({
    queryKey: ['products', search],
    initialPageParam: 1,
    queryFn: ({ pageParam }) => fetchProducts({ page: pageParam as number, search }),
    getNextPageParam: (last: ProductPage) =>
      last.page < last.lastPage ? last.page + 1 : undefined,
  });
}
