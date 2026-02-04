<script lang="ts">
export let item: {
  processName: string;
  exePath: string;
  commercialName: string;
  icon: string;
  otherInfo: string;
};
</script>

<!-- The outer label is now in the parent component -->
{#if item.icon}
  <img
    src="data:image/png;base64,{item.icon}"
    class="me-2"
    style="width: 24px; height: 24px;"
    alt="App Icon"
  />
{:else}
  <div class="me-2" style="width: 24px; height: 24px;"></div>
{/if}
<span class="fw-bold me-2">{item.commercialName || item.processName}</span>
<span class="text-muted">{item.otherInfo}</span>
<span class="text-muted ms-auto">{item.processName}</span>
