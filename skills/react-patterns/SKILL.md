---
name: react-patterns
description: React + TypeScript development patterns for building maintainable, performant web applications with TanStack Query, Zustand, Zod, and Tailwind CSS.
---

# React Development Patterns

Modern React patterns with TypeScript for building robust, maintainable web applications.

## When to Activate

- Building React components with TypeScript
- Managing server state with TanStack Query
- Managing client state with Zustand
- Validating data with Zod schemas
- Implementing forms, routing, and authentication flows
- Optimizing React performance

## Project Structure

```
web/
├── src/
│   ├── api/                # API client and type definitions
│   │   ├── client.ts       # Axios/fetch wrapper with interceptors
│   │   ├── types.ts        # API response types (generated from OpenAPI)
│   │   └── endpoints/      # Per-resource API functions
│   ├── components/
│   │   ├── common/         # Reusable UI components (Button, Input, Modal)
│   │   ├── layout/         # App shell, Sidebar, Header
│   │   └── features/       # Feature-specific components
│   ├── hooks/              # Custom hooks
│   ├── pages/              # Page components (route-level)
│   ├── store/              # Zustand stores
│   ├── lib/                # Utilities, constants, helpers
│   ├── App.tsx
│   └── main.tsx
├── package.json
├── tsconfig.json
├── vite.config.ts
└── tailwind.config.ts
```

## API Client Pattern

### Type-Safe API Client with Interceptors

```typescript
import axios, { AxiosInstance, AxiosError } from 'axios'

interface ApiError {
  code: string
  message: string
  details?: Array<{ field: string; message: string }>
}

interface ApiResponse<T> {
  data: T
}

interface ApiListResponse<T> {
  data: T[]
  meta: { next_cursor?: string; has_next: boolean }
}

function createApiClient(baseURL: string): AxiosInstance {
  const client = axios.create({ baseURL })

  client.interceptors.request.use((config) => {
    const token = localStorage.getItem('auth_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  })

  client.interceptors.response.use(
    (response) => response,
    (error: AxiosError<{ error: ApiError }>) => {
      if (error.response?.status === 401) {
        localStorage.removeItem('auth_token')
        window.location.href = '/login'
      }
      return Promise.reject(error)
    }
  )

  return client
}

export const api = createApiClient('/api/v1')
```

## TanStack Query Patterns

### Query Key Factory

```typescript
export const queryKeys = {
  projects: {
    all: ['projects'] as const,
    detail: (id: string) => ['projects', id] as const,
  },
  issues: {
    all: (projectId: string) => ['projects', projectId, 'issues'] as const,
    detail: (projectId: string, issueId: string) =>
      ['projects', projectId, 'issues', issueId] as const,
  },
  notifications: {
    all: ['notifications'] as const,
    unreadCount: ['notifications', 'unread-count'] as const,
  },
} as const
```

### Infinite Query for Cursor Pagination

```typescript
export function useIssues(projectId: string) {
  return useInfiniteQuery({
    queryKey: queryKeys.issues.all(projectId),
    queryFn: ({ pageParam }) => issuesApi.list(projectId, pageParam),
    getNextPageParam: (lastPage) =>
      lastPage.meta.has_next ? lastPage.meta.next_cursor : undefined,
    initialPageParam: undefined as string | undefined,
  })
}
```

### Mutation with Optimistic Update

```typescript
export function useCloseIssue(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (issueId: string) => issuesApi.close(projectId, issueId),
    onMutate: async (issueId) => {
      await queryClient.cancelQueries({
        queryKey: queryKeys.issues.detail(projectId, issueId),
      })
      const previous = queryClient.getQueryData(
        queryKeys.issues.detail(projectId, issueId)
      )
      queryClient.setQueryData(
        queryKeys.issues.detail(projectId, issueId),
        (old: ApiResponse<Issue> | undefined) =>
          old ? { data: { ...old.data, status: 'closed' as const } } : old
      )
      return { previous }
    },
    onError: (_err, issueId, context) => {
      if (context?.previous) {
        queryClient.setQueryData(
          queryKeys.issues.detail(projectId, issueId),
          context.previous
        )
      }
    },
    onSettled: (_data, _error, issueId) => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.issues.detail(projectId, issueId),
      })
    },
  })
}
```

## Zustand State Management

```typescript
import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface AuthState {
  user: User | null
  token: string | null
  setAuth: (user: User, token: string) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      setAuth: (user, token) => {
        localStorage.setItem('auth_token', token)
        set({ user, token })
      },
      logout: () => {
        localStorage.removeItem('auth_token')
        set({ user: null, token: null })
      },
    }),
    { name: 'auth-storage' }
  )
)

// Select only what you need to prevent unnecessary re-renders
const isAuthenticated = useAuthStore((state) => state.user !== null)
```

## Zod Validation

```typescript
import { z } from 'zod'

export const createIssueSchema = z.object({
  title: z.string().min(1, 'Title is required').max(200),
  body: z.string().max(50000).optional(),
})

export type CreateIssueInput = z.infer<typeof createIssueSchema>
```

## Routing with Auth Guard

```typescript
import { Navigate, Outlet } from 'react-router-dom'

export function AuthGuard() {
  const isAuthenticated = useAuthStore((state) => state.user !== null)
  if (!isAuthenticated) return <Navigate to="/login" replace />
  return <Outlet />
}
```

## Anti-Patterns to Avoid

### State Management

```typescript
// BAD: Storing server state in useState/Zustand
const [issues, setIssues] = useState<Issue[]>([])
useEffect(() => { fetch('/api/issues').then(r => r.json()).then(setIssues) }, [])

// GOOD: Use TanStack Query for server state
const { data: issues } = useIssues(projectId)
```

### useEffect Pitfalls

```typescript
// BAD: useEffect for data fetching (race conditions, no caching)
useEffect(() => {
  fetch(`/api/issues/${id}`).then(r => r.json()).then(setIssue)
}, [id])

// GOOD: Use TanStack Query
const { data: issue } = useIssue(projectId, id)

// BAD: Synchronizing derived state
const [filteredItems, setFilteredItems] = useState<Item[]>([])
useEffect(() => { setFilteredItems(items.filter(i => i.active)) }, [items])

// GOOD: Derive with useMemo
const filteredItems = useMemo(() => items.filter(i => i.active), [items])
```

### Performance

```typescript
// BAD: New objects/functions in render without memoization
<Child style={{ color: 'red' }} onClick={() => doSomething()} />

// GOOD: Stable references
const handleClick = useCallback(() => doSomething(), [])
<Child style={staticStyle} onClick={handleClick} />

// BAD: Index as key for dynamic lists
{items.map((item, i) => <Item key={i} {...item} />)}

// GOOD: Stable unique IDs
{items.map((item) => <Item key={item.id} {...item} />)}
```

## Testing with React Testing Library

```typescript
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>
  )
}

test('creates issue on form submit', async () => {
  const user = userEvent.setup()
  renderWithProviders(<CreateIssueForm projectId="1" onSuccess={vi.fn()} />)

  await user.type(screen.getByLabelText('Title'), 'Fix login bug')
  await user.click(screen.getByRole('button', { name: 'Create Issue' }))

  await waitFor(() => {
    expect(screen.queryByText('Creating...')).not.toBeInTheDocument()
  })
})
```

## Quick Reference

| Pattern | When to Use |
|---------|-------------|
| TanStack Query | All server state (API data) |
| Zustand | Client-only shared state (auth, UI prefs) |
| Zod | Input validation (forms, API responses) |
| useMemo | Expensive derived computations |
| useCallback | Stable function refs passed to children |
| React.memo | Pure components with stable props |
| Suspense + lazy | Code splitting for route components |
