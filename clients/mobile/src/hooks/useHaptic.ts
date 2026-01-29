import { useSelector } from 'react-redux';
import ReactNativeHapticFeedback from 'react-native-haptic-feedback';
import type { RootState } from '@/store';

type HapticType = 'impactLight' | 'impactMedium' | 'impactHeavy' | 'notificationSuccess' | 'notificationWarning' | 'notificationError';

export const useHaptic = () => {
  const hapticEnabled = useSelector((state: RootState) => state.ui.hapticEnabled);

  const trigger = (type: HapticType = 'impactLight') => {
    if (hapticEnabled) {
      ReactNativeHapticFeedback.trigger(type, {
        enableVibrateFallback: true,
        ignoreAndroidSystemSettings: false,
      });
    }
  };

  return { trigger };
};
