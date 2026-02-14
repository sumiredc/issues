---
name: react-native-patterns
description: React Native + Expo patterns for cross-platform mobile apps with React Navigation, TanStack Query, and platform-specific optimizations.
---

# React Native Development Patterns

Modern React Native patterns with Expo for performant cross-platform mobile applications.

## When to Activate

- Building React Native screens and components
- Configuring React Navigation
- Implementing OAuth in mobile apps
- Handling platform-specific behavior
- Optimizing FlatList performance
- Managing push notifications

## Project Structure

```
mobile/
├── src/
│   ├── api/                # API client (shared types with web)
│   ├── components/
│   │   ├── common/         # Button, TextInput, Card
│   │   └── features/       # Feature-specific components
│   ├── hooks/
│   ├── navigation/
│   │   ├── AppNavigator.tsx
│   │   └── types.ts        # Navigation param types
│   ├── screens/
│   ├── store/              # Zustand (same pattern as web)
│   ├── lib/
│   └── theme/
├── app.json
├── App.tsx
└── package.json
```

## Type-Safe Navigation

```typescript
import type { NativeStackScreenProps } from '@react-navigation/native-stack'

export type RootStackParamList = {
  Login: undefined
  ProjectList: undefined
  ProjectDetail: { projectId: string }
  IssueDetail: { projectId: string; issueId: string }
  CreateIssue: { projectId: string }
  Notifications: undefined
}

export type ScreenProps<T extends keyof RootStackParamList> =
  NativeStackScreenProps<RootStackParamList, T>
```

### Navigator with Auth Flow

```typescript
import { createNativeStackNavigator } from '@react-navigation/native-stack'
import { useAuthStore } from '../store/auth'

const Stack = createNativeStackNavigator<RootStackParamList>()

export function AppNavigator() {
  const isAuthenticated = useAuthStore((state) => state.token !== null)

  return (
    <Stack.Navigator>
      {isAuthenticated ? (
        <>
          <Stack.Screen name="ProjectList" component={ProjectListScreen} />
          <Stack.Screen name="ProjectDetail" component={ProjectDetailScreen} />
          <Stack.Screen name="IssueDetail" component={IssueDetailScreen} />
          <Stack.Screen
            name="CreateIssue"
            component={CreateIssueScreen}
            options={{ presentation: 'modal' }}
          />
        </>
      ) : (
        <Stack.Screen name="Login" component={LoginScreen}
          options={{ headerShown: false }} />
      )}
    </Stack.Navigator>
  )
}
```

## OAuth with expo-auth-session

```typescript
import * as AuthSession from 'expo-auth-session'
import * as WebBrowser from 'expo-web-browser'

WebBrowser.maybeCompleteAuthSession()

export function useGoogleAuth() {
  const setAuth = useAuthStore((state) => state.setAuth)

  const [request, , promptAsync] = AuthSession.useAuthRequest(
    {
      clientId: process.env.EXPO_PUBLIC_GOOGLE_CLIENT_ID!,
      redirectUri: AuthSession.makeRedirectUri({ scheme: 'issues' }),
      scopes: ['openid', 'profile', 'email'],
      responseType: AuthSession.ResponseType.Code,
    },
    { authorizationEndpoint: 'https://accounts.google.com/o/oauth2/v2/auth' }
  )

  const handleSignIn = async () => {
    const result = await promptAsync()
    if (result.type === 'success' && result.params.code) {
      const { data } = await api.post('/auth/google/callback', {
        code: result.params.code,
        redirect_uri: AuthSession.makeRedirectUri({ scheme: 'issues' }),
      })
      setAuth(data.data.user, data.data.token)
    }
  }

  return { handleSignIn, isReady: !!request }
}
```

## Secure Token Storage

```typescript
import * as SecureStore from 'expo-secure-store'
import { create } from 'zustand'
import { createJSONStorage, persist } from 'zustand/middleware'

const secureStoreAdapter = {
  getItem: (key: string) => SecureStore.getItemAsync(key),
  setItem: (key: string, value: string) => SecureStore.setItemAsync(key, value),
  removeItem: (key: string) => SecureStore.deleteItemAsync(key),
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      setAuth: (user, token) => set({ user, token }),
      logout: () => set({ user: null, token: null }),
    }),
    {
      name: 'auth-storage',
      storage: createJSONStorage(() => secureStoreAdapter),
    }
  )
)
```

## Optimized FlatList

```typescript
import { FlatList, ActivityIndicator } from 'react-native'
import { useCallback } from 'react'

export function IssueList({ projectId }: { projectId: string }) {
  const { data, fetchNextPage, hasNextPage, isFetchingNextPage, isLoading } =
    useIssues(projectId)

  const issues = data?.pages.flatMap((page) => page.data) ?? []

  const renderItem = useCallback(
    ({ item }: { item: Issue }) => <IssueCard issue={item} />,
    []
  )

  const keyExtractor = useCallback(
    (item: Issue) => item.id.toString(),
    []
  )

  const handleEndReached = useCallback(() => {
    if (hasNextPage && !isFetchingNextPage) fetchNextPage()
  }, [hasNextPage, isFetchingNextPage, fetchNextPage])

  if (isLoading) return <ActivityIndicator size="large" />

  return (
    <FlatList
      data={issues}
      renderItem={renderItem}
      keyExtractor={keyExtractor}
      onEndReached={handleEndReached}
      onEndReachedThreshold={0.5}
      removeClippedSubviews={true}
      maxToRenderPerBatch={10}
      windowSize={5}
      initialNumToRender={10}
      getItemLayout={(_data, index) => ({
        length: ITEM_HEIGHT,
        offset: ITEM_HEIGHT * index,
        index,
      })}
    />
  )
}

const ITEM_HEIGHT = 80
```

### Memoized List Items

```typescript
import { memo } from 'react'
import { Pressable, Text, View, StyleSheet } from 'react-native'

export const IssueCard = memo<IssueCardProps>(({ issue, onPress }) => (
  <Pressable
    onPress={() => onPress?.(issue.id)}
    style={({ pressed }) => [styles.card, pressed && styles.cardPressed]}
  >
    <Text style={styles.title} numberOfLines={1}>{issue.title}</Text>
    <StatusBadge status={issue.status} />
  </Pressable>
))

const styles = StyleSheet.create({
  card: { padding: 16, borderBottomWidth: StyleSheet.hairlineWidth, borderBottomColor: '#e5e7eb' },
  cardPressed: { backgroundColor: '#f9fafb' },
  title: { fontSize: 16, fontWeight: '600', flex: 1 },
})
```

## Anti-Patterns to Avoid

### Navigation

```typescript
// BAD: Untyped navigation params
navigation.navigate('IssueDetail', { id: issue.id })

// GOOD: Typed params
navigation.navigate('IssueDetail', { projectId, issueId: issue.id })
```

### Performance

```typescript
// BAD: Inline renderItem (new function every render)
<FlatList renderItem={({ item }) => <Card item={item} />} />

// GOOD: Stable reference
const renderItem = useCallback(({ item }) => <Card item={item} />, [])
<FlatList renderItem={renderItem} />

// BAD: ScrollView for long lists
<ScrollView>{items.map(item => <Card key={item.id} item={item} />)}</ScrollView>

// GOOD: FlatList for any list with potential scroll
<FlatList data={items} renderItem={renderItem} keyExtractor={keyExtractor} />
```

### Styling

```typescript
// BAD: StyleSheet inside component (recreated every render)
function MyComponent() {
  const styles = StyleSheet.create({ container: { flex: 1 } })
  return <View style={styles.container} />
}

// GOOD: StyleSheet outside component
function MyComponent() {
  return <View style={styles.container} />
}
const styles = StyleSheet.create({ container: { flex: 1 } })

// BAD: Inline style objects
<View style={{ flex: 1, padding: 16 }} />

// GOOD: Use StyleSheet
<View style={styles.container} />
```

### Security

```typescript
// BAD: AsyncStorage for tokens (not encrypted)
await AsyncStorage.setItem('auth_token', token)

// GOOD: SecureStore for sensitive data
await SecureStore.setItemAsync('auth_token', token)
```

## Testing

```typescript
import { render, screen, fireEvent } from '@testing-library/react-native'

test('renders issue card', () => {
  const issue = { id: '1', title: 'Fix bug', status: 'open' }
  render(<IssueCard issue={issue} />)
  expect(screen.getByText('Fix bug')).toBeTruthy()
})

test('calls onPress with issue id', () => {
  const onPress = jest.fn()
  render(<IssueCard issue={{ id: '1', title: 'Fix bug' }} onPress={onPress} />)
  fireEvent.press(screen.getByText('Fix bug'))
  expect(onPress).toHaveBeenCalledWith('1')
})
```

## Quick Reference

| Pattern | When to Use |
|---------|-------------|
| FlatList + getItemLayout | Fixed-height lists |
| FlatList + onEndReached | Infinite scroll / cursor pagination |
| memo + useCallback | List items, frequently re-rendered components |
| SecureStore | Auth tokens, sensitive data |
| expo-auth-session | OAuth flows (Google, GitHub) |
| expo-notifications | Push notifications |
| SafeAreaView | Screen containers |
| Platform.select | Platform-specific values |
