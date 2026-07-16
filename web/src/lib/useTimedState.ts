import { useCallback, useEffect, useRef, useState } from "react";

export function useTimedState<T>(initialValue: T) {
  const [value, setValue] = useState(initialValue);
  const timerRef = useRef<number | undefined>(undefined);

  const cancelReset = useCallback(() => {
    if (timerRef.current === undefined) return;
    window.clearTimeout(timerRef.current);
    timerRef.current = undefined;
  }, []);

  const reset = useCallback(() => {
    cancelReset();
    setValue(initialValue);
  }, [cancelReset, initialValue]);

  const setTemporarily = useCallback((nextValue: T, duration = 1200) => {
    cancelReset();
    setValue(nextValue);
    timerRef.current = window.setTimeout(() => {
      timerRef.current = undefined;
      setValue(initialValue);
    }, duration);
  }, [cancelReset, initialValue]);

  useEffect(() => cancelReset, [cancelReset]);

  return [value, setTemporarily, reset] as const;
}
