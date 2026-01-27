import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatSize(kb: number): string {
  if (kb === 0) return "0 B";
  const units = ["KB", "MB", "GB", "TB", "PB"];
  const i = Math.floor(Math.log(kb) / Math.log(1024));
  return `${(kb / Math.pow(1024, i)).toFixed(2)} ${units[i]}`;
}
