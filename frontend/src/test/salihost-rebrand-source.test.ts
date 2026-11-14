import { readFileSync } from 'node:fs';
import { expect, test } from 'vitest';

test('custom login and dashboard routes have working branded navigation', () => {
  const login = readFileSync(new URL('../pages/login/LoginPage.tsx', import.meta.url), 'utf8');
  const index = readFileSync(new URL('../pages/index/IndexPage.tsx', import.meta.url), 'utf8');
  const sidebar = readFileSync(new URL('../layouts/AppSidebar.tsx', import.meta.url), 'utf8');
  expect(login).toContain('salihost-auth-shell');
  expect(login).toContain('salihost-login-story');
  expect(index).toContain('salihost-quicknav');
  expect(index).toContain('navigate(link.route)');
  expect(sidebar).toContain('salihost-nav-heading');
  expect(sidebar).toContain('groupedMenuItems');
});

test('brand rework keeps open-source credits and its GPL license discoverable', () => {
  const sidebar = readFileSync(new URL('../layouts/AppSidebar.tsx', import.meta.url), 'utf8');
  expect(sidebar).toContain('Open-source credits &amp; licenses');
  expect(sidebar).toContain('MHSanaei');
  expect(sidebar).toContain('GPL-3.0 license');
  expect(sidebar).toContain('About SaliHost Control Center');
});

test('installable app uses only SaliHost icons', () => {
  const manifest = JSON.parse(
    readFileSync(new URL('../../public/manifest.webmanifest', import.meta.url), 'utf8'),
  );
  expect(manifest.name).toContain('SaliHost');
  const paths: string[] = manifest.icons.map((icon: { src: string }) => icon.src);
  expect(paths).toContain('icons/salihost-192.png');
  expect(paths).toContain('icons/salihost-512.png');
  expect(paths).not.toContain('icons/3x-ui-512.png');
});
