<script>
  import { onMount } from 'svelte';
  import { getAddress } from '@ethersproject/address';
  import { CloudflareProvider } from '@ethersproject/providers';
  import { setDefaults as setToast, toast } from 'bulma-toast';

  let input = null;
  let selectedNetwork = '';
  let faucetInfo = {
    default_network: '',
    active_networks: {},
    supported_networks: {},
    hcaptcha_sitekey: '',
  };

  let mounted = false;
  let hcaptchaLoaded = false;

  onMount(async () => {
    const res = await fetch('/api/info');
    faucetInfo = await res.json();
    selectedNetwork = faucetInfo.default_network;
    mounted = true;
  });

  window.hcaptchaOnLoad = () => {
    hcaptchaLoaded = true;
  };

  // Reactive statements
  $: currentNetwork = faucetInfo.active_networks[selectedNetwork] || {};
  $: document.title = 'CSO FAUCET';

  let widgetID;
  $: if (mounted && hcaptchaLoaded && faucetInfo.hcaptcha_sitekey) {
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
    if (address === null || address === '') {
      toast({ message: 'Address required', type: 'is-warning' });
      return;
    }

    if (!selectedNetwork) {
      toast({ message: 'Please select a network', type: 'is-warning' });
      return;
    }

    if (address.endsWith('.eth')) {
      try {
        const provider = new CloudflareProvider();
        address = await provider.resolveName(address);
        if (!address) {
          toast({ message: 'Invalid ENS name', type: 'is-warning' });
          return;
        }
      } catch (error) {
        toast({ message: error.reason || 'ENS resolution failed', type: 'is-warning' });
        return;
      }
    }

    try {
      address = getAddress(address);
    } catch (error) {
      toast({ message: error.reason || 'Invalid address', type: 'is-warning' });
      return;
    }

    try {
      let headers = {
        'Content-Type': 'application/json',
      };

      if (hcaptchaLoaded && faucetInfo.hcaptcha_sitekey) {
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
          network: selectedNetwork,
        }),
      });

      let { msg } = await res.json();
      
      if (res.ok) {
        // Extract transaction hash from message (format: "Txhash: 0x...")
        const txHashMatch = msg.match(/Txhash:\s*(0x[a-fA-F0-9]{64})/);
        if (txHashMatch) {
          const txHash = txHashMatch[1];
          showTransactionSuccess(txHash, selectedNetwork);
        } else {
          toast({ message: msg, type: 'is-success' });
        }
      } else {
        toast({ message: msg, type: 'is-warning' });
      }
    } catch (err) {
      console.error(err);
      toast({ message: 'Request failed', type: 'is-error' });
    }
  }

  function capitalize(str) {
    if (!str) return '';
    const lower = str.toLowerCase();
    return str.charAt(0).toUpperCase() + lower.slice(1);
  }

  // Network icon function removed for cleaner UI
  
  function getExplorerUrl(network, txHash) {
    const explorerUrls = {
      'mainnet': 'https://etherscan.io/tx/',
      'sepolia': 'https://sepolia.etherscan.io/tx/',  
      'holesky': 'https://holesky.etherscan.io/tx/',
      'goerli': 'https://goerli.etherscan.io/tx/',
      
      'polygon': 'https://polygonscan.com/tx/',
      'polygon-mumbai': 'https://mumbai.polygonscan.com/tx/',
      'polygon-amoy': 'https://amoy.polygonscan.com/tx/',
      
      'bsc': 'https://bscscan.com/tx/',
      'bsc-testnet': 'https://testnet.bscscan.com/tx/',
      
      'arbitrum': 'https://arbiscan.io/tx/',
      'arbitrum-sepolia': 'https://sepolia.arbiscan.io/tx/',
      
      'optimism': 'https://optimistic.etherscan.io/tx/',
      'optimism-sepolia': 'https://sepolia-optimism.etherscan.io/tx/',
      
      'avalanche': 'https://snowtrace.io/tx/',
      'avalanche-fuji': 'https://testnet.snowtrace.io/tx/',
      
      'base': 'https://basescan.org/tx/',
      'base-sepolia': 'https://sepolia.basescan.org/tx/',
      
      'fantom': 'https://ftmscan.com/tx/',
      'fantom-testnet': 'https://testnet.ftmscan.com/tx/',
      
      'linea': 'https://lineascan.build/tx/',
      'linea-sepolia': 'https://sepolia.lineascan.build/tx/',
      
      'zksync': 'https://explorer.zksync.io/tx/',
      'zksync-sepolia': 'https://sepolia.explorer.zksync.io/tx/'
    };

    const baseUrl = explorerUrls[network];
    return baseUrl ? baseUrl + txHash : '#';
  }

  function showTransactionSuccess(txHash, network) {
    const explorerUrl = getExplorerUrl(network, txHash);
    
    if (explorerUrl !== '#') {
      const message = `<a href="${explorerUrl}" target="_blank" rel="noopener noreferrer" style="color: #00d1b2; text-decoration: underline;">${txHash}</a>`;
      toast({ 
        message, 
        type: 'is-info', 
        duration: 8000,
        position: 'bottom-center',
        className: 'custom-long-toast'
      });
    } else {
      toast({ 
        message: `<a href="${explorerUrl}" target="_blank" rel="noopener noreferrer" style="color: #00d1b2; text-decoration: underline;">${txHash}</a>`,
        type: 'is-info',
        duration: 8000,
        position: 'bottom-center',
        className: 'custom-long-toast'
      });
    }
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
                <i class="fas fa-bath" />
              </span>
              <span><b>CSO Faucet</b></span>
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
        <div class="column is-8 is-offset-2">
          <h1 class="title">
            CSO Faucet
          </h1>
          <h2 class="subtitle">
            Get testnet tokens for multiple blockchain networks
          </h2>

          <div class="box">
            <!-- Network Selection -->
            <div class="field">
              <label class="label">Select Network</label>
              <div class="control">
                <div class="select is-fullwidth is-large">
                  <select bind:value={selectedNetwork}>
                    <option value="">Choose a network...</option>
                    {#each Object.entries(faucetInfo.active_networks) as [network, info]}
                      <option value={network}>
                        {info.name} ({info.symbol})
                        {#if info.is_testnet}(Testnet){/if}
                      </option>
                    {/each}
                  </select>
                </div>
              </div>
            </div>

            <!-- Network Info Display -->
            {#if selectedNetwork && currentNetwork.name}
              <div class="notification is-dark is-light">
                <div class="level">
                  <div class="level-left">
                    <div class="level-item">
                      <div>
                        <p class="heading">Network</p>
                        <p class="title is-5">{currentNetwork.name}</p>
                      </div>
                    </div>
                  </div>
                  <div class="level-right">
                    <div class="level-item">
                      <div>
                        <p class="heading">Payout</p>
                        <p class="title is-5">{currentNetwork.payout} {currentNetwork.symbol}</p>
                      </div>
                    </div>
                    <div class="level-item">
                      <div>
                        <p class="heading">Faucet Account</p>
                        <p class="title is-6">{currentNetwork.account?.slice(0, 12)}...</p>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            {/if}

            <!-- Address Input -->
            <div class="field">
              <label class="label">Wallet Address or ENS Name</label>
              <div class="control">
                <input
                  bind:value={input}
                  class="input is-large"
                  type="text"
                  placeholder="0x... or name.eth"
                  disabled={!selectedNetwork}
                />
              </div>
            </div>

            <!-- hCaptcha -->
            <div id="hcaptcha" data-size="invisible"></div>

            <!-- Submit Button -->
            <div class="field">
              <div class="control">
                <button
                  on:click={handleRequest}
                  class="button is-success is-large is-fullwidth"
                  disabled={!selectedNetwork || !input}
                >
                  <span class="icon">
                    <i class="fas fa-faucet"></i>
                  </span>
                  <span>Request {currentNetwork.symbol || 'Tokens'}</span>
                </button>
              </div>
            </div>
          </div>

          <!-- Active Networks Overview removed -->
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
    padding: 1rem 0;
    line-height: 1.5;
    color: #b8b8d1;
    font-weight: 300;
  }
  
  .box {
    border-radius: 16px;
    background: rgba(255, 255, 255, 0.05);
    backdrop-filter: blur(10px);
    border: 1px solid rgba(255, 255, 255, 0.1);
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
  }
  
  .notification.is-dark.is-light {
    background: rgba(255, 255, 255, 0.08);
    color: #ffffff;
    border: 1px solid rgba(255, 255, 255, 0.15);
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
  
  .select select {
    background: rgba(255, 255, 255, 0.1);
    border: 1px solid rgba(255, 255, 255, 0.2);
    color: #ffffff;
    font-family: 'Inter', sans-serif;
  }
  
  .select select:focus {
    background: rgba(255, 255, 255, 0.15);
    border-color: #00d1b2;
  }
  
  .select select option {
    background: #1a1a3e;
    color: #ffffff;
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
  
  .label {
    color: #ffffff;
    font-weight: 500;
    font-family: 'Inter', sans-serif;
  }
  
  .navbar {
    background: rgba(15, 15, 35, 0.9);
    backdrop-filter: blur(10px);
  }
  
  .navbar-brand .navbar-item {
    color: #ffffff;
    font-weight: 600;
  }
  
  .card {
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    transition: all 0.3s ease;
    cursor: pointer;
  }
  
  .card:hover {
    transform: translateY(-4px);
    box-shadow: 0 12px 40px rgba(0, 0, 0, 0.4);
    background: rgba(255, 255, 255, 0.08);
  }
  
  .card.has-background-primary-light {
    background: rgba(0, 209, 178, 0.2);
    border-color: #00d1b2;
  }
  
  .notification {
    background: rgba(118, 113, 113, 0.9);
    margin-bottom: 1rem;
  }
  
  .title.is-5, .title.is-6 {
    color: #ffffff;
    font-family: 'Inter', sans-serif;
  }
  
  .subtitle.is-6 {
    color: #b8b8d1;
  }
  
  .tag.is-warning {
    background: #ffdd57;
    color: #000000;
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