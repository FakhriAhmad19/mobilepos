import { api } from '../../api/client';
import { ApiSuccess, Product } from '../../types/api';

export interface StockLevel {
  id: number;
  product_id: number;
  warehouse_id: number;
  quantity: number;
  reserved: number;
  available_stock: number;
  is_low_stock: boolean;
  product?: (Product & { category?: { id: number; name: string } | null }) | null;
  warehouse?: { id: number; name: string } | null;
}

export interface InventoryPage {
  items: StockLevel[];
  page: number;
  lastPage: number;
}

interface FetchParams {
  page?: number;
  search?: string;
  lowStock?: boolean;
}

/** GET /inventory — paginated stock levels (PRD §8.9). */
export async function fetchInventory({
  page = 1,
  search,
  lowStock,
}: FetchParams): Promise<InventoryPage> {
  const { data } = await api.get<ApiSuccess<StockLevel[]>>('/inventory', {
    params: {
      page,
      per_page: 20,
      ...(search ? { search } : {}),
      ...(lowStock ? { low_stock: 1 } : {}),
    },
  });
  const meta = (data.meta ?? {}) as { current_page?: number; last_page?: number };
  return {
    items: data.data,
    page: meta.current_page ?? page,
    lastPage: meta.last_page ?? page,
  };
}
