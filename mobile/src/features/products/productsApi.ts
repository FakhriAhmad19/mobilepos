import { api } from '../../api/client';
import { ApiSuccess, Product } from '../../types/api';

export interface ProductPage {
  items: Product[];
  page: number;
  lastPage: number;
  total: number;
}

interface FetchParams {
  page?: number;
  perPage?: number;
  search?: string;
  categoryId?: number;
}

/** GET /products — paginated, searchable product list (PRD §8.4, §8.16). */
export async function fetchProducts({
  page = 1,
  perPage = 20,
  search,
  categoryId,
}: FetchParams): Promise<ProductPage> {
  const { data } = await api.get<ApiSuccess<Product[]>>('/products', {
    params: {
      page,
      per_page: perPage,
      ...(search ? { search } : {}),
      ...(categoryId ? { category_id: categoryId } : {}),
    },
  });

  const meta = (data.meta ?? {}) as { current_page?: number; last_page?: number; total?: number };
  return {
    items: data.data,
    page: meta.current_page ?? page,
    lastPage: meta.last_page ?? page,
    total: meta.total ?? data.data.length,
  };
}
