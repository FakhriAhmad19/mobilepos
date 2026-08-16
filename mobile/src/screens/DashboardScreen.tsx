import { ScrollView, StyleSheet, Text, TouchableOpacity, View } from 'react-native';

import { SalesPoint } from '../features/dashboard/dashboardApi';
import { useDashboard } from '../features/dashboard/useDashboard';
import { logout as logoutApi } from '../features/auth/authApi';
import { useAuthStore } from '../features/auth/authStore';
import { rupiah } from '../lib/format';
import { colors } from '../theme/colors';

export function DashboardScreen() {
  const user = useAuthStore((s) => s.user);
  const signOut = useAuthStore((s) => s.signOut);
  const { data, isLoading, isError, refetch, isRefetching } = useDashboard();

  const handleSignOut = async () => {
    try {
      await logoutApi();
    } catch {
      // ignore — clearing locally is enough
    }
    await signOut();
  };

  return (
    <ScrollView contentContainerStyle={styles.container}>
      <Text style={styles.greeting}>Halo, {user?.name ?? 'Pengguna'}</Text>
      <Text style={styles.role}>Role: {user?.role ?? '-'}</Text>

      {isLoading ? (
        <Text style={styles.muted}>Memuat dashboard…</Text>
      ) : isError ? (
        <View style={styles.card}>
          <Text style={styles.errorText}>Gagal memuat dashboard</Text>
          <TouchableOpacity style={styles.smallBtn} onPress={() => refetch()}>
            <Text style={styles.smallBtnText}>Coba lagi</Text>
          </TouchableOpacity>
        </View>
      ) : data ? (
        <>
          <View style={styles.kpiGrid}>
            <Kpi label="Penjualan Hari Ini" value={rupiah(data.summary.today_sales)} highlight />
            <Kpi label="Transaksi Hari Ini" value={String(data.summary.today_transactions)} />
            <Kpi label="Total Produk" value={String(data.summary.total_products)} />
            <Kpi label="Total Stok" value={String(data.summary.total_stock)} />
          </View>

          {data.summary.low_stock > 0 && (
            <View style={styles.alert}>
              <Text style={styles.alertText}>
                ⚠️  {data.summary.low_stock} produk mencapai stok minimum
              </Text>
            </View>
          )}

          <View style={styles.card}>
            <Text style={styles.cardTitle}>Penjualan 7 Hari Terakhir</Text>
            <BarChart points={data.sales_7_days} />
          </View>

          <View style={styles.card}>
            <Text style={styles.cardTitle}>Produk Terlaris</Text>
            {data.top_products.length === 0 ? (
              <Text style={styles.muted}>Belum ada penjualan</Text>
            ) : (
              data.top_products.map((p, i) => (
                <View key={p.product_id} style={styles.topRow}>
                  <Text style={styles.topRank}>{i + 1}</Text>
                  <Text style={styles.topName} numberOfLines={1}>
                    {p.product_name}
                  </Text>
                  <Text style={styles.topQty}>{p.qty} terjual</Text>
                </View>
              ))
            )}
          </View>

          <TouchableOpacity style={styles.smallBtn} onPress={() => refetch()}>
            <Text style={styles.smallBtnText}>{isRefetching ? 'Menyegarkan…' : 'Segarkan'}</Text>
          </TouchableOpacity>
        </>
      ) : null}

      <TouchableOpacity style={styles.signOut} onPress={handleSignOut}>
        <Text style={styles.signOutText}>Keluar</Text>
      </TouchableOpacity>
    </ScrollView>
  );
}

function Kpi({ label, value, highlight }: { label: string; value: string; highlight?: boolean }) {
  return (
    <View style={[styles.kpi, highlight && styles.kpiHighlight]}>
      <Text style={[styles.kpiValue, highlight && styles.kpiValueHighlight]}>{value}</Text>
      <Text style={[styles.kpiLabel, highlight && styles.kpiLabelHighlight]}>{label}</Text>
    </View>
  );
}

function BarChart({ points }: { points: SalesPoint[] }) {
  const max = Math.max(1, ...points.map((p) => p.total));
  return (
    <View style={styles.chart}>
      {points.map((p) => (
        <View key={p.date} style={styles.chartCol}>
          <View style={styles.barTrack}>
            <View style={[styles.bar, { height: `${Math.round((p.total / max) * 100)}%` }]} />
          </View>
          <Text style={styles.chartLabel}>{p.date.slice(5)}</Text>
        </View>
      ))}
    </View>
  );
}

const styles = StyleSheet.create({
  container: { padding: 20, backgroundColor: colors.background, flexGrow: 1 },
  greeting: { fontSize: 24, fontWeight: '800', color: colors.text },
  role: { fontSize: 14, color: colors.textMuted, marginBottom: 16 },
  muted: { color: colors.textMuted, paddingVertical: 8 },
  errorText: { color: colors.danger, marginBottom: 8 },

  kpiGrid: { flexDirection: 'row', flexWrap: 'wrap', justifyContent: 'space-between' },
  kpi: {
    width: '48%',
    backgroundColor: colors.surface,
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: 14,
    padding: 16,
    marginBottom: 12,
  },
  kpiHighlight: { backgroundColor: colors.primary, borderColor: colors.primary },
  kpiValue: { fontSize: 20, fontWeight: '800', color: colors.text },
  kpiValueHighlight: { color: '#fff' },
  kpiLabel: { fontSize: 12, color: colors.textMuted, marginTop: 4 },
  kpiLabelHighlight: { color: 'rgba(255,255,255,0.85)' },

  alert: {
    backgroundColor: '#fef3c7',
    borderRadius: 12,
    padding: 14,
    marginBottom: 12,
  },
  alertText: { color: '#92400e', fontWeight: '600' },

  card: {
    backgroundColor: colors.surface,
    borderRadius: 14,
    padding: 18,
    borderWidth: 1,
    borderColor: colors.border,
    marginBottom: 12,
  },
  cardTitle: { fontSize: 16, fontWeight: '700', color: colors.text, marginBottom: 12 },

  chart: { flexDirection: 'row', alignItems: 'flex-end', height: 140 },
  chartCol: { flex: 1, alignItems: 'center' },
  barTrack: { flex: 1, width: 18, justifyContent: 'flex-end' },
  bar: { width: 18, backgroundColor: colors.primary, borderRadius: 4, minHeight: 2 },
  chartLabel: { fontSize: 10, color: colors.textMuted, marginTop: 6 },

  topRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 6,
    borderBottomWidth: StyleSheet.hairlineWidth,
    borderBottomColor: colors.border,
  },
  topRank: {
    width: 22,
    fontWeight: '800',
    color: colors.primary,
  },
  topName: { flex: 1, color: colors.text, fontWeight: '600' },
  topQty: { color: colors.textMuted, fontSize: 12 },

  smallBtn: {
    alignSelf: 'flex-start',
    backgroundColor: colors.primary,
    paddingHorizontal: 16,
    paddingVertical: 10,
    borderRadius: 8,
    marginTop: 4,
  },
  smallBtnText: { color: '#fff', fontWeight: '600' },
  signOut: { marginTop: 24, alignItems: 'center', paddingVertical: 12 },
  signOutText: { color: colors.danger, fontWeight: '700' },
});
