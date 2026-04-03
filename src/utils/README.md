# Mobile App React Native

A cross-platform mobile application built with React Native.

## Features

- Cross-platform support (iOS & Android)
- Modern UI with React Native Paper
- State management with Redux Toolkit
- Navigation with React Navigation
- API integration with Axios

## Prerequisites

- Node.js (v18 or higher)
- npm (v9 or higher) or yarn
- React Native CLI
- Xcode (for iOS development)
- Android Studio (for Android development)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/your-username/mobile-app-react-native.git
```

2. Install dependencies:
```bash
cd mobile-app-react-native
npm install
# or
yarn install
```

3. For iOS:
```bash
cd ios && pod install && cd ..
```

## Running the App

### Android
```bash
npm run android
# or
yarn android
```

### iOS
```bash
npm run ios
# or
yarn ios
```

## Scripts

- `start`: Start Metro bundler
- `test`: Run Jest tests
- `lint`: Run ESLint
- `build:android`: Build Android APK
- `build:ios`: Build iOS app

## Configuration

Copy `.env.example` to `.env` and update the values:
```bash
cp .env.example .env
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

MIT