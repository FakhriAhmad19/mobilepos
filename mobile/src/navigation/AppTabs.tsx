import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { Text } from 'react-native';

import { DashboardScreen } from '../screens/DashboardScreen';
import { InventoryScreen } from '../screens/InventoryScreen';
import { POSScreen } from '../screens/POSScreen';
import { ProductsScreen } from '../screens/ProductsScreen';
import { TransactionsScreen } from '../screens/TransactionsScreen';
import { colors } from '../theme/colors';

const Tab = createBottomTabNavigator();

// Simple emoji tab icons keep the foundation dependency-free; swap for an
// icon set in a later phase.
function icon(glyph: string) {
  return ({ color }: { color: string }) => <Text style={{ color, fontSize: 18 }}>{glyph}</Text>;
}

/**
 * Authenticated tab area. Expands into role-specific tabs
 * (POS, Inventory, Reports) in later phases (PRD §15).
 */
export function AppTabs() {
  return (
    <Tab.Navigator
      screenOptions={{
        tabBarActiveTintColor: colors.primary,
        tabBarInactiveTintColor: colors.textMuted,
        headerStyle: { backgroundColor: colors.surface },
      }}
    >
      <Tab.Screen
        name="Dashboard"
        component={DashboardScreen}
        options={{ title: 'Beranda', tabBarIcon: icon('🏠') }}
      />
      <Tab.Screen
        name="POS"
        component={POSScreen}
        options={{ title: 'Kasir', tabBarIcon: icon('🧾') }}
      />
      <Tab.Screen
        name="Transactions"
        component={TransactionsScreen}
        options={{ title: 'Transaksi', tabBarIcon: icon('📄') }}
      />
      <Tab.Screen
        name="Products"
        component={ProductsScreen}
        options={{ title: 'Produk', tabBarIcon: icon('📦') }}
      />
      <Tab.Screen
        name="Inventory"
        component={InventoryScreen}
        options={{ title: 'Stok', tabBarIcon: icon('🏬') }}
      />
    </Tab.Navigator>
  );
}
