import { NavigationContainer } from '@react-navigation/native';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { useEffect } from 'react';
import { ActivityIndicator, View } from 'react-native';

import { setUnauthorizedHandler } from '../api/client';
import { useAuthStore } from '../features/auth/authStore';
import { LoginScreen } from '../screens/LoginScreen';
import { colors } from '../theme/colors';
import { AppTabs } from './AppTabs';

const Stack = createNativeStackNavigator();

/**
 * Switches between the auth flow and the app based on session state.
 * Later phases expand the authed area into role-based tab navigators
 * (admin / cashier / warehouse — PRD §15).
 */
export function RootNavigator() {
  const token = useAuthStore((s) => s.token);
  const initializing = useAuthStore((s) => s.initializing);
  const bootstrap = useAuthStore((s) => s.bootstrap);
  const signOut = useAuthStore((s) => s.signOut);

  useEffect(() => {
    bootstrap();
    // Log the user out automatically if the API rejects the token.
    setUnauthorizedHandler(() => {
      void signOut();
    });
  }, [bootstrap, signOut]);

  if (initializing) {
    return (
      <View style={{ flex: 1, justifyContent: 'center', backgroundColor: colors.background }}>
        <ActivityIndicator size="large" color={colors.primary} />
      </View>
    );
  }

  return (
    <NavigationContainer>
      <Stack.Navigator>
        {token ? (
          <Stack.Screen name="App" component={AppTabs} options={{ headerShown: false }} />
        ) : (
          <Stack.Screen
            name="Login"
            component={LoginScreen}
            options={{ headerShown: false }}
          />
        )}
      </Stack.Navigator>
    </NavigationContainer>
  );
}
