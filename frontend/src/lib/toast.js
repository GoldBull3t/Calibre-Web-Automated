import { writable } from 'svelte/store';

export const toasts = writable([]);

let counter = 0;

export function addToast({ message, type = 'info', duration = 3500 }) {
  const id = ++counter;
  const item = { id, message, type };
  toasts.update(current => [...current, item]);

  if (duration > 0) {
    setTimeout(() => {
      removeToast(id);
    }, duration);
  }
  return id;
}

export function removeToast(id) {
  toasts.update(current => current.filter(t => t.id !== id));
}

export const toast = {
  success: (msg, duration = 3500) => addToast({ message: msg, type: 'success', duration }),
  error: (msg, duration = 4500) => addToast({ message: msg, type: 'error', duration }),
  info: (msg, duration = 3500) => addToast({ message: msg, type: 'info', duration }),
  warning: (msg, duration = 4000) => addToast({ message: msg, type: 'warning', duration })
};

