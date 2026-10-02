# @techshop/api-client

Typed client for the Go API (`backend/`), used by every web and mobile app.

```ts
import { createApiClient } from "@techshop/api-client";

const api = createApiClient({ baseUrl: process.env.API_URL! });
const health = await api.health();
```

Add response types to `src/types.ts` and endpoint helpers to `src/client.ts` as the API grows. Uses only `fetch`, so it runs in Node, browsers, and React Native.
