# Deployment Guide

## Pre-Deployment Checklist

### Code Quality
- [ ] All TypeScript errors resolved
- [ ] All ESLint warnings fixed
- [ ] Tests passing (unit and integration)
- [ ] No console.log statements in production code
- [ ] Environment variables configured correctly

### Security
- [ ] API keys moved to environment variables
- [ ] SSL pinning configured with correct certificates
- [ ] Root/jailbreak detection enabled
- [ ] Biometric authentication tested
- [ ] Secure storage implementation verified

### Performance
- [ ] Images optimized
- [ ] Bundle size < 50MB
- [ ] App startup time < 2 seconds
- [ ] Memory leaks checked
- [ ] WebSocket reconnection tested

### Assets
- [ ] App icon (1024x1024 for iOS, 512x512 for Android)
- [ ] Splash screen for all device sizes
- [ ] Screenshots for App Store/Play Store
- [ ] Privacy policy URL
- [ ] Terms of service URL

## iOS Deployment

### 1. Update Version

Edit `ios/TradingEngine/Info.plist`:

```xml
<key>CFBundleShortVersionString</key>
<string>1.0.0</string>
<key>CFBundleVersion</key>
<string>1</string>
```

### 2. Configure Signing

1. Open `ios/TradingEngine.xcworkspace` in Xcode
2. Select project > Signing & Capabilities
3. Select your team
4. Ensure "Automatically manage signing" is checked

### 3. Build Archive

```bash
# In Xcode: Product > Archive
# Or via command line:
xcodebuild -workspace ios/TradingEngine.xcworkspace \
  -scheme TradingEngine \
  -configuration Release \
  -archivePath build/TradingEngine.xcarchive \
  archive
```

### 4. Upload to App Store Connect

1. Window > Organizer
2. Select archive
3. Click "Distribute App"
4. Select "App Store Connect"
5. Upload

### 5. Submit for Review

1. Log into App Store Connect
2. Select your app
3. Add app information:
   - Name
   - Subtitle
   - Privacy policy URL
   - Category
   - Screenshots
   - Description
   - Keywords
4. Submit for review

## Android Deployment

### 1. Update Version

Edit `android/app/build.gradle`:

```gradle
defaultConfig {
    versionCode 1
    versionName "1.0.0"
}
```

### 2. Generate Signing Key

```bash
keytool -genkeypair -v -storetype PKCS12 \
  -keystore tradingengine.keystore \
  -alias tradingengine \
  -keyalg RSA \
  -keysize 2048 \
  -validity 10000
```

### 3. Configure Signing

Create `android/gradle.properties`:

```properties
MYAPP_RELEASE_STORE_FILE=tradingengine.keystore
MYAPP_RELEASE_KEY_ALIAS=tradingengine
MYAPP_RELEASE_STORE_PASSWORD=your_password
MYAPP_RELEASE_KEY_PASSWORD=your_password
```

### 4. Build Release APK/AAB

```bash
# Build APK
cd android
./gradlew assembleRelease

# Build AAB (recommended for Play Store)
./gradlew bundleRelease
```

Output:
- APK: `android/app/build/outputs/apk/release/app-release.apk`
- AAB: `android/app/build/outputs/bundle/release/app-release.aab`

### 5. Upload to Google Play Console

1. Log into Google Play Console
2. Select your app (or create new)
3. Production > Create new release
4. Upload AAB file
5. Fill in release details
6. Add store listing:
   - Title
   - Short description
   - Full description
   - Screenshots
   - Feature graphic
   - Icon
   - Privacy policy URL
7. Submit for review

## Firebase Setup

### iOS

1. Download `GoogleService-Info.plist` from Firebase Console
2. Add to `ios/TradingEngine/`
3. Ensure it's added to build target in Xcode

### Android

1. Download `google-services.json` from Firebase Console
2. Add to `android/app/`

## Push Notifications

### iOS (APNs)

1. Create APNs certificate in Apple Developer Portal
2. Upload to Firebase Console:
   - Project Settings > Cloud Messaging
   - APNs Certificates
3. Enable Push Notifications in Xcode:
   - Capabilities > Push Notifications

### Android (FCM)

Already configured via `google-services.json`

## Deep Linking

### iOS

Update `ios/TradingEngine/Info.plist`:

```xml
<key>CFBundleURLTypes</key>
<array>
  <dict>
    <key>CFBundleURLSchemes</key>
    <array>
      <string>tradingengine</string>
    </array>
  </dict>
</array>
```

### Android

Update `android/app/src/main/AndroidManifest.xml`:

```xml
<intent-filter>
  <action android:name="android.intent.action.VIEW" />
  <category android:name="android.intent.category.DEFAULT" />
  <category android:name="android.intent.category.BROWSABLE" />
  <data android:scheme="tradingengine" />
</intent-filter>
```

## CodePush (OTA Updates)

### Setup

```bash
npm install -g appcenter-cli
appcenter login
appcenter apps create -d TradingEngine-iOS -o iOS -p React-Native
appcenter apps create -d TradingEngine-Android -o Android -p React-Native
```

### Deploy

```bash
# iOS
appcenter codepush release-react -a your-org/TradingEngine-iOS -d Production

# Android
appcenter codepush release-react -a your-org/TradingEngine-Android -d Production
```

## Monitoring

### Crashlytics

1. Enable in Firebase Console
2. Add to both iOS and Android
3. Test crash reporting

### Analytics

Track key events:
- App launches
- Trades executed
- Deposits/withdrawals
- Login/logout
- Screen views

## Rollback Plan

### iOS
1. Remove from sale in App Store Connect
2. Submit previous version for expedited review

### Android
1. Deactivate release in Play Console
2. Activate previous release

### CodePush
```bash
appcenter codepush rollback -a your-org/TradingEngine-iOS Production
appcenter codepush rollback -a your-org/TradingEngine-Android Production
```

## Post-Deployment

- [ ] Monitor crash reports
- [ ] Check analytics for adoption
- [ ] Monitor API error rates
- [ ] Test on various devices
- [ ] Respond to user reviews
- [ ] Monitor performance metrics

## Support

For deployment issues:
- iOS: https://developer.apple.com/support/
- Android: https://support.google.com/googleplay/android-developer
- Firebase: https://firebase.google.com/support
