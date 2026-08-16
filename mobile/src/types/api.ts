/**
 * Standard API envelopes as defined in the PRD (§11).
 */

export interface ApiSuccess<T> {
  success: true;
  message: string;
  data: T;
  meta?: Record<string, unknown>;
}

export interface ApiError {
  success: false;
  message: string;
  errors?: Record<string, string[] | string>;
}

export type ApiResponse<T> = ApiSuccess<T> | ApiError;

/** Roles from the PRD's RBAC model (§8.2). */
export type Role = 'admin' | 'cashier' | 'warehouse';

export interface User {
  id: number;
  name: string;
  email: string;
  phone?: string;
  role: Role;
  status: 'active' | 'inactive';
  last_login_at?: string | null;
}

export interface HealthData {
  service: string;
  env: string;
  database: 'up' | 'down';
  timestamp: string;
}

/** Pagination block returned in `meta` for list endpoints (PRD §11). */
export interface PaginationMeta {
  current_page: number;
  per_page: number;
  total: number;
  last_page: number;
}

export interface Paginated<T> {
  data: T[];
  meta: PaginationMeta;
}

export interface Category {
  id: number;
  name: string;
  description?: string;
  status: 'active' | 'inactive';
}

export interface Unit {
  id: number;
  name: string;
  code: string;
}

export interface Product {
  id: number;
  sku: string;
  barcode?: string;
  name: string;
  category_id: number;
  unit_id: number;
  purchase_price: number;
  selling_price: number;
  minimum_stock: number;
  status: 'active' | 'inactive';
  category?: Category | null;
  unit?: Unit | null;
}

export interface PaymentMethod {
  id: number;
  code: string;
  name: string;
  is_active: boolean;
}

export interface OrderItem {
  id: number;
  product_id: number;
  product_name: string;
  sku: string;
  unit_price: number;
  quantity: number;
  discount: number;
  subtotal: number;
}

export interface Payment {
  id: number;
  payment_method_id: number;
  amount: number;
  amount_paid: number;
  change: number;
  status: string;
  payment_method?: PaymentMethod | null;
}

export interface Order {
  id: number;
  order_number: string;
  user_id: number;
  customer_id?: number | null;
  warehouse_id: number;
  subtotal: number;
  discount: number;
  tax: number;
  grand_total: number;
  status: 'completed' | 'cancelled' | 'pending';
  notes?: string;
  ordered_at?: string | null;
  created_at?: string;
  user?: User | null;
  customer?: { id: number; name: string } | null;
  order_items?: OrderItem[];
  payments?: Payment[];
}
