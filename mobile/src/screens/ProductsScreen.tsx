import { useMemo, useState } from 'react';
import {
  ActivityIndicator,
  FlatList,
  RefreshControl,
  StyleSheet,
  Text,
  TextInput,
  View,
} from 'react-native';

import { useProducts } from '../features/products/useProducts';
import { colors } from '../theme/colors';
import { Product } from '../types/api';

const rupiah = (n: number) =>
  'Rp ' + Math.round(n).toLocaleString('id-ID');

/**
 * Product browse screen (PRD §8.16) — searchable, paginated list backed by
 * GET /products. Feeds product selection for POS in a later phase.
 */
export function ProductsScreen() {
  const [term, setTerm] = useState('');
  const [search, setSearch] = useState('');

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
  } = useProducts(search);

  const products = useMemo(
    () => data?.pages.flatMap((p) => p.items) ?? [],
    [data],
  );

  // Debounce-free simple submit: search on end of typing via onSubmitEditing.
  const applySearch = () => setSearch(term.trim());

  return (
    <View style={styles.container}>
      <TextInput
        style={styles.search}
        placeholder="Cari nama, SKU, atau barcode…"
        value={term}
        onChangeText={setTerm}
        onSubmitEditing={applySearch}
        returnKeyType="search"
        autoCapitalize="none"
      />

      {isLoading ? (
        <View style={styles.center}>
          <ActivityIndicator size="large" color={colors.primary} />
        </View>
      ) : isError ? (
        <View style={styles.center}>
          <Text style={styles.errorText}>
            {(error as { message?: string })?.message ?? 'Gagal memuat produk'}
          </Text>
        </View>
      ) : (
        <FlatList
          data={products}
          keyExtractor={(item) => String(item.id)}
          contentContainerStyle={products.length === 0 ? styles.center : styles.list}
          renderItem={({ item }) => <ProductRow product={item} />}
          ListEmptyComponent={<Text style={styles.muted}>Tidak ada produk</Text>}
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

function ProductRow({ product }: { product: Product }) {
  return (
    <View style={styles.row}>
      <View style={styles.rowMain}>
        <Text style={styles.name} numberOfLines={1}>
          {product.name}
        </Text>
        <Text style={styles.meta}>
          {product.sku}
          {product.category?.name ? ` · ${product.category.name}` : ''}
        </Text>
      </View>
      <Text style={styles.price}>{rupiah(product.selling_price)}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  search: {
    margin: 16,
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
  price: { fontSize: 15, fontWeight: '700', color: colors.primary },
  muted: { color: colors.textMuted },
  errorText: { color: colors.danger, textAlign: 'center' },
});
