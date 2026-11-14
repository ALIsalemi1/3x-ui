import { render } from '@testing-library/react';
import { expect, test } from 'vitest';

import { ThemeProvider, buildAntdThemeConfig, PANEL_ACCENTS } from '@/hooks/useTheme';

test('SaliHost is the default panel accent without removing existing colors', () => {
  expect(PANEL_ACCENTS.salihost).toBe('#087f91');
  expect(PANEL_ACCENTS.ocean).toBeDefined();
  expect(PANEL_ACCENTS.lagoon).toBeDefined();
  expect(PANEL_ACCENTS.violet).toBeDefined();

  const light = buildAntdThemeConfig(false, false);
  expect(light.token?.colorPrimary).toBe(PANEL_ACCENTS.salihost);
  expect(light.token?.colorBgLayout).toBe('#eef5f8');

  const dark = buildAntdThemeConfig(true, false);
  expect(dark.token?.colorPrimary).toBe(PANEL_ACCENTS.salihost);
  expect(dark.token?.colorBgContainer).toBe('#112437');

  const ultra = buildAntdThemeConfig(true, true);
  expect(ultra.token?.colorBgBase).toBe('#000');
  expect(ultra.token?.colorBgContainer).toBe('#101013');
});

test('previous accent choice still renders with its original palette', () => {
  const classic = buildAntdThemeConfig(false, false, 'ocean');
  expect(classic.token?.colorPrimary).toBe('#1675e8');
  expect(classic.token?.colorBgLayout).not.toBe('#eef5f8');
});

test('SaliHost identity marker is applied without removing accessibility theme state', () => {
  const { unmount } = render(
    <ThemeProvider>
      <span>SaliHost</span>
    </ThemeProvider>,
  );
  expect(document.documentElement.getAttribute('data-panel-brand')).toBe('salihost');
  expect(document.documentElement.style.colorScheme).toMatch(/^(dark|light)$/);
  unmount();
});
