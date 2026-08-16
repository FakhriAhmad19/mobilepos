import { useMemo, useState } from 'react';
import {
  ActivityIndicator,
  FlatList,
  RefreshControl,
  StyleSheet,
  Text,
  TextInput,
  TouchableOpacity,
  View,
} from 'react-native';

import { StockLevel } from '../features/inventory/inventoryApi';
import { useInventory } from '../features/inventory/useInventory';
import { colors } from '../theme/colors';

/**
 * Inventory screen (PRD §8.9) — stock levels per product/warehouse with a
 * low-stock filter (§8.24). Read-only; stock operations are warehouse-role
 * actions added in a later phase.
 */
export function InventoryScreen() {
  const [term, setTerm] = useState('');
  const [search, setSearch] = useState('');
  const [lowOnly, setLowOnly] = useState(false);

  const {
    data,
    isLoading,
    isError,
    error,
    refetch,
    isRefetching,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  } = useInventory(search, lowOnly);

  const rows = useMemo(() => data?.pages.flatMap((p) => p.items) ?? [], [data]);

  return (
    <View style={styles.container}>
      <TextInput
        style={styles.search}
        placeholder="Cari produk…"
        value={term}
        onChangeText={setTerm}
        onSubmitEditing={() => setSearch(term.trim())}
        returnKeyType="search"
        autoCapitalize="none"
      />

      <TouchableOpacity
        style={[styles.filter, lowOnly && styles.filterActive]}
        onPress={() => setLowOnly((v) => !v)}
      >
        <Text style={[styles.filterText, lowOnly && styles.filterTextActive]}>
          {lowOnly ? '● ' : '○ '}Stok menipis saja
        </Text>
      </TouchableOpacity>

      {isLoading ? (
        <View style={styles.center}>
          <ActivityIndicator size="large" color={colors.primary} />
        </View>
      ) : isError ? (
        <View style={styles.center}>
          <Text style={styles.errorText}>
            {(error as { message?: string })?.message ?? 'Gagal memuat inventori'}
          </Text>
        </View>
      ) : (
        <FlatList
          data={rows}
          keyExtractor={(item) => String(item.id)}
          contentContainerStyle={rows.length === 0 ? styles.center : styles.list}
          renderItem={({ item }) => <StockRow row={item} />}
          ListEmptyComponent={<Text style={styles.muted}>Tidak ada data stok</Text>}
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
    </View>
  );
}

function StockRow({ row }: { row: StockLevel }) {
  return (
    <View style={styles.row}>
      <View style={styles.rowMain}>
        <Text style={styles.name} numberOfLines={1}>
          {row.product?.name ?? `Produk #${row.product_id}`}
        </Text>
        <Text style={styles.meta}>
          {row.product?.sku ?? ''}
          {row.warehouse?.name ? ` · ${row.warehouse.name}` : ''}
        </Text>
      </View>
      <View style={styles.qtyBox}>
        <Text style={[styles.qty, row.is_low_stock && styles.qtyLow]}>{row.available_stock}</Text>
        {row.is_low_stock && <Text style={styles.lowBadge}>menipis</Text>}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  search: {
    marginHorizontal: 16,
    marginTop: 16,
    marginBottom: 8,
    backgroundColor: colors.surface,
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: 10,
    paddingHorizontal: 14,
    paddingVertical: 12,
    fontSize: 15,
    color: colors.text,
  },
  filter: {
    marginHorizontal: 16,
    marginBottom: 8,
    alignSelf: 'flex-start',
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 16,
    borderWidth: 1,
    borderColor: colors.border,
  },
  filterActive: { backgroundColor: colors.primary, borderColor: colors.primary },
  filterText: { fontSize: 13, color: colors.textMuted },
  filterTextActive: { color: '#fff', fontWeight: '600' },
  list: { paddingHorizontal: 16, paddingBottom: 24 },
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
  name: { fontSize: 15, fontWeight: '700', color: colors.text },
  meta: { fontSize: 12, color: colors.textMuted, marginTop: 2 },
  qtyBox: { alignItems: 'flex-end' },
  qty: { fontSize: 18, fontWeight: '800', color: colors.text },
  qtyLow: { color: colors.danger },
  lowBadge: { fontSize: 10, color: colors.danger, fontWeight: '700', marginTop: 2 },
  muted: { color: colors.textMuted },
  errorText: { color: colors.danger, textAlign: 'center' },
});
