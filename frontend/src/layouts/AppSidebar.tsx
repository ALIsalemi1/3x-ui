import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { ComponentType, CSSProperties } from 'react';
import { useLocation, useNavigate } from 'react-router';
import { useTranslation } from 'react-i18next';
import { Drawer, Dropdown, Layout, Menu, Modal, Tooltip } from 'antd';
import type { MenuProps } from 'antd';
import {
  ApiOutlined,
  BgColorsOutlined,
  ApartmentOutlined,
  CloseOutlined,
  CloudServerOutlined,
  ClusterOutlined,
  CodeOutlined,
  DashboardOutlined,
  DatabaseOutlined,
  DiscordOutlined,
  ExportOutlined,
  GlobalOutlined,
  InfoCircleOutlined,
  ImportOutlined,
  LogoutOutlined,
  MailOutlined,
  MenuOutlined,
  MessageOutlined,
  MoonFilled,
  MoonOutlined,
  PushpinFilled,
  PushpinOutlined,
  ReadOutlined,
  RobotOutlined,
  SafetyOutlined,
  SearchOutlined,
  SettingOutlined,
  SunOutlined,
  SwapOutlined,
  TagsOutlined,
  TeamOutlined,
  ToolOutlined,
} from '@ant-design/icons';

import { HttpUtil } from '@/utils';
import { formatPanelVersion } from '@/lib/panel-version';
import { PANEL_ACCENTS, pauseAnimationsUntilLeave, useTheme } from '@/hooks/useTheme';
import type { PanelAccent } from '@/hooks/useTheme';
import { useAllSettings } from '@/api/queries/useAllSettings';
import { useCommandPalette } from '@/components/command-palette/useCommandPalette';
import './AppSidebar.css';

// The palette listens for Ctrl as well as Cmd, so the chip must not show a
// Mac glyph to the Linux and Windows operators who are most of this panel's.
const SHORTCUT_MODIFIER = /Mac|iPhone|iPad|iPod/.test(navigator.userAgent) ? '⌘' : 'Ctrl';
const REPO_URL = 'https://github.com/ALIsalemi1/3x-ui';
const UPSTREAM_URL = 'https://github.com/MHSanaei/3x-ui';
const LOGOUT_KEY = '__logout__';
const RAIL_WIDTH = 72;
const SIDER_WIDTH = 248;
const SIDEBAR_PINNED_KEY = 'sidebar-pinned';

let hoveredAcrossRemounts = false;

type IconName =
  | 'dashboard'
  | 'inbound'
  | 'team'
  | 'groups'
  | 'setting'
  | 'tool'
  | 'cluster'
  | 'hosts'
  | 'logout'
  | 'bot'
  | 'apidocs'
  | 'outbound'
  | 'routing';

const iconByName: Record<IconName, ComponentType> = {
  dashboard: DashboardOutlined,
  inbound: ImportOutlined,
  team: TeamOutlined,
  groups: TagsOutlined,
  setting: SettingOutlined,
  tool: ToolOutlined,
  cluster: ClusterOutlined,
  hosts: GlobalOutlined,
  logout: LogoutOutlined,
  apidocs: ApiOutlined,
  outbound: ExportOutlined,
  routing: SwapOutlined,
  bot: RobotOutlined,
};

// Compact typographic identity avoids a remote logo dependency in the panel.
function SaliHostBrand({ collapsed = false }: { collapsed?: boolean }) {
  return (
    <span className="salihost-identity" aria-label="SaliHost.ir">
      <span className="salihost-symbol" aria-hidden="true">
        S
      </span>
      {!collapsed && (
        <span className="salihost-wordmark">
          <strong>
            Sali<span>Host</span>
          </strong>
          <small>CONTROL CENTER · salihost.ir</small>
        </span>
      )}
    </span>
  );
}

const ACCENT_NAMES: Record<PanelAccent, string> = {
  salihost: 'SaliHost',
  ocean: 'Ocean',
  lagoon: 'Lagoon',
  violet: 'Violet',
  rose: 'Rose',
  amber: 'Amber',
};

function AccentPicker({ collapsed = false }: { collapsed?: boolean }) {
  const { accent, setAccent } = useTheme();
  const options = (Object.keys(PANEL_ACCENTS) as PanelAccent[]).map((key) => ({
    key,
    label: (
      <span className="salihost-accent-option">
        <span className="salihost-accent-swatch" style={{ background: PANEL_ACCENTS[key] }} />
        {ACCENT_NAMES[key]}
      </span>
    ),
  }));
  return (
    <Dropdown
      trigger={['click']}
      placement="topLeft"
      menu={{
        items: options,
        selectedKeys: [accent],
        onClick: ({ key }) => setAccent(key as PanelAccent),
      }}
    >
      <button
        type="button"
        className={`salihost-accent-toggle${collapsed ? ' is-collapsed' : ''}`}
        aria-label="رنگ پنل / Panel accent color"
        title={`Panel accent: ${ACCENT_NAMES[accent]}`}
      >
        <BgColorsOutlined />
        {!collapsed && <span>Theme · {ACCENT_NAMES[accent]}</span>}
        <span className="salihost-current-accent" style={{ background: PANEL_ACCENTS[accent] }} />
      </button>
    </Dropdown>
  );
}

function DocsButton({ ariaLabel, onClick }: { ariaLabel: string; onClick: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="sidebar-docs"
      aria-label={ariaLabel}
      title={ariaLabel}
    >
      <ReadOutlined />
    </button>
  );
}

function VersionBadge({ version, collapsed }: { version: string; collapsed?: boolean }) {
  const [aboutOpen, setAboutOpen] = useState(false);
  const label = version ? formatPanelVersion(version) : 'Beta';
  return (
    <>
      <button
        type="button"
        className="sider-version salihost-about-trigger"
        aria-label="About SaliHost Control Center"
        title="About SaliHost Control Center"
        onClick={() => setAboutOpen(true)}
      >
        <InfoCircleOutlined />
        {!collapsed && (
          <span className="sider-version-text">
            SaliHost <span className="salihost-build-label">{label}</span>
          </span>
        )}
      </button>
      <Modal
        open={aboutOpen}
        title="SaliHost Control Center"
        onCancel={() => setAboutOpen(false)}
        footer={null}
        width={500}
        centered
        destroyOnHidden
      >
        <div className="salihost-about">
          <div className="salihost-about-kicker">SALIHOST.IR / NETWORK OPERATIONS</div>
          <p>Integrated network operations, subscription management and customer sales.</p>
          <div className="salihost-about-version">Build {label}</div>
          <details className="salihost-legal-notice">
            <summary>Open-source credits &amp; licenses</summary>
            <p>
              This modified distribution is based on 3X-UI by MHSanaei and distributed under the GNU
              GPL version 3. Original copyright notices and license terms remain in effect.
            </p>
            <p className="salihost-legal-links">
              <a href={UPSTREAM_URL} target="_blank" rel="noopener noreferrer">
                Upstream project
              </a>
              <a href={REPO_URL} target="_blank" rel="noopener noreferrer">
                SaliHost source
              </a>
              <a href={`${REPO_URL}/blob/main/LICENSE`} target="_blank" rel="noopener noreferrer">
                GPL-3.0 license
              </a>
            </p>
          </details>
        </div>
      </Modal>
    </>
  );
}

function ThemeCycleButton({
  id,
  isDark,
  isUltra,
  onCycle,
  ariaLabel,
}: {
  id: string;
  isDark: boolean;
  isUltra: boolean;
  onCycle: () => void;
  ariaLabel: string;
}) {
  const icon = !isDark ? <SunOutlined /> : !isUltra ? <MoonOutlined /> : <MoonFilled />;
  return (
    <button
      id={id}
      type="button"
      className="sidebar-theme-cycle"
      aria-label={ariaLabel}
      title={ariaLabel}
      onClick={onCycle}
    >
      {icon}
    </button>
  );
}

function readSidebarPinned() {
  try {
    return localStorage.getItem(SIDEBAR_PINNED_KEY) === 'true';
  } catch {
    return false;
  }
}

function saveSidebarPinned(pinned: boolean) {
  try {
    localStorage.setItem(SIDEBAR_PINNED_KEY, String(pinned));
  } catch {}
}

export default function AppSidebar() {
  const { t, i18n } = useTranslation();
  const { isDark, isUltra, toggleTheme, toggleUltra } = useTheme();
  const { open: openCommandPalette } = useCommandPalette();
  const navigate = useNavigate();
  const { pathname, hash } = useLocation();
  const { allSetting } = useAllSettings();
  const showSubFormats = !!(allSetting.subJsonEnable || allSetting.subClashEnable);
  const showSubBalancers = !!allSetting.subJsonEnable;

  const [hovered, setHovered] = useState(() => hoveredAcrossRemounts);
  const [pinned, setPinned] = useState(readSidebarPinned);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [userRole, setUserRole] = useState<string>('');
  const [roleFetched, setRoleFetched] = useState(false);

  useEffect(() => {
    HttpUtil.get<{ role: string }>('/panel/api/user/info', undefined, { silent: true })
      .then((resp) => {
        if (resp.obj?.role) {
          setUserRole(resp.obj.role);
        }
        setRoleFetched(true);
      })
      .catch(() => {
        setRoleFetched(true);
      });
  }, []);

  const isAdmin = roleFetched ? userRole === 'admin' : true;

  const railCollapsed = !hovered && !pinned;
  const railStyle = useMemo(
    () => ({ '--sider-rail': `${pinned ? SIDER_WIDTH : RAIL_WIDTH}px` }) as CSSProperties,
    [pinned],
  );
  const rootRef = useRef<HTMLDivElement>(null);

  const updateHovered = useCallback((value: boolean) => {
    hoveredAcrossRemounts = value;
    setHovered(value);
  }, []);

  const togglePinned = useCallback(() => {
    const next = !pinned;
    saveSidebarPinned(next);
    setPinned(next);
  }, [pinned]);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      const el = rootRef.current;
      if (el) updateHovered(el.matches(':hover'));
    }, 150);
    return () => window.clearTimeout(timer);
  }, [updateHovered]);

  const currentTheme: 'light' | 'dark' = isDark ? 'dark' : 'light';
  const panelVersion = window.X_UI_CUR_VER || '';

  const tabs = useMemo<{ key: string; icon: IconName; title: string }[]>(() => {
    const allTabs: { key: string; icon: IconName; title: string }[] = [
      { key: '/', icon: 'dashboard', title: t('menu.dashboard') },
      { key: '/reseller', icon: 'dashboard', title: t('pages.reseller.title', 'Reseller') },
      { key: '/inbounds', icon: 'inbound', title: t('menu.inbounds') },
      { key: '/clients', icon: 'team', title: t('menu.clients') },
      { key: '/groups', icon: 'groups', title: t('menu.groups') },
      { key: '/nodes', icon: 'cluster', title: t('menu.nodes') },
      { key: '/hosts', icon: 'hosts', title: t('menu.hosts') },
      { key: '/outbound', icon: 'outbound', title: t('menu.outbounds') },
      { key: '/routing', icon: 'routing', title: t('menu.routing') },
      { key: '/settings', icon: 'setting', title: t('menu.settings') },
      { key: '/xray', icon: 'tool', title: t('menu.xray') },
      { key: '/api-docs', icon: 'apidocs', title: t('menu.apiDocs') },
      { key: '/admins', icon: 'team', title: t('menu.admins') },
      { key: '/multiply', icon: 'inbound', title: t('menu.multiply') },
      { key: '/telegram-bot', icon: 'bot', title: t('menu.telegramBot') },
      { key: LOGOUT_KEY, icon: 'logout', title: t('logout') },
    ];
    const availableTabs = allTabs.filter(
      (tab) => tab.key !== '/telegram-bot' || allSetting.resellerBotEnable,
    );
    if (!isAdmin) {
      return availableTabs.filter((tab) =>
        ['/', '/reseller', '/inbounds', '/clients', '/telegram-bot', LOGOUT_KEY].includes(tab.key),
      );
    }
    return availableTabs;
  }, [t, isAdmin, allSetting.resellerBotEnable]);

  const navItems = useMemo(() => tabs.filter((tab) => tab.icon !== 'logout'), [tabs]);
  const utilItems = useMemo(() => tabs.filter((tab) => tab.icon === 'logout'), [tabs]);

  const settingsChildren = useMemo<NonNullable<MenuProps['items']>>(() => {
    const children: NonNullable<MenuProps['items']> = [
      {
        key: '/settings#general',
        icon: <SettingOutlined />,
        label: t('pages.settings.panelSettings'),
      },
      {
        key: '/settings#security',
        icon: <SafetyOutlined />,
        label: t('pages.settings.securitySettings'),
      },
      {
        key: '/settings#telegram',
        icon: <MessageOutlined />,
        label: t('pages.settings.TGBotSettings'),
      },
      { key: '/settings#email', icon: <MailOutlined />, label: t('pages.settings.emailSettings') },
      {
        key: '/settings#discord',
        icon: <DiscordOutlined />,
        label: t('pages.settings.discordSettings'),
      },
      {
        key: '/settings#subscription',
        icon: <CloudServerOutlined />,
        label: t('pages.settings.subSettings'),
      },
    ];
    if (showSubFormats) {
      children.push({
        key: '/settings#subscription-formats',
        icon: <CodeOutlined />,
        label: t('menu.subFormats'),
      });
    }
    if (showSubBalancers) {
      children.push({
        key: '/settings#subscription-balancers',
        icon: <ApartmentOutlined />,
        label: t('pages.settings.subBalancers.menu'),
      });
    }
    return children;
  }, [t, showSubFormats, showSubBalancers]);

  const xrayChildren = useMemo<NonNullable<MenuProps['items']>>(
    () => [
      { key: '/xray#basic', icon: <SettingOutlined />, label: t('pages.xray.basicTemplate') },
      { key: '/xray#balancer', icon: <ClusterOutlined />, label: t('pages.xray.Balancers') },
      { key: '/xray#dns', icon: <DatabaseOutlined />, label: 'DNS' },
      { key: '/xray#advanced', icon: <CodeOutlined />, label: t('pages.xray.advancedTemplate') },
    ],
    [t],
  );

  const settingsActive = pathname === '/settings';
  const xrayActive = pathname === '/xray';
  const selectedKey = settingsActive
    ? `/settings${hash || '#general'}`
    : xrayActive
      ? `/xray${hash || '#basic'}`
      : pathname === ''
        ? '/'
        : pathname;

  const openSubmenu = settingsActive ? '/settings' : xrayActive ? '/xray' : null;
  const [openKeys, setOpenKeys] = useState<string[]>(() => (openSubmenu ? [openSubmenu] : []));
  if (openSubmenu && !openKeys.includes(openSubmenu)) {
    setOpenKeys([...openKeys, openSubmenu]);
  }

  const toMenuItems = useCallback(
    (items: typeof tabs): MenuProps['items'] =>
      items.map((tab) => {
        const Icon = iconByName[tab.icon];
        if (tab.key === '/settings') {
          return { key: tab.key, icon: <Icon />, label: tab.title, children: settingsChildren };
        }
        if (tab.key === '/xray') {
          return { key: tab.key, icon: <Icon />, label: tab.title, children: xrayChildren };
        }
        return { key: tab.key, icon: <Icon />, label: tab.title, title: '' };
      }),
    [settingsChildren, xrayChildren],
  );

  const groupedMenuItems = useMemo<MenuProps['items']>(() => {
    const persian = i18n.language.startsWith('fa');
    const groups = [
      {
        id: 'workspace',
        title: persian ? 'فضای کاری' : 'WORKSPACE',
        routes: ['/', '/inbounds', '/clients', '/groups'],
      },
      {
        id: 'commerce',
        title: persian ? 'فروش و مشتریان' : 'COMMERCE',
        routes: ['/reseller', '/telegram-bot'],
      },
      {
        id: 'infrastructure',
        title: persian ? 'زیرساخت و شبکه' : 'INFRASTRUCTURE',
        routes: ['/nodes', '/hosts', '/outbound', '/routing', '/xray', '/multiply'],
      },
      {
        id: 'management',
        title: persian ? 'مدیریت و ابزارها' : 'MANAGEMENT',
        routes: ['/settings', '/admins', '/api-docs'],
      },
    ];
    return groups
      .map((group) => ({
        type: 'group' as const,
        key: `salihost-${group.id}`,
        label: <span className="salihost-nav-heading">{group.title}</span>,
        children: toMenuItems(navItems.filter((tab) => group.routes.includes(tab.key))) ?? [],
      }))
      .filter((group) => group.children.length > 0);
  }, [i18n.language, navItems, toMenuItems]);

  const openLink = useCallback(
    async (key: string) => {
      if (key === LOGOUT_KEY) {
        await HttpUtil.post('/logout');
        window.location.href = window.X_UI_BASE_PATH || '/';
        return;
      }
      navigate(key);
    },
    [navigate],
  );

  const onMenuClick = useCallback<NonNullable<MenuProps['onClick']>>(
    ({ key }) => {
      openLink(String(key));
    },
    [openLink],
  );

  const cycleTheme = useCallback(
    (id: string) => {
      pauseAnimationsUntilLeave(id);
      if (!isDark) {
        toggleTheme();
        if (isUltra) toggleUltra();
      } else if (!isUltra) {
        toggleUltra();
      } else {
        toggleUltra();
        toggleTheme();
      }
    },
    [isDark, isUltra, toggleTheme, toggleUltra],
  );

  return (
    <div
      ref={rootRef}
      className={`ant-sidebar${pinned ? ' sidebar-pinned' : ''}`}
      style={railStyle}
      onMouseEnter={() => updateHovered(true)}
      onMouseLeave={() => updateHovered(false)}
    >
      <Layout.Sider
        theme={currentTheme}
        width={SIDER_WIDTH}
        collapsedWidth={RAIL_WIDTH}
        collapsed={railCollapsed}
      >
        <div className="sider-brand">
          <div className="brand-block">
            <SaliHostBrand collapsed={railCollapsed} />
          </div>
          {!railCollapsed && (
            <div className="brand-actions">
              <button
                type="button"
                className="sidebar-pin"
                aria-label={t('menu.pinSidebar')}
                aria-pressed={pinned}
                title={t(pinned ? 'menu.unpinSidebar' : 'menu.pinSidebar')}
                onClick={togglePinned}
              >
                {pinned ? <PushpinFilled /> : <PushpinOutlined />}
              </button>
              <DocsButton
                ariaLabel={t('menu.docs') || 'Documentation'}
                onClick={() => navigate('/api-docs')}
              />
              <ThemeCycleButton
                id="theme-cycle"
                isDark={isDark}
                isUltra={isUltra}
                onCycle={() => cycleTheme('theme-cycle')}
                ariaLabel={t('menu.theme')}
              />
            </div>
          )}
        </div>
        <Tooltip
          title={
            railCollapsed ? t('commandPalette.title') || 'Command Palette (Ctrl + K)' : undefined
          }
          placement="right"
        >
          <button
            type="button"
            className={`sidebar-command-trigger${railCollapsed ? ' collapsed' : ''}`}
            onClick={openCommandPalette}
            aria-label={t('commandPalette.title') || 'Command Palette (Ctrl + K)'}
          >
            <span className="sidebar-command-left">
              <SearchOutlined className="sidebar-command-icon" />
              <span className="sidebar-command-text">
                {t('commandPalette.search') || 'Search...'}
              </span>
            </span>
            <span className="sidebar-command-kbd">
              <span className="kbd-cmd">{SHORTCUT_MODIFIER}</span>
              <span className="kbd-key">K</span>
            </span>
          </button>
        </Tooltip>
        <Menu
          theme={currentTheme}
          mode="inline"
          selectedKeys={[selectedKey]}
          openKeys={railCollapsed ? undefined : openKeys}
          onOpenChange={(keys) => setOpenKeys(keys as string[])}
          className="sider-nav"
          items={groupedMenuItems}
          onClick={onMenuClick}
        />
        <Menu
          theme={currentTheme}
          mode="inline"
          selectedKeys={[selectedKey]}
          className="sider-utility"
          items={toMenuItems(utilItems)}
          onClick={onMenuClick}
        />
        <div className="sider-footer">
          <AccentPicker collapsed={railCollapsed} />
          <VersionBadge version={panelVersion} collapsed={railCollapsed} />
        </div>
      </Layout.Sider>

      <Drawer
        placement="left"
        closable={false}
        open={drawerOpen}
        rootClassName={currentTheme}
        size="min(82vw, 320px)"
        styles={{
          wrapper: { padding: 0 },
          body: { padding: 0, display: 'flex', flexDirection: 'column', height: '100%' },
          header: { display: 'none' },
        }}
        onClose={() => setDrawerOpen(false)}
      >
        <div className="drawer-header">
          <div className="brand-block">
            <SaliHostBrand />
          </div>
          <div className="drawer-header-actions">
            <DocsButton
              ariaLabel={t('menu.docs') || 'Documentation'}
              onClick={() => navigate('/api-docs')}
            />
            <ThemeCycleButton
              id="theme-cycle-drawer"
              isDark={isDark}
              isUltra={isUltra}
              onCycle={() => cycleTheme('theme-cycle-drawer')}
              ariaLabel={t('menu.theme')}
            />
            <button
              className="drawer-close"
              type="button"
              aria-label={t('close')}
              onClick={() => setDrawerOpen(false)}
            >
              <CloseOutlined />
            </button>
          </div>
        </div>
        <button
          type="button"
          className="sidebar-command-trigger"
          onClick={() => {
            setDrawerOpen(false);
            openCommandPalette();
          }}
          aria-label={t('commandPalette.title') || 'Command Palette (Ctrl + K)'}
          style={{ margin: '8px 12px 4px', width: 'calc(100% - 24px)' }}
        >
          <span style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <SearchOutlined className="sidebar-command-icon" />
            <span>{t('commandPalette.search') || 'Search...'}</span>
          </span>
          <span className="sidebar-command-kbd">
            <span className="kbd-cmd">{SHORTCUT_MODIFIER}</span>
            <span className="kbd-key">K</span>
          </span>
        </button>
        <Menu
          theme={currentTheme}
          mode="inline"
          selectedKeys={[selectedKey]}
          openKeys={openKeys}
          onOpenChange={(keys) => setOpenKeys(keys as string[])}
          className="drawer-menu drawer-nav"
          items={groupedMenuItems}
          onClick={(info) => {
            onMenuClick(info);
            setDrawerOpen(false);
          }}
        />
        <Menu
          theme={currentTheme}
          mode="inline"
          selectedKeys={[selectedKey]}
          className="drawer-menu drawer-utility"
          items={toMenuItems(utilItems)}
          onClick={(info) => {
            onMenuClick(info);
            setDrawerOpen(false);
          }}
        />
        <div className="drawer-footer">
          <AccentPicker />
          <VersionBadge version={panelVersion} />
        </div>
      </Drawer>

      {!drawerOpen && (
        <button
          className="drawer-handle"
          type="button"
          aria-label={t('menu.openMenu')}
          onClick={() => setDrawerOpen(true)}
        >
          <MenuOutlined />
        </button>
      )}
    </div>
  );
}
