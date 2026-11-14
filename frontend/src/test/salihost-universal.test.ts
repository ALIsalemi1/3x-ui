import { describe, expect, test } from 'vitest';

import globalSkin from '@/styles/salihost-universal.css?raw';
import oldSkin from '@/styles/salihost-theme.css?raw';
import panelEntry from '@/main.tsx?raw';
import loginEntry from '@/entries/login.tsx?raw';
import subscriptionEntry from '@/entries/subpage.tsx?raw';

describe('SaliHost complete frontend theme coverage', () => {
  test('every Vite entry point loads the shared skin', () => {
    for (const entry of [panelEntry, loginEntry, subscriptionEntry]) {
      expect(entry).toContain("import '@/styles/salihost-universal.css'");
      expect(entry).toContain("import '@/styles/salihost-theme.css'");
    }
  });

  test('all panel routes and the public subscription page have themed surfaces', () => {
    const pageClasses = [
      'index-page',
      'inbounds-page',
      'clients-page',
      'groups-page',
      'hosts-page',
      'nodes-page',
      'settings-page',
      'xray-page',
      'admins-page',
      'api-docs-page',
      'reseller-page',
      'multiply-page',
      'telegram-bot-page',
      'subscription-page',
    ];
    for (const page of pageClasses) {
      expect(globalSkin, 'missing universal theme for ' + page).toContain('.' + page);
    }
    expect(oldSkin).toContain('.salihost-overview-banner');
  });

  test('portal-based menus, forms and transient dialogs are included', () => {
    const widgets = [
      '.ant-modal',
      '.ant-drawer',
      '.ant-popover-inner',
      '.ant-dropdown-menu',
      '.ant-select-dropdown',
      '.ant-picker-dropdown',
      '.ant-notification-notice',
      '.ant-message-notice-content',
      '.ant-table-wrapper',
      '.ant-tabs',
      '.ant-input',
      '.ant-switch',
      '.ant-pagination-item',
      '.command-palette-modal',
      '.swagger-ui',
      '.docs-wrapper',
    ];
    for (const widget of widgets) {
      expect(globalSkin, 'missing themed widget: ' + widget).toContain(widget);
    }
  });

  test('all color modes are defined and danger actions retain native styling', () => {
    expect(globalSkin).toContain("html[data-panel-brand='salihost'] body");
    expect(globalSkin).toContain("html[data-panel-brand='salihost'] body.dark");
    expect(globalSkin).toContain("html[data-panel-brand='salihost'][data-theme='ultra-dark'] body.dark");
    expect(globalSkin).toContain('.ant-btn-dangerous');
    expect(globalSkin).toContain('.ant-tag-error');
    expect(globalSkin).toContain('@media (max-width: 768px)');
    const openBraces = (globalSkin.match(/{/g) || []).length;
    const closeBraces = (globalSkin.match(/}/g) || []).length;
    expect(openBraces).toBe(closeBraces);
  });
});
