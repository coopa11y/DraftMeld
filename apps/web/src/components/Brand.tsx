export function Brand() {
  return (
    <span className="brand">
      <svg className="brand__mark" aria-hidden="true" viewBox="0 0 40 40">
        <path d="M3 4h25l9 8v16l-9 8H3l5-8V12L3 4Z" fill="currentColor" />
        <path d="m16 12 11 8-11 8V12Zm4 7v3l3-2-3-1Z" className="brand__cut" />
      </svg>
      <span>DraftMeld</span>
    </span>
  );
}
