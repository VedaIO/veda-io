<!--
  This component provides a user-friendly "From" and "To" date range picker.
  It uses the `flatpickr` library directly because the `svelte-flatpickr` wrapper
  was found to be incompatible with the project's modern Svelte and Vite setup,
  causing build failures. Initializing `flatpickr` in `onMount` is a more robust
  approach that avoids these build issues.
-->
<script lang="ts">
import flatpickr from 'flatpickr';
import { createEventDispatcher, onDestroy, onMount } from 'svelte';
import 'flatpickr/dist/flatpickr.css';
import { Vietnamese } from 'flatpickr/dist/l10n/vn.js';
import type { Instance } from 'flatpickr/dist/types/instance';
import type { Options } from 'flatpickr/dist/types/options';

export let since: Date | null = null;
export let until: Date | null = null;

const dispatch = createEventDispatcher();

let fromInput: HTMLInputElement;
let toInput: HTMLInputElement;
let fromInstance: Instance | undefined;
let toInstance: Instance | undefined;

const commonOptions: Partial<Options> = {
  enableTime: true,
  time_24hr: true,
  dateFormat: 'Y-m-d H:i',
  locale: Vietnamese,
};

onMount(() => {
  const fromOptions: Options = {
    ...commonOptions,
    defaultDate: since,
    onChange: (selectedDates) => {
      if (!selectedDates[0]) {
        since = null;
      } else {
        since = selectedDates[0];
        toInstance?.set('minDate', since);
      }
      dispatch('change', { since, until });
    },
  };

  const toOptions: Options = {
    ...commonOptions,
    defaultDate: until,
    onChange: (selectedDates) => {
      if (!selectedDates[0]) {
        until = null;
      } else {
        until = selectedDates[0];
        fromInstance?.set('maxDate', until);
      }
      dispatch('change', { since, until });
    },
  };

  if (fromInput) {
    fromInstance = flatpickr(fromInput, fromOptions);
  }
  if (toInput) {
    toInstance = flatpickr(toInput, toOptions);
  }
});

onDestroy(() => {
  fromInstance?.destroy();
  toInstance?.destroy();
});
</script>

<div class="row g-2 align-items-center">
  <div class="col-auto">
    <label for="from-date-picker" class="col-form-label">Từ:</label>
  </div>
  <div class="col">
    <input
      id="from-date-picker"
      bind:this={fromInput}
      placeholder="Thời gian bắt đầu"
      class="form-control"
    />
  </div>
  <div class="col-auto">
    <label for="to-date-picker" class="col-form-label">Đến:</label>
  </div>
  <div class="col">
    <input
      id="to-date-picker"
      bind:this={toInput}
      placeholder="Thời gian kết thúc"
      class="form-control"
    />
  </div>
</div>
