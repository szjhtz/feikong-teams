export function LoadingSurface({ label }: { label: string }) {
  return (
    <div className="sketch-surface flex min-w-40 items-center justify-center gap-2 rounded-xl px-5 py-4 text-sm text-muted-foreground">
      <span className="h-1.5 w-1.5 animate-pulse rounded-full bg-primary" />
      <span className="h-1.5 w-1.5 animate-pulse rounded-full bg-primary/70 [animation-delay:120ms]" />
      <span className="h-1.5 w-1.5 animate-pulse rounded-full bg-primary/45 [animation-delay:240ms]" />
      <span className="ml-1">{label}</span>
    </div>
  );
}
