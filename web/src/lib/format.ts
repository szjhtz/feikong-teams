export function formatTime(value?: string | number | null) {
  if (!value) return "";
  const date = typeof value === "number" ? new Date(value * 1000) : new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return date.toLocaleString();
}

export function shortID(value?: string) {
  if (!value) return "";
  return value.length > 10 ? value.slice(0, 8) : value;
}

export function formatBytes(size?: number) {
  const value = Number(size || 0);
  if (value < 1024) return `${value} B`;
  const units = ["KB", "MB", "GB", "TB"];
  let current = value / 1024;
  let index = 0;
  while (current >= 1024 && index < units.length - 1) {
    current /= 1024;
    index += 1;
  }
  return `${current.toFixed(current >= 10 ? 0 : 1)} ${units[index]}`;
}

export function truncateText(value: string, max = 120) {
  if (value.length <= max) return value;
  return `${value.slice(0, max)}...`;
}
