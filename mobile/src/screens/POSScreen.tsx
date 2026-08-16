import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useMemo, useState } from 'react';
import {
  ActivityIndicator,
  FlatList,
  Modal,
  StyleSheet,
  Text,
  TextInput,
  TouchableOpacity,
  View,
} from 'react-native';

import { Receipt } from '../components/Receipt';
import { useProducts } from '../features/products/useProducts';
import { useCartStore } from '../features/pos/cartStore';
import { checkout, fetchPaymentMethods } from '../features/pos/posApi';
import { rupiah } from '../lib/format';
import { colors } from '../theme/colors';
import { Order, PaymentMethod, Product } from '../types/api';

// The seeded default warehouse. A warehouse picker can replace this later.
const DEFAULT_WAREHOUSE_ID = 1;

export function POSScreen() {
  const queryClient = useQueryClient();
  const [term, setTerm] = useState('');
  const [search, setSearch] = useState('');
  const [payOpen, setPayOpen] = useState(false);
  const [receipt, setReceipt] = useState<Order | null>(null);

  const { lines, addProduct, setQuantity, clear, subtotal, itemCount } = useCartStore();
  const total = subtotal();

  const { data, isLoading } = useProducts(search);
  const products = useMemo(() => data?.pages.flatMap((p) => p.items) ?? [], [data]);

  const onCheckoutDone = (order: Order) => {
    clear();
    setPayOpen(false);
    setReceipt(order);
    queryClient.invalidateQueries({ queryKey: ['inventory'] });
    queryClient.invalidateQueries({ queryKey: ['orders'] });
    queryClient.invalidateQueries({ queryKey: ['products'] });
  };

  return (
    <View style={styles.container}>
      <TextInput
        style={styles.search}
        placeholder="Cari produk untuk ditambahkan…"
        value={term}
        onChangeText={setTerm}
        onSubmitEditing={() => setSearch(term.trim())}
        returnKeyType="search"
        autoCapitalize="none"
      />

      {isLoading ? (
        <View style={styles.centerSmall}>
          <ActivityIndicator color={colors.primary} />
        </View>
      ) : (
        <FlatList
          data={products}
          keyExtractor={(item) => String(item.id)}
          style={styles.results}
          keyboardShouldPersistTaps="handled"
          renderItem={({ item }) => <ProductPick product={item} onAdd={addProduct} />}
          ListEmptyComponent={<Text style={styles.muted}>Cari lalu ketuk produk untuk menambah</Text>}
        />
      )}

      {/* Cart */}
      <View style={styles.cart}>
        <Text style={styles.cartTitle}>Keranjang ({itemCount()})</Text>
        {lines.length === 0 ? (
          <Text style={styles.muted}>Keranjang kosong</Text>
        ) : (
          lines.map((l) => (
            <View key={l.product.id} style={styles.cartRow}>
              <View style={styles.cartMain}>
                <Text style={styles.cartName} numberOfLines={1}>
                  {l.product.name}
                </Text>
                <Text style={styles.cartPrice}>{rupiah(l.product.selling_price)}</Text>
              </View>
              <View style={styles.stepper}>
                <Stepper label="−" onPress={() => setQuantity(l.product.id, l.quantity - 1)} />
                <Text style={styles.qty}>{l.quantity}</Text>
                <Stepper label="+" onPress={() => setQuantity(l.product.id, l.quantity + 1)} />
              </View>
            </View>
          ))
        )}

        <View style={styles.totalRow}>
          <Text style={styles.totalLabel}>Total</Text>
          <Text style={styles.totalValue}>{rupiah(total)}</Text>
        </View>

        <TouchableOpacity
          style={[styles.payBtn, lines.length === 0 && styles.payBtnDisabled]}
          disabled={lines.length === 0}
          onPress={() => setPayOpen(true)}
        >
          <Text style={styles.payBtnText}>Bayar</Text>
        </TouchableOpacity>
      </View>

      <PaymentModal
        visible={payOpen}
        total={total}
        items={lines.map((l) => ({ product_id: l.product.id, quantity: l.quantity }))}
        onClose={() => setPayOpen(false)}
        onDone={onCheckoutDone}
      />

      <Modal visible={!!receipt} animationType="slide" onRequestClose={() => setReceipt(null)}>
        <View style={styles.receiptWrap}>
          {receipt && <Receipt order={receipt} />}
          <TouchableOpacity style={styles.newBtn} onPress={() => setReceipt(null)}>
            <Text style={styles.payBtnText}>Transaksi Baru</Text>
          </TouchableOpacity>
        </View>
      </Modal>
    </View>
  );
}

function ProductPick({ product, onAdd }: { product: Product; onAdd: (p: Product) => void }) {
  return (
    <TouchableOpacity style={styles.pick} onPress={() => onAdd(product)}>
      <View style={{ flex: 1 }}>
        <Text style={styles.pickName} numberOfLines={1}>
          {product.name}
        </Text>
        <Text style={styles.pickSku}>{product.sku}</Text>
      </View>
      <Text style={styles.pickPrice}>{rupiah(product.selling_price)}</Text>
    </TouchableOpacity>
  );
}

function Stepper({ label, onPress }: { label: string; onPress: () => void }) {
  return (
    <TouchableOpacity style={styles.stepBtn} onPress={onPress}>
      <Text style={styles.stepBtnText}>{label}</Text>
    </TouchableOpacity>
  );
}

function PaymentModal({
  visible,
  total,
  items,
  onClose,
  onDone,
}: {
  visible: boolean;
  total: number;
  items: { product_id: number; quantity: number }[];
  onClose: () => void;
  onDone: (order: Order) => void;
}) {
  const [methodId, setMethodId] = useState<number | null>(null);
  const [amount, setAmount] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const { data: methods } = useQuery({
    queryKey: ['payment-methods'],
    queryFn: fetchPaymentMethods,
    enabled: visible,
  });

  const selected = methodId ?? methods?.[0]?.id ?? null;
  const paid = Number(amount) || 0;
  const isCash = methods?.find((m) => m.id === selected)?.code === 'cash';
  const effectivePaid = isCash ? paid : total; // non-cash: exact amount
  const change = Math.max(0, effectivePaid - total);
  const canPay = selected !== null && effectivePaid + 0.001 >= total;

  const submit = async () => {
    if (!selected) return;
    setSubmitting(true);
    setError(null);
    try {
      const order = await checkout({
        warehouse_id: DEFAULT_WAREHOUSE_ID,
        items,
        payments: [{ payment_method_id: selected, amount_paid: effectivePaid }],
      });
      setAmount('');
      setMethodId(null);
      onDone(order);
    } catch (e) {
      setError((e as { message?: string })?.message ?? 'Checkout gagal');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Modal visible={visible} transparent animationType="slide" onRequestClose={onClose}>
      <View style={styles.modalBackdrop}>
        <View style={styles.modalCard}>
          <Text style={styles.modalTitle}>Pembayaran</Text>
          <Text style={styles.modalTotal}>{rupiah(total)}</Text>

          <Text style={styles.modalLabel}>Metode</Text>
          <View style={styles.methods}>
            {(methods ?? []).map((m: PaymentMethod) => (
              <TouchableOpacity
                key={m.id}
                style={[styles.method, selected === m.id && styles.methodActive]}
                onPress={() => setMethodId(m.id)}
              >
                <Text style={[styles.methodText, selected === m.id && styles.methodTextActive]}>
                  {m.name}
                </Text>
              </TouchableOpacity>
            ))}
          </View>

          {isCash && (
            <>
              <Text style={styles.modalLabel}>Uang diterima</Text>
              <TextInput
                style={styles.amount}
                keyboardType="number-pad"
                placeholder={String(total)}
                value={amount}
                onChangeText={setAmount}
              />
              <View style={styles.line}>
                <Text style={styles.modalLabel}>Kembalian</Text>
                <Text style={styles.change}>{rupiah(change)}</Text>
              </View>
            </>
          )}

          {error && <Text style={styles.error}>{error}</Text>}

          <View style={styles.modalActions}>
            <TouchableOpacity style={styles.cancelBtn} onPress={onClose}>
              <Text style={styles.cancelText}>Batal</Text>
            </TouchableOpacity>
            <TouchableOpacity
              style={[styles.confirmBtn, (!canPay || submitting) && styles.payBtnDisabled]}
              disabled={!canPay || submitting}
              onPress={submit}
            >
              {submitting ? (
                <ActivityIndicator color="#fff" />
              ) : (
                <Text style={styles.payBtnText}>Selesaikan</Text>
              )}
            </TouchableOpacity>
          </View>
        </View>
      </View>
    </Modal>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  search: {
    margin: 12,
    marginBottom: 6,
    backgroundColor: colors.surface,
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: 10,
    paddingHorizontal: 14,
    paddingVertical: 10,
    fontSize: 15,
    color: colors.text,
  },
  results: { maxHeight: 220, paddingHorizontal: 12 },
  centerSmall: { padding: 16, alignItems: 'center' },
  pick: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: 10,
    padding: 12,
    marginBottom: 8,
  },
  pickName: { fontSize: 14, fontWeight: '600', color: colors.text },
  pickSku: { fontSize: 11, color: colors.textMuted },
  pickPrice: { fontSize: 14, fontWeight: '700', color: colors.primary },
  muted: { color: colors.textMuted, padding: 8 },

  cart: {
    flex: 1,
    backgroundColor: colors.surface,
    borderTopWidth: 1,
    borderColor: colors.border,
    padding: 16,
  },
  cartTitle: { fontSize: 16, fontWeight: '800', color: colors.text, marginBottom: 8 },
  cartRow: { flexDirection: 'row', alignItems: 'center', paddingVertical: 8 },
  cartMain: { flex: 1, paddingRight: 12 },
  cartName: { fontSize: 14, fontWeight: '600', color: colors.text },
  cartPrice: { fontSize: 12, color: colors.textMuted },
  stepper: { flexDirection: 'row', alignItems: 'center' },
  stepBtn: {
    width: 32,
    height: 32,
    borderRadius: 8,
    backgroundColor: colors.background,
    borderWidth: 1,
    borderColor: colors.border,
    alignItems: 'center',
    justifyContent: 'center',
  },
  stepBtnText: { fontSize: 18, color: colors.text },
  qty: { minWidth: 32, textAlign: 'center', fontSize: 15, fontWeight: '700', color: colors.text },

  totalRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginTop: 12,
    paddingTop: 12,
    borderTopWidth: 1,
    borderColor: colors.border,
  },
  totalLabel: { fontSize: 16, fontWeight: '700', color: colors.text },
  totalValue: { fontSize: 18, fontWeight: '800', color: colors.primary },
  payBtn: {
    backgroundColor: colors.primary,
    borderRadius: 10,
    paddingVertical: 14,
    alignItems: 'center',
    marginTop: 12,
  },
  payBtnDisabled: { opacity: 0.4 },
  payBtnText: { color: '#fff', fontSize: 16, fontWeight: '700' },

  modalBackdrop: { flex: 1, backgroundColor: 'rgba(0,0,0,0.4)', justifyContent: 'flex-end' },
  modalCard: {
    backgroundColor: colors.surface,
    borderTopLeftRadius: 20,
    borderTopRightRadius: 20,
    padding: 20,
  },
  modalTitle: { fontSize: 16, fontWeight: '700', color: colors.text },
  modalTotal: { fontSize: 28, fontWeight: '800', color: colors.primary, marginVertical: 8 },
  modalLabel: { fontSize: 13, color: colors.textMuted, marginTop: 12, marginBottom: 6 },
  methods: { flexDirection: 'row', flexWrap: 'wrap' },
  method: {
    paddingHorizontal: 14,
    paddingVertical: 8,
    borderRadius: 18,
    borderWidth: 1,
    borderColor: colors.border,
    marginRight: 8,
    marginBottom: 8,
  },
  methodActive: { backgroundColor: colors.primary, borderColor: colors.primary },
  methodText: { color: colors.text, fontSize: 13 },
  methodTextActive: { color: '#fff', fontWeight: '700' },
  amount: {
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: 10,
    paddingHorizontal: 14,
    paddingVertical: 12,
    fontSize: 18,
    color: colors.text,
  },
  line: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' },
  change: { fontSize: 18, fontWeight: '800', color: colors.success },
  error: { color: colors.danger, marginTop: 12 },
  modalActions: { flexDirection: 'row', marginTop: 20 },
  cancelBtn: { flex: 1, paddingVertical: 14, alignItems: 'center' },
  cancelText: { color: colors.textMuted, fontWeight: '700' },
  confirmBtn: {
    flex: 2,
    backgroundColor: colors.primary,
    borderRadius: 10,
    paddingVertical: 14,
    alignItems: 'center',
  },

  receiptWrap: { flex: 1, backgroundColor: colors.background },
  newBtn: {
    backgroundColor: colors.primary,
    borderRadius: 10,
    paddingVertical: 14,
    alignItems: 'center',
    margin: 20,
  },
});
