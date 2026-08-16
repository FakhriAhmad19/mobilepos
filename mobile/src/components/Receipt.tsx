import { ScrollView, StyleSheet, Text, View } from 'react-native';

import { rupiah } from '../lib/format';
import { Order } from '../types/api';
import { colors } from '../theme/colors';

/** Renders an order as a receipt (PRD §8.22). */
export function Receipt({ order }: { order: Order }) {
  const totalPaid = (order.payments ?? []).reduce((s, p) => s + p.amount_paid, 0);
  const change = (order.payments ?? []).reduce((s, p) => s + p.change, 0);

  return (
    <ScrollView contentContainerStyle={styles.container}>
      <Text style={styles.brand}>KASIRKU</Text>
      <Text style={styles.sub}>Mobile POS</Text>
      <Text style={styles.trx}>{order.order_number}</Text>
      {order.status === 'cancelled' && <Text style={styles.void}>DIBATALKAN</Text>}

      <View style={styles.divider} />

      {(order.order_items ?? []).map((item) => (
        <View key={item.id} style={styles.itemRow}>
          <View style={styles.itemMain}>
            <Text style={styles.itemName}>{item.product_name}</Text>
            <Text style={styles.itemQty}>
              {item.quantity} × {rupiah(item.unit_price)}
            </Text>
          </View>
          <Text style={styles.itemTotal}>{rupiah(item.subtotal)}</Text>
        </View>
      ))}

      <View style={styles.divider} />

      <Line label="Subtotal" value={rupiah(order.subtotal)} />
      {order.discount > 0 && <Line label="Diskon" value={'-' + rupiah(order.discount)} />}
      {order.tax > 0 && <Line label="Pajak" value={rupiah(order.tax)} />}
      <Line label="Total" value={rupiah(order.grand_total)} bold />
      <Line label="Dibayar" value={rupiah(totalPaid)} />
      <Line label="Kembali" value={rupiah(change)} />

      <View style={styles.divider} />
      <Text style={styles.footer}>Terima kasih 🙏</Text>
    </ScrollView>
  );
}

function Line({ label, value, bold }: { label: string; value: string; bold?: boolean }) {
  return (
    <View style={styles.line}>
      <Text style={[styles.lineLabel, bold && styles.bold]}>{label}</Text>
      <Text style={[styles.lineValue, bold && styles.bold]}>{value}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { padding: 20 },
  brand: { fontSize: 22, fontWeight: '800', textAlign: 'center', color: colors.text },
  sub: { fontSize: 12, color: colors.textMuted, textAlign: 'center' },
  trx: { fontSize: 13, color: colors.textMuted, textAlign: 'center', marginTop: 8 },
  void: { color: colors.danger, fontWeight: '800', textAlign: 'center', marginTop: 6 },
  divider: {
    borderBottomWidth: StyleSheet.hairlineWidth,
    borderColor: colors.border,
    borderStyle: 'dashed',
    marginVertical: 12,
  },
  itemRow: { flexDirection: 'row', marginBottom: 8 },
  itemMain: { flex: 1, paddingRight: 12 },
  itemName: { fontSize: 14, fontWeight: '600', color: colors.text },
  itemQty: { fontSize: 12, color: colors.textMuted, marginTop: 2 },
  itemTotal: { fontSize: 14, fontWeight: '600', color: colors.text },
  line: { flexDirection: 'row', justifyContent: 'space-between', paddingVertical: 3 },
  lineLabel: { fontSize: 14, color: colors.textMuted },
  lineValue: { fontSize: 14, color: colors.text },
  bold: { fontWeight: '800', color: colors.text },
  footer: { textAlign: 'center', color: colors.textMuted, marginTop: 8 },
});
