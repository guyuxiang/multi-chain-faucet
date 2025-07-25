<script>
  import { onMount } from 'svelte';
  import 'animate.css';
  import 'bulma/css/bulma.css';
  import Faucet from './Faucet.svelte';
  import MultiChainFaucet from './MultiChainFaucet.svelte';

  let isMultiChain = false;
  let loading = true;

  onMount(async () => {
    try {
      // Check if this is a multi-chain faucet by calling the info API
      const res = await fetch('/api/info');
      const data = await res.json();
      
      // If response has active_networks, it's multi-chain mode
      isMultiChain = data.active_networks !== undefined;
    } catch (error) {
      console.error('Failed to detect faucet mode:', error);
      // Default to single chain mode
      isMultiChain = false;
    }
    loading = false;
  });
</script>

<svelte:head>
  <link
    href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css"
    rel="stylesheet"
  />
</svelte:head>

{#if loading}
  <div class="hero is-info is-fullheight">
    <div class="hero-body">
      <div class="container has-text-centered">
        <div class="is-size-1">
          <i class="fas fa-spinner fa-spin"></i>
        </div>
        <p class="title">Loading Faucet...</p>
      </div>
    </div>
  </div>
{:else if isMultiChain}
  <MultiChainFaucet />
{:else}
  <Faucet />
{/if}
