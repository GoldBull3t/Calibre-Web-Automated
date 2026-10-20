<script>
  export let count = 18;
</script>

<div class="skeleton-grid">
  {#each Array(count) as _, i}
    <div class="skeleton-card glass" style="animation-delay: {i * 40}ms">
      <div class="skeleton-cover shimmer"></div>
      <div class="skeleton-info">
        <div class="skeleton-line title shimmer"></div>
        <div class="skeleton-line author shimmer"></div>
        <div class="skeleton-line badge shimmer"></div>
      </div>
    </div>
  {/each}
</div>

<style>
  .skeleton-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(170px, 1fr));
    gap: 20px;
    width: 100%;
  }

  @media (max-width: 600px) {
    .skeleton-grid {
      grid-template-columns: repeat(auto-fill, minmax(130px, 1fr));
      gap: 12px;
    }
  }

  .skeleton-card {
    display: flex;
    flex-direction: column;
    border-radius: var(--radius-md, 12px);
    overflow: hidden;
    background: var(--bg-card, rgba(22, 27, 34, 0.7));
    border: 1px solid var(--border-color, rgba(255, 255, 255, 0.08));
  }

  .skeleton-cover {
    width: 100%;
    aspect-ratio: 2 / 3;
    background: rgba(255, 255, 255, 0.04);
  }

  .skeleton-info {
    padding: 12px 10px;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .skeleton-line {
    border-radius: 4px;
    background: rgba(255, 255, 255, 0.06);
  }

  .skeleton-line.title {
    height: 14px;
    width: 85%;
  }

  .skeleton-line.author {
    height: 11px;
    width: 60%;
  }

  .skeleton-line.badge {
    height: 10px;
    width: 40%;
    margin-top: 4px;
  }

  .shimmer {
    position: relative;
    overflow: hidden;
  }

  .shimmer::after {
    content: '';
    position: absolute;
    top: 0;
    right: 0;
    bottom: 0;
    left: 0;
    transform: translateX(-100%);
    background-image: linear-gradient(
      90deg,
      rgba(255, 255, 255, 0) 0,
      rgba(255, 255, 255, 0.06) 20%,
      rgba(255, 255, 255, 0.12) 60%,
      rgba(255, 255, 255, 0)
    );
    animation: shimmer 1.6s infinite ease-in-out;
  }

  @keyframes shimmer {
    100% {
      transform: translateX(100%);
    }
  }
</style>

