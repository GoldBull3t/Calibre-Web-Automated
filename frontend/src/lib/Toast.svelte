<script>
  // @ts-nocheck
  import { toasts, removeToast } from './toast.js';
</script>

<div class="toast-container" aria-live="polite" aria-atomic="true">
  {#each $toasts as item (item.id)}
    <div class="toast-item glass {item.type}">
      <div class="toast-icon">
        {#if item.type === 'success'}
          <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"></polyline></svg>
        {:else if item.type === 'error'}
          <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.5"><circle cx="12" cy="12" r="10"></circle><line x1="15" y1="9" x2="9" y2="15"></line><line x1="9" y1="9" x2="15" y2="15"></line></svg>
        {:else if item.type === 'warning'}
          <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path><line x1="12" y1="9" x2="12" y2="13"></line><line x1="12" y1="17" x2="12.01" y2="17"></line></svg>
        {:else}
          <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.5"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="16" x2="12" y2="12"></line><line x1="12" y1="8" x2="12.01" y2="8"></line></svg>
        {/if}
      </div>
      <div class="toast-content">
        <p class="toast-msg">{item.message}</p>
      </div>
      <button class="toast-close" on:click={() => removeToast(item.id)} aria-label="Fechar notificação">
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
      </button>
    </div>
  {/each}
</div>

<style>
  .toast-container {
    position: fixed;
    top: 20px;
    right: 20px;
    z-index: 99999;
    display: flex;
    flex-direction: column;
    gap: 10px;
    max-width: 420px;
    width: calc(100vw - 40px);
    pointer-events: none;
  }

  .toast-item {
    pointer-events: auto;
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 16px;
    border-radius: 12px;
    background: var(--bg-card, rgba(22, 27, 34, 0.88));
    backdrop-filter: blur(16px);
    -webkit-backdrop-filter: blur(16px);
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.35), 0 0 0 1px rgba(255, 255, 255, 0.08);
    color: var(--text-primary, #f0f6fc);
    animation: toast-in 0.28s cubic-bezier(0.16, 1, 0.3, 1) forwards;
  }

  @keyframes toast-in {
    from {
      opacity: 0;
      transform: translateY(-12px) scale(0.96);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }

  .toast-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .toast-item.success {
    border-left: 4px solid #10b981;
  }
  .toast-item.success .toast-icon {
    color: #10b981;
  }

  .toast-item.error {
    border-left: 4px solid #ef4444;
  }
  .toast-item.error .toast-icon {
    color: #ef4444;
  }

  .toast-item.warning {
    border-left: 4px solid #f59e0b;
  }
  .toast-item.warning .toast-icon {
    color: #f59e0b;
  }

  .toast-item.info {
    border-left: 4px solid #06b6d4;
  }
  .toast-item.info .toast-icon {
    color: #06b6d4;
  }

  .toast-content {
    flex: 1;
    min-width: 0;
  }

  .toast-msg {
    margin: 0;
    font-size: 0.9rem;
    font-weight: 500;
    line-height: 1.4;
    word-break: break-word;
  }

  .toast-close {
    background: transparent;
    border: none;
    color: var(--text-secondary, #8b949e);
    cursor: pointer;
    padding: 4px;
    border-radius: 6px;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.15s ease;
  }

  .toast-close:hover {
    color: var(--text-primary, #ffffff);
    background: rgba(255, 255, 255, 0.1);
  }
</style>

