import { useQuery } from '@tanstack/react-query';
import { useMemo, useState } from 'react';
import {
  ActivityIndicator,
  FlatList,
  Modal,
  RefreshControl,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';

import { Receipt } from '../components/Receipt';
import { fetchOrder } from '../features/pos/ordersApi';
import { useOrders } from '../features/pos/useOrders';
import { rupiah } from '../lib/format';
import { colors } from '../theme/colors';
import { Order } from '../types/api';

/** Transaction history (PRD §8.21) with tap-through to the receipt (§8.22). */
export function TransactionsScreen() {
  const [selectedId, setSelectedId] = useState<number | null>(null);

  const {
    data,
    isLoading,
    isError,
    refetch,
    isRefetching,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  } = useOrders();

  const orders = useMemo(() => data?.pages.flatMap((p) => p.items) ?? [], [data]);

  return (
    <View style={styles.container}>
      {isLoading ? (
        <View style={styles.center}>
          <ActivityIndicator size="large" color={colors.primary} />
        </View>
      ) : isError ? (
        <View style={styles.center}>
          <Text style={styles.muted}>Gagal memuat transaksi</Text>
        </View>
      ) : (
        <FlatList
          data={orders}
          keyExtractor={(item) => String(item.id)}
          contentContainerStyle={orders.length === 0 ? styles.center : styles.list}
          renderItem={({ item }) => (
            <TouchableOpacity style={styles.row} onPress={() => setSelectedId(item.id)}>
              <View style={styles.rowMain}>
                <Text style={styles.trx}>{item.order_number}</Text>
                <Text style={styles.meta}>
                  {(item.created_at ?? item.ordered_at ?? '').slice(0, 16)}
                  {item.status === 'cancelled' ? ' · dibatalkan' : ''}
                </Text>
              </View>
              <Text style={[styles.total, item.status === 'cancelled' && styles.cancelled]}>
                {rupiah(item.grand_total)}
              </Text>
            </TouchableOpacity>
          )}
          ListEmptyComponent={<Text style={styles.muted}>Belum ada transaksi</Text>}
          refreshControl={
            <RefreshControl refreshing={isRefetching} onRefresh={refetch} tintColor={colors.primary} />
          }
          onEndReachedThreshold={0.4}
          onEndReached={() => {
            if (hasNextPage && !isFetchingNextPage) fetchNextPage();
          }}
          ListFooterComponent={
            isFetchingNextPage ? (
              <ActivityIndicator style={{ marginVertical: 16 }} color={colors.primary} />
            ) : null
          }
        />
      )}

      <ReceiptModal id={selectedId} onClose={() => setSelectedId(null)} />
    </View>
  );
}

function ReceiptModal({ id, onClose }: { id: number | null; onClose: () => void }) {
  const { data: order, isLoading } = useQuery({
    queryKey: ['order', id],
    queryFn: () => fetchOrder(id as number),
    enabled: id !== null,
  });

  return (
    <Modal visible={id !== null} animationType="slide" onRequestClose={onClose}>
      <View style={styles.modalWrap}>
        {isLoading || !order ? (
          <View style={styles.center}>
            <ActivityIndicator size="large" color={colors.primary} />
          </View>
        ) : (
          <Receipt order={order as Order} />
        )}
        <TouchableOpacity style={styles.closeBtn} onPress={onClose}>
          <Text style={styles.closeText}>Tutup</Text>
        </TouchableOpacity>
      </View>
    </Modal>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  list: { padding: 16 },
  center: { flexGrow: 1, alignItems: 'center', justifyContent: 'center', padding: 24 },
  row: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: 12,
    padding: 14,
    marginBottom: 10,
  },
  rowMain: { flex: 1, paddingRight: 12 },
  trx: { fontSize: 14, fontWeight: '700', color: colors.text },
  meta: { fontSize: 12, color: colors.textMuted, marginTop: 2 },
  total: { fontSize: 15, fontWeight: '800', color: colors.primary },
  cancelled: { color: colors.textMuted, textDecorationLine: 'line-through' },
  muted: { color: colors.textMuted },
  modalWrap: { flex: 1, backgroundColor: colors.background },
  closeBtn: {
    backgroundColor: colors.primary,
    borderRadius: 10,
    paddingVertical: 14,
    alignItems: 'center',
    margin: 20,
  },
  closeText: { color: '#fff', fontSize: 16, fontWeight: '700' },
});
