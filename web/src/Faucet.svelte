<script>
  import { onMount } from 'svelte';
  import { getAddress } from '@ethersproject/address';
  import { CloudflareProvider } from '@ethersproject/providers';
  import { setDefaults as setToast, toast } from 'bulma-toast';

  let input = null;
  let faucetInfo = {
    account: '0x0000000000000000000000000000000000000000',
    network: 'testnet',
    payout: 1,
    symbol: 'ETH',
    hcaptcha_sitekey: '',
    explorer_url: '',
  };

  let mounted = false;
  let hcaptchaLoaded = false;

  onMount(async () => {
    const res = await fetch('/api/info');
    faucetInfo = await res.json();
    mounted = true;
  });

  window.hcaptchaOnLoad = () => {
    hcaptchaLoaded = true;
  };

  $: document.title = 'FAUCET';

  let widgetID;
  $: if (mounted && hcaptchaLoaded) {
    widgetID = window.hcaptcha.render('hcaptcha', {
      sitekey: faucetInfo.hcaptcha_sitekey,
    });
  }

  setToast({
    position: 'bottom-right',
    dismissible: true,
    pauseOnHover: true,
    closeOnClick: true,
    animate: { in: 'fadeIn', out: 'fadeOut' },
  });

  async function handleRequest() {
    let address = input;
    if (address === null) {
      toast({ message: 'input required', type: 'is-warning' });
      return;
    }

    if (address.endsWith('.eth')) {
      try {
        const provider = new CloudflareProvider();
        address = await provider.resolveName(address);
        if (!address) {
          toast({ message: 'invalid ENS name', type: 'is-warning' });
          return;
        }
      } catch (error) {
        toast({ message: error.reason, type: 'is-warning' });
        return;
      }
    }

    try {
      address = getAddress(address);
    } catch (error) {
      toast({ message: error.reason, type: 'is-warning' });
      return;
    }

    try {
      let headers = {
        'Content-Type': 'application/json',
      };

      if (hcaptchaLoaded) {
        const { response } = await window.hcaptcha.execute(widgetID, {
          async: true,
        });
        headers['h-captcha-response'] = response;
      }

      const res = await fetch('/api/claim', {
        method: 'POST',
        headers,
        body: JSON.stringify({
          address,
        }),
      });

      let { msg } = await res.json();
      
      if (res.ok) {
        // Extract transaction hash from message (format: "Txhash: 0x...")
        const txHashMatch = msg.match(/Txhash:\s*(0x[a-fA-F0-9]{64})/);
        if (txHashMatch) {
          const txHash = txHashMatch[1];
          showTransactionSuccess(txHash);
        } else {
          toast({ message: msg, type: 'is-success' });
        }
      } else {
        toast({ message: msg, type: 'is-warning' });
      }
    } catch (err) {
      console.error(err);
    }
  }

  function capitalize(str) {
    const lower = str.toLowerCase();
    return str.charAt(0).toUpperCase() + lower.slice(1);
  }

  function getExplorerUrl(txHash) {
    const baseUrl = faucetInfo.explorer_url;
    if (!baseUrl) {
      return '#';
    }
    return `${baseUrl}${txHash}`;
  }

  function showTransactionSuccess(txHash) {
    const explorerUrl = getExplorerUrl(txHash);
    const networkName = faucetInfo.network;
    
    const link = explorerUrl !== '#'
      ? `<a href="${explorerUrl}" target="_blank" rel="noopener noreferrer" style="color: #00d1b2; text-decoration: underline;">${txHash}</a>`
      : txHash;

    toast({
      message: link,
      type: 'is-info',
      duration: 8000,
      position: 'bottom-center',
      className: 'custom-long-toast'
    });
  }
</script>

<svelte:head>
  {#if mounted && faucetInfo.hcaptcha_sitekey}
    <script
      src="https://hcaptcha.com/1/api.js?onload=hcaptchaOnLoad&render=explicit"
      async
      defer
    ></script>
  {/if}
</svelte:head>

<main>
  <section class="hero is-dark is-fullheight">
    <div class="hero-head">
      <nav class="navbar">
        <div class="container">
          <div class="navbar-brand">
            <a class="navbar-item" href="../..">
              <span class="icon">
                <i class="fa fa-bath" />
              </span>
              <span><b>{faucetInfo.symbol} Faucet</b></span>
            </a>
          </div>
          <div id="navbarMenu" class="navbar-menu">
            <div class="navbar-end">
              <!-- Removed View Source button -->
            </div>
          </div>
        </div>
      </nav>
    </div>

    <div class="hero-body">
      <div class="container has-text-centered">
        <div class="column is-6 is-offset-3">
          <h1 class="title">
            Receive {faucetInfo.payout}
            {faucetInfo.symbol} per request
          </h1>
          <h2 class="subtitle">
            Serving from {faucetInfo.account}
          </h2>
          <div id="hcaptcha" data-size="invisible"></div>
          <div class="box">
            <div class="field is-grouped">
              <p class="control is-expanded">
                <input
                  bind:value={input}
                  class="input is-rounded"
                  type="text"
                  placeholder="Enter your address or ENS name"
                />
              </p>
              <p class="control">
                <button
                  on:click={handleRequest}
                  class="button is-success is-rounded"
                >
                  Request
                </button>
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</main>

<style>
  @import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap');
  
  :global(body) {
    font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
    background: #0f0f23;
    color: #ffffff;
  }
  
  .hero.is-dark {
    background: linear-gradient(135deg, #0f0f23 0%, #1a1a3e 50%, #0f0f23 100%);
    min-height: 100vh;
  }
  
  .hero .title {
    color: #ffffff;
    font-weight: 600;
    text-shadow: 0 2px 4px rgba(0,0,0,0.3);
    font-family: 'Inter', sans-serif;
  }
  
  .hero .subtitle {
    padding: 3rem 0;
    line-height: 1.5;
    color: #b8b8d1;
    font-weight: 300;
  }
  
  .box {
    border-radius: 19px;
    background: rgba(255, 255, 255, 0.05);
    backdrop-filter: blur(10px);
    border: 1px solid rgba(255, 255, 255, 0.1);
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
  }
  
  .input {
    background: rgba(255, 255, 255, 0.1);
    border: 1px solid rgba(255, 255, 255, 0.2);
    color: #ffffff;
    font-family: 'Inter', sans-serif;
  }
  
  .input:focus {
    background: rgba(255, 255, 255, 0.15);
    border-color: #00d1b2;
    box-shadow: 0 0 0 0.125em rgba(0, 209, 178, 0.25);
  }
  
  .input::placeholder {
    color: rgba(255, 255, 255, 0.6);
  }
  
  .button.is-success {
    background: linear-gradient(45deg, #00d1b2, #00b894);
    border: none;
    font-weight: 500;
    transition: all 0.3s ease;
    font-family: 'Inter', sans-serif;
  }
  
  .button.is-success:hover {
    background: linear-gradient(45deg, #00b894, #00a085);
    transform: translateY(-2px);
    box-shadow: 0 8px 25px rgba(0, 209, 178, 0.3);
  }
  
  .navbar {
    background: rgba(15, 15, 35, 0.9);
    backdrop-filter: blur(10px);
  }
  
  .navbar-brand .navbar-item {
    color: #ffffff;
    font-weight: 600;
  }
  
  /* Custom toast styles */
  :global(.custom-long-toast) {
    background: transparent !important;
    border: none !important;
    box-shadow: none !important;
    padding: 0 !important;
    bottom: 20px !important;
    left: 50% !important;
    transform: translateX(-50%) !important;
    position: fixed !important;
  }
  
  :global(.custom-long-toast .toast-body) {
    background: transparent !important;
    border: none !important;
    padding: 0 !important;
  }
</style>
