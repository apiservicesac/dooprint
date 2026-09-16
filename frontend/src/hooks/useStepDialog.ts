import { useCallback, useEffect, useRef, useState } from "react";

export function useStepDialog() {
  const [currentStep, setCurrentStep] = useState(0);
  const transitioning = useRef(false);
  const clickTimeout = useRef<number | null>(null);

  useEffect(
    () => () => {
      if (clickTimeout.current) clearTimeout(clickTimeout.current);
    },
    [],
  );

  const onNextStep = useCallback(() => {
    transitioning.current = true;
    if (clickTimeout.current) clearTimeout(clickTimeout.current);
    clickTimeout.current = window.setTimeout(() => {
      transitioning.current = false;
    }, 600);
  }, []);

  const next = useCallback(
    (stepsLength: number) => {
      if (transitioning.current) {
        return;
      }

      if (currentStep < stepsLength - 1) {
        setCurrentStep(currentStep + 1);
        onNextStep();
      }
    },
    [currentStep, onNextStep],
  );

  const back = useCallback(() => {
    if (transitioning.current) {
      return;
    }

    if (currentStep <= 0) {
      return;
    }

    setCurrentStep(currentStep - 1);
    onNextStep();
  }, [currentStep, onNextStep]);

  return { currentStep, next, back, setCurrentStep };
}
