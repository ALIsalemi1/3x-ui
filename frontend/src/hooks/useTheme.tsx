import { createContext, useCallback, useContext, useLayoutEffect, useMemo, useState } from 'react';
import type { ReactNode } from 'react';
import { theme as antdTheme } from 'antd';
import type { ThemeConfig } from 'antd';

export const PANEL_ACCENTS = {
  salihost: '#087f91',
  ocean: '#1675e8',
  lagoon: '#008b82',
  violet: '#7957d5',
  rose: '#c7446c',
  amber: '#ae6500',
} as const;
export type PanelAccent = keyof typeof PANEL_ACCENTS;
const storedAccent = localStorage.getItem('panel-accent') || 'salihost';
const initialAccent: PanelAccent =
  storedAccent in PANEL_ACCENTS ? (storedAccent as PanelAccent) : 'salihost';

const STORAGE_DARK = 'dark-mode';
const STORAGE_ULTRA = 'isUltraDarkThemeEnabled';

function readBool(key: string, fallback: boolean): boolean {
  const raw = localStorage.getItem(key);
  if (raw === null) return fallback;
  return raw === 'true';
}

function applyDom(isDark: boolean, isUltra: boolean) {
  document.documentElement.setAttribute('data-panel-brand', 'salihost');
  // Every entry point may be served beneath a custom webBasePath. Build the
  // favicon path from that base instead of assuming the site lives at '/'.
  let favicon = document.querySelector<HTMLLinkElement>('link[data-salihost-favicon]');
  if (!favicon) {
    favicon = document.createElement('link');
    favicon.rel = 'icon';
    favicon.type = 'image/svg+xml';
    favicon.setAttribute('data-salihost-favicon', 'true');
    document.head.appendChild(favicon);
  }
  const root = window.X_UI_BASE_PATH || '/';
  favicon.href = `${root.endsWith('/') ? root : root + '/'}icons/salihost-mark.svg`;
  document.body.classList.remove('dark', 'light');
  document.body.classList.add(isDark ? 'dark' : 'light');
  // Native scrollbars read color-scheme, not the body class.
  document.documentElement.style.colorScheme = isDark ? 'dark' : 'light';
  if (isUltra) {
    document.documentElement.setAttribute('data-theme', 'ultra-dark');
  } else {
    document.documentElement.removeAttribute('data-theme');
  }
  const msg = document.getElementById('message');
  if (msg) {
    msg.classList.remove('dark', 'light');
    msg.classList.add(isDark ? 'dark' : 'light');
  }
}

// module load so the document is in the right theme before React mounts.
const initialDark = readBool(STORAGE_DARK, true);
const initialUltra = readBool(STORAGE_ULTRA, false);
applyDom(initialDark, initialUltra);

const SALIHOST_LIGHT_TOKENS = {
  colorBgBase: '#f3f8fb',
  colorBgLayout: '#eef5f8',
  colorBgContainer: '#ffffff',
  colorBgElevated: '#ffffff',
};
const SALIHOST_DARK_TOKENS = {
  colorBgBase: '#081522',
  colorBgLayout: '#081522',
  colorBgContainer: '#112437',
  colorBgElevated: '#1a3045',
};
const SALIHOST_DARK_LAYOUT_TOKENS = {
  bodyBg: '#081522',
  headerBg: '#0b1b2c',
  headerColor: '#f3fdff',
  footerBg: '#081522',
  siderBg: '#0b1b2c',
  triggerBg: '#112437',
  triggerColor: '#f3fdff',
};
const SALIHOST_DARK_MENU_TOKENS = {
  darkItemBg: '#0b1b2c',
  darkSubMenuItemBg: '#102337',
  darkPopupBg: '#112437',
};

const DARK_TOKENS = {
  colorBgBase: '#17212b',
  colorBgLayout: '#17212b',
  colorBgContainer: '#202d3a',
  colorBgElevated: '#2a3b4c',
};
const ULTRA_DARK_TOKENS = {
  colorBgBase: '#000',
  colorBgLayout: '#000',
  colorBgContainer: '#101013',
  colorBgElevated: '#1a1a1e',
};
const DARK_LAYOUT_TOKENS = {
  bodyBg: '#17212b',
  headerBg: '#131d27',
  headerColor: '#ffffff',
  footerBg: '#17212b',
  siderBg: '#131d27',
  triggerBg: '#202d3a',
  triggerColor: '#ffffff',
};
const ULTRA_DARK_LAYOUT_TOKENS = {
  bodyBg: '#000',
  headerBg: '#050507',
  headerColor: '#ffffff',
  footerBg: '#000',
  siderBg: '#050507',
  triggerBg: '#1a1a1e',
  triggerColor: '#ffffff',
};
const DARK_MENU_TOKENS = {
  darkItemBg: '#131d27',
  darkSubMenuItemBg: '#17212b',
  darkPopupBg: '#202d3a',
};
const ULTRA_DARK_MENU_TOKENS = {
  darkItemBg: '#050507',
  darkSubMenuItemBg: '#000',
  darkPopupBg: '#101013',
};
const DARK_CARD_TOKENS = {
  colorBorderSecondary: 'rgba(255, 255, 255, 0.06)',
};
const ULTRA_DARK_CARD_TOKENS = {
  colorBorderSecondary: 'rgba(255, 255, 255, 0.04)',
};
const STATISTIC_TOKENS = {
  contentFontSize: 17,
  titleFontSize: 11,
};
const LIGHT_CONTRAST_TOKENS = {
  colorTextDescription: 'rgba(0, 0, 0, 0.58)',
  colorTextTertiary: 'rgba(0, 0, 0, 0.58)',
  colorTextPlaceholder: '#767676',
  colorError: '#cf1322',
  colorErrorText: '#cf1322',
  colorSuccessText: '#237804',
};
const LIGHT_BUTTON_TOKENS = {
  colorPrimary: '#0958d9',
  colorPrimaryHover: '#2468e5',
  colorPrimaryActive: '#073ea8',
};

// hashed:false drops the `:where(.css-<hash>)` wrapper antd puts around every
// rule. It costs nothing in specificity - `:where()` contributes zero, so the
// panel's own `.ant-*` overrides still win - and it removes roughly 5,700
// wrappers, 16% of the generated stylesheet, from what the browser has to parse.
//
// cssVar.key pins the CSS-variable scope. Every panel page mounts its own
// ConfigProvider (there is no root one), and without a fixed key each mints a
// fresh useId-derived scope, so navigating re-serialises and re-injects the whole
// token block under a new class instead of reusing the one already in the head.
const SHARED_STYLE_CONFIG = {
  hashed: false,
  cssVar: { key: 'xui' },
} as const;

export function buildAntdThemeConfig(
  isDark: boolean,
  isUltra: boolean,
  accent: PanelAccent = 'salihost',
  compact = false,
): ThemeConfig {
  const sharedTokens = {
    colorPrimary: PANEL_ACCENTS[accent],
    colorInfo: PANEL_ACCENTS[accent],
    colorSuccess: '#119b7a',
    colorWarning: '#df9b26',
    borderRadius: 10,
    controlHeight: compact ? 30 : 36,
  };

  if (!isDark) {
    return {
      ...SHARED_STYLE_CONFIG,
      algorithm: antdTheme.defaultAlgorithm,
      token: {
        ...(accent === 'salihost' ? SALIHOST_LIGHT_TOKENS : {}),
        ...LIGHT_CONTRAST_TOKENS,
        ...sharedTokens,
      },
      components: {
        Statistic: STATISTIC_TOKENS,
        Button: accent === 'ocean' ? LIGHT_BUTTON_TOKENS : {},
      },
    };
  }
  return {
    ...SHARED_STYLE_CONFIG,
    algorithm: antdTheme.darkAlgorithm,
    token: {
      ...(isUltra
        ? ULTRA_DARK_TOKENS
        : accent === 'salihost'
          ? SALIHOST_DARK_TOKENS
          : DARK_TOKENS),
      ...sharedTokens,
    },
    components: {
      Layout: isUltra
        ? ULTRA_DARK_LAYOUT_TOKENS
        : accent === 'salihost'
          ? SALIHOST_DARK_LAYOUT_TOKENS
          : DARK_LAYOUT_TOKENS,
      Menu: isUltra
        ? ULTRA_DARK_MENU_TOKENS
        : accent === 'salihost'
          ? SALIHOST_DARK_MENU_TOKENS
          : DARK_MENU_TOKENS,
      Card: isUltra ? ULTRA_DARK_CARD_TOKENS : DARK_CARD_TOKENS,
      Statistic: STATISTIC_TOKENS,
    },
  };
}

export function pauseAnimationsUntilLeave(elementId: string): void {
  document.documentElement.setAttribute('data-theme-animations', 'off');
  const el = document.getElementById(elementId);
  if (!el) return;
  const restore = () => {
    document.documentElement.removeAttribute('data-theme-animations');
    el.removeEventListener('mouseleave', restore);
    el.removeEventListener('touchend', restore);
  };
  el.addEventListener('mouseleave', restore);
  el.addEventListener('touchend', restore);
}

interface ThemeContextValue {
  accent: PanelAccent;
  setAccent: (accent: PanelAccent) => void;
  compact: boolean;
  setCompact: (compact: boolean) => void;
  isDark: boolean;
  isUltra: boolean;
  toggleTheme: () => void;
  toggleUltra: () => void;
  antdThemeConfig: ThemeConfig;
}

const ThemeContext = createContext<ThemeContextValue | null>(null);

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [accent, setAccent] = useState<PanelAccent>(initialAccent);
  const [compact, setCompact] = useState(() => readBool('panel-compact', false));
  useLayoutEffect(() => {
    localStorage.setItem('panel-accent', accent);
    localStorage.setItem('panel-compact', String(compact));
    document.documentElement.style.setProperty('--panel-accent', PANEL_ACCENTS[accent]);
  }, [accent, compact]);
  const [isDark, setIsDark] = useState<boolean>(initialDark);
  const [isUltra, setIsUltra] = useState<boolean>(initialUltra);

  useLayoutEffect(() => {
    applyDom(isDark, isUltra);
    localStorage.setItem(STORAGE_DARK, String(isDark));
    localStorage.setItem(STORAGE_ULTRA, String(isUltra));
  }, [isDark, isUltra]);

  const toggleTheme = useCallback(() => setIsDark((v) => !v), []);
  const toggleUltra = useCallback(() => setIsUltra((v) => !v), []);

  const antdThemeConfig = useMemo(
    () => buildAntdThemeConfig(isDark, isUltra, accent, compact),
    [isDark, isUltra, accent, compact],
  );

  const value = useMemo<ThemeContextValue>(
    () => ({
      isDark,
      isUltra,
      toggleTheme,
      toggleUltra,
      antdThemeConfig,
      accent,
      setAccent,
      compact,
      setCompact,
    }),
    [isDark, isUltra, toggleTheme, toggleUltra, antdThemeConfig, accent, compact],
  );

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>;
}

export function useTheme(): ThemeContextValue {
  const ctx = useContext(ThemeContext);
  if (!ctx) throw new Error('useTheme must be used inside <ThemeProvider>');
  return ctx;
}
