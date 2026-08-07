import { useEffect, useRef } from "react";

export function useViewHeadingFocus<T extends HTMLElement>(ready = true) {
  const heading = useRef<T>(null);

  useEffect(() => {
    if (ready) heading.current?.focus();
  }, [ready]);

  return heading;
}
