import { api } from '../../api/client';
import { ApiSuccess } from '../../types/api';

export interface DashboardSummary {
  total_products: number;
  total_stock: number;
  low_stock: number;
  today_sales: number;
  today_transactions: number;
}

export interface SalesPoint {
  date: string;
  total: number;
  count: number;
}

export interface TopProduct {
  product_id: number;
  product_name: string;
  qty: number;
  revenue: number;
}

export interface DashboardData {
  summary: DashboardSummary;
  sales_7_days: SalesPoint[];
  top_products: TopProduct[];
}

/** GET /dashboard — KPI summary, 7-day sales, top products (PRD §8.23). */
export async function fetchDashboard(): Promise<DashboardData> {
  const { data } = await api.get<ApiSuccess<DashboardData>>('/dashboard');
  return data.data;
}
