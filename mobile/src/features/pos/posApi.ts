import { api } from '../../api/client';
import { ApiSuccess, Order, PaymentMethod } from '../../types/api';

export interface CheckoutItem {
  product_id: number;
  quantity: number;
  discount?: number;
}

export interface CheckoutPayment {
  payment_method_id: number;
  amount_paid: number;
  reference?: string;
}

export interface CheckoutRequest {
  warehouse_id: number;
  customer_id?: number | null;
  discount?: number;
  tax?: number;
  notes?: string;
  items: CheckoutItem[];
  payments: CheckoutPayment[];
}

/** GET /payment-methods — active tender types (PRD §8.20). */
export async function fetchPaymentMethods(): Promise<PaymentMethod[]> {
  const { data } = await api.get<ApiSuccess<PaymentMethod[]>>('/payment-methods', {
    params: { active: 1 },
  });
  return data.data;
}

/** POST /orders — atomic checkout, returns the completed order/receipt (§8.19). */
export async function checkout(body: CheckoutRequest): Promise<Order> {
  const { data } = await api.post<ApiSuccess<Order>>('/orders', body);
  return data.data;
}
