import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatAge(timestamp: string): string {
  if (!timestamp) return "-";
  const date = new Date(timestamp);
  if (isNaN(date.getTime())) return "-";

  const now = new Date();
  const diff = Math.abs(now.getTime() - date.getTime());

  const seconds = Math.floor(diff / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);

  if (days > 0) return `${days}d`;
  if (hours > 0) return `${hours}h`;
  if (minutes > 0) return `${minutes}m`;
  return `${seconds}s`;
}

export function toYaml(obj: any, indent = 0): string {
  const prefix = "  ".repeat(indent);

  if (obj === null) return "null";
  if (typeof obj !== "object") return String(obj);

  if (Array.isArray(obj)) {
    if (obj.length === 0) return " []";
    return obj.map(item => `\n${prefix}- ${toYaml(item, indent + 1).trim()}`).join("");
  }

  const entries = Object.entries(obj);
  if (entries.length === 0) return " {}";

  return entries.map(([key, value]) => {
    if (typeof value === "object" && value !== null) {
      return `\n${prefix}${key}:${toYaml(value, indent + 1)}`;
    }
    return `\n${prefix}${key}: ${value}`;
  }).join("");
}
