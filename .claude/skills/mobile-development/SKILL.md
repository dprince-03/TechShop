---
name: mobile-development
description: Build and maintain mobile apps with React Native (Expo or bare), Flutter, or native Android/iOS — project structure, navigation, state and data fetching, offline support, secure storage, push notifications, deep links, environment config, builds and signing, OTA updates, and App Store / Google Play release requirements. Use this skill whenever the user is working on a mobile app feature, screen, build, release, or mobile-specific bug, or asks how to structure or ship a mobile app.
---

# Mobile Development

## Identify the stack first
`app.json`/`app.config.*` (Expo), `pubspec.yaml` (Flutter), `android/` + `ios/` (bare RN or native). Follow existing structure; use defaults below only when missing.

## React Native (Expo preferred for new apps)

```
app/                  # expo-router screens (file-based routing)
src/
  components/
  features/<feature>/ # screens, hooks, api, types per feature
  lib/                # api client, auth, storage, notifications
  theme/
```
- Data fetching: TanStack Query (caching, retries, offline persistence).
- State: local + React Query; Zustand for small global state.
- Forms: React Hook Form + zod.
- Secure storage: `expo-secure-store` / `react-native-keychain` for tokens — **never AsyncStorage for secrets**.
- Env: `app.config.ts` + EAS environment variables; anything bundled is public — no secrets in the app.
- Builds: EAS Build; OTA via EAS Update for JS-only changes (native changes need a store release).

## Flutter

```
lib/
  core/ (theme, router, network, storage)
  features/<feature>/{data,domain,presentation}/
```
- State: Riverpod or Bloc (match project); routing: go_router.
- Networking: dio with interceptors (auth refresh, logging in debug only).
- Secure storage: `flutter_secure_storage`.
- Release: `flutter build appbundle --obfuscate --split-debug-info=build/symbols`.

## Cross-cutting rules

- **API contract**: handle 401 → refresh token → retry once → logout. Handle offline and timeouts gracefully with retry UI.
- **Versioning**: send app version header; backend can force-update old versions (minimum supported version endpoint).
- **Offline**: decide per feature — read-only cache vs queued writes with conflict handling.
- **Push notifications**: register token after login, remove on logout; no sensitive data in payloads; deep-link to content on tap.
- **Deep links**: Universal Links / App Links for sensitive flows; validate all parameters.
- **Performance**: FlatList/ListView.builder for long lists, memoize expensive components, optimize images, avoid heavy work on the JS/UI thread.
- **Accessibility**: labels on touchables (`accessibilityLabel` / `Semantics`), dynamic text sizes, 44pt touch targets.
- **Security**: see `security-audit` skill → mobile checklist (storage, pinning, root detection, obfuscation).

## Store release checklist

- [ ] Version and build number bumped
- [ ] Release build tested on real devices (low-end Android included)
- [ ] Privacy policy URL; Apple privacy labels + `PrivacyInfo.xcprivacy`; Google Play Data Safety form accurate
- [ ] In-app account deletion available
- [ ] Permissions justified with clear usage strings
- [ ] Crash reporting (Sentry/Crashlytics) configured with symbol upload
- [ ] Screenshots, description, and release notes updated
- [ ] Staged rollout (Play) / phased release (App Store) enabled
- [ ] Signing keys backed up and access-controlled
