# 03 — Mobile Security

For Android and iOS apps: native (Kotlin/Java, Swift), React Native, Flutter, and hybrid apps. Maps to **OWASP MASVS v2**, tested with **OWASP MASTG**, and the **OWASP Mobile Top 10**.

---

## MASVS Control Groups (overview)

| Group | Focus |
|---|---|
| MASVS-STORAGE | Secure storage of sensitive data |
| MASVS-CRYPTO | Correct cryptography usage |
| MASVS-AUTH | Authentication and authorization |
| MASVS-NETWORK | Secure network communication |
| MASVS-PLATFORM | Secure interaction with the platform and other apps |
| MASVS-CODE | Code quality, updates, dependencies |
| MASVS-RESILIENCE | Anti-tampering, anti-reverse-engineering |
| MASVS-PRIVACY | User privacy and data minimization |

---

## 1. Secure Local Storage (MASVS-STORAGE)

- [ ] `[CRITICAL]` Tokens, keys, and credentials stored in **iOS Keychain** / **Android Keystore** (via EncryptedSharedPreferences or DataStore + Tink)
- [ ] `[CRITICAL]` No secrets in `SharedPreferences`, `NSUserDefaults`, `AsyncStorage`, plain SQLite, or files
- [ ] `[HIGH]` React Native: use `react-native-keychain` or `expo-secure-store`, not AsyncStorage
- [ ] `[HIGH]` Flutter: use `flutter_secure_storage`
- [ ] `[HIGH]` Local databases with sensitive data encrypted (SQLCipher, Realm encryption)
- [ ] `[HIGH]` Sensitive data excluded from device backups (`android:allowBackup="false"` or backup rules; iOS `isExcludedFromBackup`)
- [ ] `[MEDIUM]` Data wiped on logout
- [ ] `[MEDIUM]` Keyboard cache disabled on sensitive fields (`secureTextEntry`, `textNoSuggestions`)
- [ ] `[MEDIUM]` No sensitive data in logs (`Log.d`, `print`, `console.log`) in release builds

**Test:** On a rooted/jailbroken device or emulator, inspect `/data/data/<package>/` and the iOS app sandbox for plaintext data.

---

## 2. Cryptography (MASVS-CRYPTO)

- [ ] `[HIGH]` Platform crypto APIs used; no custom crypto
- [ ] `[HIGH]` No hardcoded encryption keys
- [ ] `[HIGH]` AES-GCM, no ECB mode; secure random IVs
- [ ] `[MEDIUM]` Keys hardware-backed where available (StrongBox, Secure Enclave)

---

## 3. Authentication & Authorization (MASVS-AUTH)

- [ ] `[CRITICAL]` All authorization enforced server-side
- [ ] `[HIGH]` OAuth via system browser (ASWebAuthenticationSession / Custom Tabs) with **PKCE**, never embedded WebView login
- [ ] `[HIGH]` Biometric auth tied to a Keystore/Keychain key (crypto-bound), not just a boolean "success" callback
- [ ] `[HIGH]` Biometric enrollment changes invalidate keys (`setInvalidatedByBiometricEnrollment`, `.biometryCurrentSet`)
- [ ] `[HIGH]` Session timeout and remote logout supported
- [ ] `[MEDIUM]` Step-up auth for sensitive actions (transfers, profile changes)
- [ ] `[MEDIUM]` Passkeys / FIDO2 supported where feasible

---

## 4. Network Security (MASVS-NETWORK)

- [ ] `[CRITICAL]` HTTPS only; cleartext traffic disabled (`android:usesCleartextTraffic="false"`, iOS ATS enabled with no blanket exceptions)
- [ ] `[HIGH]` **Certificate / public key pinning** for high-risk apps (finance, health), with backup pins and a rotation plan
- [ ] `[HIGH]` Android Network Security Config defined
- [ ] `[HIGH]` No trust-all `TrustManager` or `HostnameVerifier`; no `NSAllowsArbitraryLoads`
- [ ] `[MEDIUM]` Pinning bypass resilience tested (Frida, objection)

**Test:** Proxy traffic through Burp/mitmproxy with a user-installed CA; app should refuse connections (if pinned) and never send data over HTTP.

---

## 5. Platform Interaction (MASVS-PLATFORM)

### Deep links & intents
- [ ] `[HIGH]` Deep link parameters validated as untrusted input
- [ ] `[HIGH]` **Android App Links** / **iOS Universal Links** verified (assetlinks.json, apple-app-site-association) instead of custom schemes for sensitive flows
- [ ] `[HIGH]` Exported Android components (activities, services, receivers, providers) minimized; `android:exported` explicit
- [ ] `[HIGH]` Content providers protected with permissions; no path traversal
- [ ] `[MEDIUM]` PendingIntents use `FLAG_IMMUTABLE`
- [ ] `[MEDIUM]` Implicit intents not used to send sensitive data

### WebView
- [ ] `[HIGH]` JavaScript disabled unless required
- [ ] `[HIGH]` `addJavascriptInterface` / JS bridges expose minimal methods and only to trusted origins
- [ ] `[HIGH]` File access disabled (`setAllowFileAccess(false)`, `setAllowUniversalAccessFromFileURLs(false)`)
- [ ] `[MEDIUM]` Only allow-listed URLs loaded
- [ ] `[MEDIUM]` React Native WebView: `originWhitelist` restricted

### UI data leaks
- [ ] `[HIGH]` Screenshots/screen recording blocked on sensitive screens (`FLAG_SECURE` on Android; iOS overlay on capture)
- [ ] `[HIGH]` App switcher snapshot obscured on iOS and Android
- [ ] `[MEDIUM]` Clipboard: sensitive data not copied, or cleared after timeout; avoid reading clipboard unnecessarily
- [ ] `[MEDIUM]` Overlay/tapjacking protection (`filterTouchesWhenObscured`)

### Permissions
- [ ] `[HIGH]` Only necessary permissions requested; runtime permissions requested in context
- [ ] `[MEDIUM]` Permission purpose strings clear (iOS `Info.plist` usage descriptions)
- [ ] `[MEDIUM]` Unused permissions removed (including those added by SDKs; check merged manifest)

### Push notifications
- [ ] `[MEDIUM]` No sensitive data (OTPs, balances, health info) in notification payloads or lock-screen previews
- [ ] `[MEDIUM]` Push tokens bound to authenticated user and removed on logout

---

## 6. Code Quality & Updates (MASVS-CODE)

- [ ] `[CRITICAL]` No hardcoded API secrets in APK/IPA (assume all bundled strings are public)
- [ ] `[HIGH]` Release builds: `debuggable=false`, logging stripped, dev menus disabled
- [ ] `[HIGH]` Minimum supported OS version receives security patches
- [ ] `[HIGH]` Forced update mechanism for critical security fixes
- [ ] `[HIGH]` Third-party SDKs audited (analytics, ads, crash reporting, attribution): data collected, permissions, reputation
- [ ] `[MEDIUM]` OTA updates (CodePush, Expo Updates, Shorebird) signed and served over HTTPS
- [ ] `[MEDIUM]` Input from IPC, files, QR codes, NFC validated

---

## 7. Resilience (MASVS-RESILIENCE)

> Resilience controls slow attackers down; they are not a substitute for server-side security. Prioritize for finance, health, gaming, and DRM apps.

- [ ] `[MEDIUM]` Code obfuscation (R8/ProGuard on Android; Swift obfuscation tools; Hermes bytecode for RN; Flutter `--obfuscate --split-debug-info`)
- [ ] `[MEDIUM]` Root / jailbreak detection with appropriate response
- [ ] `[MEDIUM]` Emulator, debugger, and hooking framework (Frida, Xposed) detection
- [ ] `[MEDIUM]` Integrity checks: tamper detection, signature verification
- [ ] `[MEDIUM]` **Play Integrity API** / **App Attest (DeviceCheck)** verified server-side
- [ ] `[LOW]` RASP (runtime application self-protection) for high-risk apps

**Test:** Decompile with jadx/apktool (Android) or class-dump/Hopper (iOS); check what logic and strings are exposed.

---

## 8. Privacy (MASVS-PRIVACY)

- [ ] `[HIGH]` Data minimization: collect only what features need
- [ ] `[HIGH]` Privacy policy accessible in-app and in store listing
- [ ] `[HIGH]` **Apple Privacy Nutrition Labels** and **Privacy Manifest** (`PrivacyInfo.xcprivacy`) accurate, including required-reason APIs
- [ ] `[HIGH]` **Google Play Data Safety** form accurate
- [ ] `[HIGH]` **App Tracking Transparency** prompt before cross-app tracking on iOS
- [ ] `[HIGH]` In-app account deletion available (required by Apple and Google)
- [ ] `[MEDIUM]` Location: approximate where precise isn't needed
- [ ] `[MEDIUM]` Advertising ID usage disclosed and respects user opt-out

---

## 9. Signing & Distribution

- [ ] `[CRITICAL]` Signing keys stored securely (Play App Signing; keystore backed up and access-restricted)
- [ ] `[HIGH]` iOS certificates and provisioning profiles managed (fastlane match or equivalent), access limited
- [ ] `[MEDIUM]` APK Signature Scheme v2+ used
- [ ] `[MEDIUM]` Sideloaded/modified builds detected server-side where relevant

---

## 10. OWASP Mobile Top 10 Cross-Check

- [ ] M1 — Improper Credential Usage
- [ ] M2 — Inadequate Supply Chain Security
- [ ] M3 — Insecure Authentication/Authorization
- [ ] M4 — Insufficient Input/Output Validation
- [ ] M5 — Insecure Communication
- [ ] M6 — Inadequate Privacy Controls
- [ ] M7 — Insufficient Binary Protections
- [ ] M8 — Security Misconfiguration
- [ ] M9 — Insecure Data Storage
- [ ] M10 — Insufficient Cryptography

---

## Tools Summary

| Purpose | Tools |
|---|---|
| Automated static + dynamic | MobSF |
| Decompile Android | jadx, apktool, dex2jar |
| Decompile iOS | Hopper, Ghidra, class-dump |
| Dynamic instrumentation | Frida, objection |
| Traffic interception | Burp Suite, mitmproxy, Charles |
| Static analysis | Semgrep, CodeQL, Android Lint, SwiftLint |
| Device testing | Rooted Android emulator, Corellium, jailbroken test device |

## References
- OWASP MASVS v2, OWASP MASTG, OWASP Mobile Top 10
- Apple App Store Review Guidelines (section 5 — Legal / Privacy)
- Google Play Developer Policy Center (User Data, Permissions)
