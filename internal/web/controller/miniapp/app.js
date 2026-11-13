(() => {
  'use strict';
  const tg = window.Telegram?.WebApp;
  const content = document.getElementById('content');
  const nav = Array.from(document.querySelectorAll('[data-screen]'));
  const strings = {
    fa: {
      tagline:'فروشگاه امن و پرسرعت',online:'آنلاین',hello:'سلام!',heroDescription:'اینترنت آزاد با سرعت و امنیت بالا. سرویس مناسب خودت رو انتخاب کن.',
      highSpeed:'اتصال پرسرعت',secure:'امن و خصوصی',quickAccess:'دسترسی سریع',buy:'خرید اشتراک',buySub:'پلن‌های ویژه',
      mySubs:'اشتراک‌های من',mySubsSub:'وضعیت سرویس‌ها',invoices:'فاکتورها',invoicesSub:'سفارش‌های من',account:'حساب کاربری',
      accountSub:'کیف پول و پشتیبانی',locations:'🌍 لوکیشن‌های فروشگاه',locationsCount:'لوکیشن',germany:'آلمان',netherlands:'هلند',
      albania:'آلبانی',locationNote:'لوکیشن واقعی سرویس به تنظیمات پلن شما بستگی دارد.',loading:'در حال دریافت اطلاعات...',home:'خانه',
      plans:'پلن‌ها',subscriptions:'اشتراک‌ها',details:'جزئیات پلن',buyNow:'ادامه خرید در ربات',featured:'محبوب',gb:'گیگ',days:'روز',
      toman:'تومان',unlimited:'نامحدود',noPlans:'در حال حاضر پلن فعالی وجود ندارد.',noSubs:'هنوز اشتراکی نداری.',
      noInvoices:'هنوز فاکتوری ثبت نشده.',wallet:'موجودی کیف پول',orders:'تعداد فاکتورها',activeSubs:'سرویس‌های من',
      profile:'حساب کاربری',balance:'موجودی',status:'وضعیت',active:'فعال',inactive:'غیرفعال',remaining:'مصرف',
      openBot:'ادامه در ربات تلگرام',subscriptionsTitle:'اشتراک‌های من',invoicesTitle:'فاکتورهای من',walletTitle:'کیف پول و راهنما',
      recentPlans:'پیشنهادهای فروشگاه',allPlans:'مشاهده همه پلن‌ها',support:'پشتیبانی',tutorial:'آموزش اتصال',
      botAction:'این عملیات در گفت‌وگوی ربات ادامه پیدا می‌کند.',secureAccess:'ورود امن با تلگرام',
      errorTitle:'ورود امن لازم است',errorBody:'لطفاً فروشگاه را از دکمه داخل ربات تلگرام باز کن. باز کردن مستقیم لینک مجاز نیست.',
      retry:'تلاش دوباره',reload:'تازه‌سازی',stale:'داده‌ها به‌روز نیستند. دوباره امتحان کن.',anonymous:'کاربر عزیز',
      paid:'پرداخت شده',pending:'در انتظار پرداخت',canceled:'لغو شده',review:'در انتظار تأیید',footer:'امن و سریع'
    },
    en: {
      tagline:'Secure premium VPN store',online:'Online',hello:'Hello!',heroDescription:'Fast, secure access. Choose a plan that fits your needs.',
      highSpeed:'High-speed access',secure:'Secure & private',quickAccess:'Quick access',buy:'Buy subscription',buySub:'Premium plans',
      mySubs:'My subscriptions',mySubsSub:'Service status',invoices:'Invoices',invoicesSub:'My orders',account:'My account',
      accountSub:'Wallet & support',locations:'🌍 Store locations',locationsCount:'locations',germany:'Germany',netherlands:'Netherlands',
      albania:'Albania',locationNote:'Actual server locations depend on the configured plan.',loading:'Loading your shop...',home:'Home',
      plans:'Plans',subscriptions:'Subscriptions',details:'Plan details',buyNow:'Continue in Telegram bot',featured:'Featured',gb:'GB',
      days:'days',toman:'Toman',unlimited:'Unlimited',noPlans:'No available plans at the moment.',noSubs:'You do not have subscriptions yet.',
      noInvoices:'No invoices yet.',wallet:'Wallet balance',orders:'Invoices',activeSubs:'My services',
      profile:'Profile',balance:'Balance',status:'Status',active:'Active',inactive:'Inactive',remaining:'Used',
      openBot:'Continue in Telegram',subscriptionsTitle:'My subscriptions',invoicesTitle:'My invoices',walletTitle:'Wallet & help',
      recentPlans:'Featured plans',allPlans:'View all plans',support:'Support',tutorial:'Setup guide',
      botAction:'This action continues in your Telegram bot chat.',secureAccess:'Secure Telegram login',
      errorTitle:'Telegram sign-in required',errorBody:'Open the store using the button in your Telegram bot. Direct links cannot sign you in.',
      retry:'Try again',reload:'Refresh',stale:'Could not load the latest data. Please retry.',anonymous:'Friend',
      paid:'Paid',pending:'Pending payment',canceled:'Canceled',review:'Awaiting approval',footer:'Private and fast'
    }
  };
  let state;
  let lang = 'fa';
  let current = 'home';
  const words = () => strings[lang];
  const format = n => new Intl.NumberFormat(lang === 'fa' ? 'fa-IR' : 'en-US').format(Number(n) || 0);
  const esc = s => String(s ?? '').replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
  const money = amount => format(amount) + ' ' + words().toman;
  const bytes = n => {
    if (!n || n <= 0) return words().unlimited;
    const g = n / (1024 * 1024 * 1024);
    return (g >= 1 ? format(g.toFixed(1)) + ' '+words().gb : format((n/(1024*1024)).toFixed(0))+' MB');
  };

  function localize() {
    document.documentElement.lang = lang;
    document.documentElement.dir = lang === 'fa' ? 'rtl' : 'ltr';
    document.querySelectorAll('[data-t]').forEach(el => { el.textContent = words()[el.dataset.t] || ''; });
  }
  function warning(message) {
    content.innerHTML = '<div class="warning">⚠️ '+esc(message)+'</div><div class="row-actions"><button class="action secondary" data-refresh="true">'+esc(words().retry)+'</button></div>';
  }
  function empty(title) { return '<div class="empty-state"><span style="font-size:30px">✦</span>'+esc(title)+'</div>'; }
  function planCards(plans) {
    return '<div class="plan-grid">'+plans.map((p,i) => {
      const quota = Number(p.trafficGB) > 0 ? format(p.trafficGB)+' '+words().gb : words().unlimited;
      return '<article class="plan-card'+(i===1?' featured':'')+'"><span class="plan-badge">'+(i===1?'✦ '+words().featured:'✧ VPN SHOP')+'</span>'
        +'<div class="plan-title">'+esc(p.name)+'</div><div class="plan-details"><span>◈ '+quota+'</span><span>◷ '+format(p.durationDays)+' '+words().days+'</span></div>'
        +'<div class="price">'+money(p.priceToman)+'</div><button class="choose" data-buy="'+Number(p.id)+'">'+esc(words().buyNow)+'</button></article>';
    }).join('')+'</div>';
  }
  function home() {
    const p = state.plans || [];
    return '<h2>✨ '+esc(words().recentPlans)+'</h2>'
      +(p.length ? planCards(p.slice(0,2)) : empty(words().noPlans))
      +'<p></p><button class="action secondary" data-screen="plans">'+esc(words().allPlans)+'</button>';
  }
  function plans() {
    const p = state.plans || [];
    return '<h2>🛍️ '+esc(words().buy)+'</h2>'+(p.length?planCards(p):empty(words().noPlans))
      +'<p class="note">'+esc(words().botAction)+'</p>';
  }
  function subscriptions() {
    const items = state.subscriptions || [];
    return '<h2>💎 '+esc(words().subscriptionsTitle)+'</h2>'
      +(items.length?'<div class="list">'+items.map(s=>{
        const total = Number(s.totalBytes)||0;
        const used = Number(s.usedBytes)||0;
        const percent = total>0 ? Math.min(100,Math.max(0,Math.round(used*100/total))) : 0;
        const status = s.enabled ? words().active : words().inactive;
        return '<div class="list-item"><div class="list-top"><strong>'+esc(s.name)+'</strong><span class="tag '+(s.enabled?'good':'warn')+'">'+esc(status)+'</span></div>'
          +'<p>'+esc(s.email)+'</p><p>◈ '+esc(words().remaining)+': '+bytes(used)+' / '+bytes(total)+'</p>'
          +'<div class="meter"><span style="width:'+percent+'%"></span></div></div>';
      }).join('')+'</div>':empty(words().noSubs))
      +'<p></p><button class="action secondary" data-bot-action="subscriptions">'+esc(words().openBot)+'</button>';
  }
  function invoices() {
    const items = state.invoices || [];
    const status = s => ({approved:words().paid,pending_payment:words().pending,pending_review:words().review,canceled:words().canceled})[s] || s;
    return '<h2>🧾 '+esc(words().invoicesTitle)+'</h2>'+(items.length?'<div class="list">'
      +items.map(o=>'<div class="list-item"><div class="list-top"><strong>#'+format(o.id)+'</strong><span class="tag">'+esc(status(o.status))+'</span></div><p>'+money(o.amountToman)+'</p></div>').join('')+'</div>'
      :empty(words().noInvoices))
      +'<p></p><button class="action secondary" data-bot-action="invoices">'+esc(words().openBot)+'</button>';
  }
  function account() {
    const sub = (state.subscriptions || []).length;
    const name = esc(state.customerName || words().anonymous);
    let html = '<h2>👤 '+esc(words().profile)+'</h2><p class="info-line">'+name+'</p>'
      +'<div class="summary-grid"><div class="metric"><small>'+esc(words().wallet)+'</small><strong>'+(state.walletEnabled?money(state.walletBalance):'—')+'</strong></div>'
      +'<div class="metric"><small>'+esc(words().activeSubs)+'</small><strong>'+format(sub)+'</strong></div>'
      +'<div class="metric"><small>'+esc(words().orders)+'</small><strong>'+format((state.invoices||[]).length)+'</strong></div>'
      +'<div class="metric"><small>'+esc(words().secureAccess)+'</small><strong>✓ Telegram</strong></div></div>'
      +'<div class="row-actions"><button class="action teal" data-bot-action="account">'+esc(words().openBot)+'</button>';
    if (state.walletEnabled) html += '<button class="action secondary" data-bot-action="wallet">'+esc(words().walletTitle)+'</button>';
    html += '</div>';
    for (const [url,label] of [[state.supportURL,words().support],[state.tutorialURL,words().tutorial]]) {
      if (url && /^https:\/\//i.test(url)) html += '<a class="support-link" rel="noopener noreferrer" target="_blank" href="'+esc(url)+'">'+esc(label)+' ↗</a>';
    }
    return html;
  }
  function render(screen = current) {
    current = screen;
    if (!state) return;
    const views = {home,plans,subscriptions,invoices,account};
    content.innerHTML = (views[screen] || home)();
    document.querySelectorAll('.bottom-nav .nav-btn').forEach(b=>{
      b.classList.toggle('active',b.dataset.screen===screen);
      if(b.dataset.screen===screen)b.setAttribute('aria-current','page');else b.removeAttribute('aria-current');
    });
    if(screen!=='home') content.scrollIntoView({behavior:'smooth',block:'start'});
  }
  function sendAction(type, planId) {
    if (!tg?.sendData || !tg.initData) {
      warning(words().errorBody); return;
    }
    const data = {type};
    if(type === 'buy') {
      if (!Number.isSafeInteger(planId) || !state.plans.some(p=>Number(p.id)===planId)) return;
      data.planId = planId;
    }
    tg.sendData(JSON.stringify(data)); // Telegram sends a verified user message and closes the Mini App.
  }
  document.addEventListener('click',event=>{
    const button = event.target.closest('button');
    if(!button)return;
    if(button.dataset.buy){sendAction('buy',Number(button.dataset.buy));return;}
    if(button.dataset.botAction){sendAction(button.dataset.botAction);return;}
    if(button.dataset.refresh){load();return;}
    if(button.dataset.screen)render(button.dataset.screen);
  });
  async function load() {
    if(!tg?.initData){warning(words().errorBody);return;}
    content.innerHTML='<div class="loading"><span class="spinner"></span>'+esc(words().loading)+'</div>';
    try{
      const response=await fetch('./state',{method:'POST',credentials:'omit',cache:'no-store',
        headers:{'X-Telegram-Init-Data':tg.initData,'Accept':'application/json'}});
      if(!response.ok) throw new Error(response.status===401?words().errorBody:words().stale);
      state=await response.json();
      lang=state.language==='en'?'en':'fa';
      localize();
      document.getElementById('customer-name').textContent=state.customerName || '';
      if(state.shopName)document.title=state.shopName+' • VPN SHOP';
      render(current);
    }catch(e){warning(e.message||words().stale);}
  }
  tg?.ready?.();
  tg?.expand?.();
  tg?.setHeaderColor?.('#0d1021');
  tg?.setBackgroundColor?.('#0b1020');
  load();
})();
