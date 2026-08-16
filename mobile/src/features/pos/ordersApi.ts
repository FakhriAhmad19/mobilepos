import { api } from '../../api/client';
import { ApiSuccess, Order } from '../../types/api';

export interface OrderPage {
  items: Order[];
  page: number;
  lastPage: number;
}

/** GET /orders — paginated transaction history (PRD §8.21). */
export async function fetchOrders(page = 1): Promise<OrderPage> {
  const { data } = await api.get<ApiSuccess<Order[]>>('/orders', {
    params: { page, per_page: 20 },
  });
  const meta = (data.meta ?? {}) as { current_page?: number; last_page?: number };
  return {
    items: data.data,
    page: meta.current_page ?? page,
    lastPage: meta.last_page ?? page,
  };
}

/** GET /orders/{id} — full order with items + payments (receipt, §8.22). */
export async function fetchOrder(id: number): Promise<Order> {
  const { data } = await api.get<ApiSuccess<Order>>(`/orders/${id}`);
  return data.data;
}
