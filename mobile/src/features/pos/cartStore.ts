import { create } from 'zustand';

import { Product } from '../../types/api';

export interface CartLine {
  product: Product;
  quantity: number;
}

interface CartState {
  lines: CartLine[];
  addProduct: (product: Product) => void;
  setQuantity: (productId: number, quantity: number) => void;
  removeLine: (productId: number) => void;
  clear: () => void;
  subtotal: () => number;
  itemCount: () => number;
}

/** Client-side POS cart (PRD §8.18). Prices come from the product record. */
export const useCartStore = create<CartState>((set, get) => ({
  lines: [],

  addProduct: (product) =>
    set((state) => {
      const existing = state.lines.find((l) => l.product.id === product.id);
      if (existing) {
        return {
          lines: state.lines.map((l) =>
            l.product.id === product.id ? { ...l, quantity: l.quantity + 1 } : l,
          ),
        };
      }
      return { lines: [...state.lines, { product, quantity: 1 }] };
    }),

  setQuantity: (productId, quantity) =>
    set((state) => ({
      lines:
        quantity <= 0
          ? state.lines.filter((l) => l.product.id !== productId)
          : state.lines.map((l) =>
              l.product.id === productId ? { ...l, quantity } : l,
            ),
    })),

  removeLine: (productId) =>
    set((state) => ({ lines: state.lines.filter((l) => l.product.id !== productId) })),

  clear: () => set({ lines: [] }),

  subtotal: () =>
    get().lines.reduce((sum, l) => sum + l.product.selling_price * l.quantity, 0),

  itemCount: () => get().lines.reduce((sum, l) => sum + l.quantity, 0),
}));
