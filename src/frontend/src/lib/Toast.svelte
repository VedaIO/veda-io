<script lang="ts">
import { Toast } from 'bootstrap';
import { onMount } from 'svelte';
import { toast } from './toastStore';

let toastElement: HTMLElement;
let toastInstance: Toast;

onMount(() => {
  toastInstance = new Toast(toastElement, { autohide: true, delay: 3000 });
});

toast.subscribe((msg) => {
  if (msg && toastInstance) {
    toastInstance.show();
  }
});
</script>

<div class="toast-container position-fixed end-0 p-3" style="top: 100px;">
  <div
    bind:this={toastElement}
    class="toast"
    role="alert"
    aria-live="assertive"
    aria-atomic="true"
  >
    {#if $toast}
      <div
        class="toast-header text-black"
        class:bg-dark={$toast.type === 'info'}
        class:bg-danger={$toast.type === 'error'}
        style={$toast.type === 'success' ? 'background-color: #f76f53;' : ''}
      >
        <strong class="me-auto">Thông báo</strong>
        <button
          type="button"
          class="btn-close"
          data-bs-dismiss="toast"
          aria-label="Close"
        ></button>
      </div>
      <div
        class="toast-body"
        style="background-color: white; color: black; border-bottom-left-radius: 0.375rem; border-bottom-right-radius: 0.375rem;"
      >
        {$toast.message}
      </div>
    {/if}
  </div>
</div>
