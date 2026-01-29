# Trading Engine Mobile App

Production-ready React Native trading application for iOS and Android.

## Features

### Core Trading
- Real-time price updates via WebSocket
- Market, limit, and stop orders
- Position management with SL/TP
- Live charts and technical analysis
- Multi-symbol watchlist
- Trade history and analytics

### Security
- Biometric authentication (Face ID, Touch ID, Fingerprint)
- Secure token storage using Keychain
- SSL pinning
- Root/jailbreak detection
- Session timeout and auto-logout
- Two-factor authentication support

### User Experience
- Dark mode support
- Haptic feedback
- Pull-to-refresh
- Offline mode with cached data
- Push notifications for order fills and alerts
- Deep linking support
- Multi-language support (i18n ready)
- Tablet optimization

### Account Management
- Deposits and withdrawals
- Real-time account balance
- Margin level monitoring
- KYC verification status
- Settings and preferences

## Tech Stack

- **React Native** 0.73+
- **TypeScript** for type safety
- **Redux Toolkit** + **RTK Query** for state management
- **React Navigation** for routing
- **WebSocket** for real-time data
- **Firebase** for push notifications
- **Keychain** for secure storage
- **Biometric** authentication
- **Reanimated** for smooth animations

## Prerequisites

- Node.js 18+
- React Native CLI
- Xcode 14+ (for iOS)
- Android Studio (for Android)
- CocoaPods (for iOS)

## Installation

```bash
# Install dependencies
npm install

# iOS - Install pods
cd ios && pod install && cd ..

# Android - No additional steps needed
```

## Development

```bash
# Start Metro bundler
npm start

# Run on iOS
npm run ios

# Run on Android
npm run android

# Type checking
npm run typecheck

# Linting
npm run lint
```

## Environment Setup

Create a `.env` file in the root directory:

```env
API_BASE_URL=https://api.tradingengine.com
WS_URL=wss://api.tradingengine.com/ws
FIREBASE_API_KEY=your_firebase_api_key
FIREBASE_PROJECT_ID=your_firebase_project_id
```

## Building for Production

### iOS

```bash
# Build release
npm run build:ios

# Or use Xcode
# Open ios/TradingEngine.xcworkspace
# Product > Archive
```

### Android

```bash
# Build APK
npm run build:android

# Output: android/app/build/outputs/apk/release/app-release.apk
```

## App Store Submission

### iOS (App Store)

1. Update version in `ios/TradingEngine/Info.plist`
2. Update build number
3. Archive in Xcode
4. Upload to App Store Connect
5. Submit for review

Required assets:
- App icon (1024x1024)
- Screenshots (all device sizes)
- Privacy policy URL
- App description

### Android (Google Play)

1. Update version in `android/app/build.gradle`
2. Build signed APK/AAB
3. Upload to Google Play Console
4. Submit for review

Required assets:
- App icon (512x512)
- Feature graphic (1024x500)
- Screenshots (phone and tablet)
- Privacy policy URL
- App description

## Configuration

### Push Notifications

1. Set up Firebase project
2. Add `google-services.json` (Android) to `android/app/`
3. Add `GoogleService-Info.plist` (iOS) to `ios/TradingEngine/`
4. Configure APNs certificates for iOS

### Deep Linking

Configure URL schemes in:
- iOS: `ios/TradingEngine/Info.plist`
- Android: `android/app/src/main/AndroidManifest.xml`

Example: `tradingengine://chart/BTCUSD`

### SSL Pinning

Update public key hashes in `src/utils/security.ts`:

```typescript
export const sslPinning = {
  publicKeyHashes: [
    'sha256/YOUR_PUBLIC_KEY_HASH_HERE',
  ],
};
```

## Code Push (OTA Updates)

```bash
# Install CodePush
npm install -g appcenter-cli

# Login
appcenter login

# Deploy to Android
npm run codepush:android

# Deploy to iOS
npm run codepush:ios
```

## Performance Optimization

- Images are optimized and lazy-loaded
- WebSocket connection pooling
- Redux state persistence for offline support
- Efficient FlatList rendering
- Code splitting for faster startup

## Testing

```bash
# Run tests
npm test

# Run with coverage
npm test -- --coverage

# E2E tests (if configured)
npm run e2e
```

## Troubleshooting

### iOS Build Issues

```bash
cd ios
pod deintegrate
pod install
cd ..
```

### Android Build Issues

```bash
cd android
./gradlew clean
cd ..
```

### Metro Bundler Issues

```bash
npm start -- --reset-cache
```

## Project Structure

```
src/
├── assets/          # Images, fonts, etc.
├── components/      # Reusable components
├── hooks/          # Custom React hooks
├── navigation/     # Navigation setup
├── screens/        # Screen components
├── services/       # API, WebSocket, notifications
├── store/          # Redux store and slices
├── theme/          # Colors, typography, spacing
├── types/          # TypeScript types
└── utils/          # Helper functions
```

## Security Best Practices

1. Never commit API keys or secrets
2. Use SSL pinning for production
3. Implement certificate transparency
4. Enable root/jailbreak detection
5. Use secure storage for tokens
6. Implement session timeout
7. Add rate limiting for API calls

## Privacy & Compliance

- GDPR compliant
- App tracking transparency (iOS 14+)
- Privacy policy included
- User data encryption
- Secure data deletion

## Support

For issues and questions:
- Email: support@tradingengine.com
- Documentation: https://docs.tradingengine.com
- GitHub Issues: https://github.com/your-org/trading-engine-mobile

## License

Proprietary - All rights reserved

---

**Built with React Native for iOS and Android**
