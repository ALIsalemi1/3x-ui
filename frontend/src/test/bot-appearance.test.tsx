import { useState } from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { expect, test } from 'vitest';
import { ThemeProvider } from '@/hooks/useTheme';
import BotAppearanceEditor, { emptyPresentation } from '@/pages/telegrambot/BotAppearanceEditor';

function Editor({ initial = emptyPresentation }: { initial?: typeof emptyPresentation }) {
  const [value, setValue] = useState(initial);
  const [welcome, setWelcome] = useState('سلام');
  const [saved, setSaved] = useState('');
  return (
    <ThemeProvider>
      <BotAppearanceEditor
        value={value}
        onChange={setValue}
        welcome={welcome}
        onWelcomeChange={setWelcome}
        username="salemi_bot"
        saving={false}
        onSave={() => setSaved(JSON.stringify(value))}
      />
      <output data-testid="saved">{saved}</output>
    </ThemeProvider>
  );
}

test('shop identity and translated menu edits appear in preview and save payload', () => {
  const { container } = render(<Editor />);
  fireEvent.change(container.querySelector('#bot-brandName')!, {
    target: { value: 'Salemi shop' },
  });
  fireEvent.change(container.querySelector('#bot-label-buy')!, { target: { value: 'خرید سریع' } });
  expect(container.querySelector('.bot-phone-header strong')?.textContent).toBe('Salemi shop');
  expect(container.querySelector('.bot-phone-menu')?.textContent).toContain('خرید سریع');
  expect(container.querySelector('.bot-shop-banner')?.getAttribute('alt')).toBe('VPN SHOP');
  expect(
    container.querySelector('.bot-phone-menu-row .bot-menu-button--success')?.textContent,
  ).toBe('خرید سریع');
  expect(
    container.querySelector('.bot-phone-menu .bot-menu-button--danger')?.textContent,
  ).toContain('پشتیبانی');
  fireEvent.click(container.querySelector('.bot-save-bar button')!);
  const saved = JSON.parse(screen.getByTestId('saved').textContent!);
  expect(saved.brandName).toBe('Salemi shop');
  expect(saved.labels['fa:buy']).toBe('خرید سریع');
});

// Preview colors are derived from the same overrides sent to Telegram.
test('custom native button colors appear in preview and save payload', () => {
  const { container } = render(
    <Editor
      initial={{
        ...emptyPresentation,
        buttonStyles: { buy: 'danger', trial: 'success', account: 'default' },
      }}
    />,
  );
  const preview = container.querySelector('.bot-phone-menu');
  expect(preview?.querySelector('.bot-menu-button--danger')?.textContent).toContain('خرید');
  expect(
    preview?.querySelector('.bot-phone-menu-row:nth-child(3) .bot-menu-button--success')
      ?.textContent,
  ).toContain('تست');
  expect(preview?.querySelector('.bot-menu-button--default')?.textContent).toContain('حساب');
  fireEvent.click(container.querySelector('.bot-save-bar button')!);
  const saved = JSON.parse(screen.getByTestId('saved').textContent!);
  expect(saved.buttonStyles).toEqual({
    buy: 'danger',
    trial: 'success',
    account: 'default',
  });
});

test('extended button accents render emoji on neutral Telegram buttons and are saved', () => {
  const { container } = render(
    <Editor
      initial={{
        ...emptyPresentation,
        buttonStyles: { buy: 'pink', trial: 'yellow', subs: 'purple', support: 'orange' },
        labels: { 'fa:buy': '🛒 خرید ویژه' },
      }}
    />,
  );
  const menu = container.querySelector('.bot-phone-menu');
  const neutral = Array.from(menu?.querySelectorAll('.bot-menu-button--default') || []);
  expect(neutral.map((button) => button.textContent)).toEqual(
    expect.arrayContaining([
      '🩷 🛒 خرید ویژه',
      '🟡 🆓 دریافت اشتراک تست',
      '🟣 💎 اشتراک‌های من',
      '🟠 📞 پشتیبانی',
    ]),
  );
  expect(menu?.querySelector('.bot-menu-button--success')).toBeNull();
  fireEvent.click(container.querySelector('.bot-save-bar button')!);
  const saved = JSON.parse(screen.getByTestId('saved').textContent!);
  expect(saved.buttonStyles.buy).toBe('pink');
  expect(saved.labels['fa:buy']).toBe('🛒 خرید ویژه');
});

test('existing color indicator in label is not duplicated', () => {
  const { container } = render(
    <Editor
      initial={{
        ...emptyPresentation,
        buttonStyles: { buy: 'purple' },
        labels: { 'fa:buy': '🟣 پیشنهاد ویژه' },
      }}
    />,
  );
  const button = container.querySelector('.bot-phone-menu-row .bot-menu-button--default');
  expect(button?.textContent).toBe('🟣 پیشنهاد ویژه');
});

test('editor and saved presentation contain no Telegram Mini App setting', () => {
  const { container } = render(<Editor />);
  expect(container.querySelector('#bot-miniapp-url')).toBeNull();
  expect(container.querySelector('.bot-phone-menu')?.textContent).not.toContain('Graphical shop');
  fireEvent.click(container.querySelector('.bot-save-bar button')!);
  expect(JSON.parse(screen.getByTestId('saved').textContent!)).not.toHaveProperty('miniAppURL');
});

test('default colorful home has four usable rows and correct accent markers', () => {
  const { container } = render(<Editor />);
  const rows = container.querySelectorAll('.bot-phone-menu-row');
  expect(rows).toHaveLength(4);
  expect(rows[0].textContent).toContain('خرید اشتراک جدید');
  expect(rows[1].textContent).toContain('🟣');
  expect(rows[1].textContent).toContain('🩷');
  expect(rows[2].textContent).toContain('🟡');
  expect(rows[2].textContent).toContain('🩵');
});

test('one-click rainbow preset changes actual saved Telegram button styles', () => {
  const { container } = render(<Editor />);
  fireEvent.click(screen.getByRole('button', { name: /رنگین‌کمانی/ }));
  const menu = container.querySelector('.bot-phone-menu');
  expect(menu?.textContent).toContain('🟣');
  expect(menu?.textContent).toContain('🩷');
  fireEvent.click(container.querySelector('.bot-save-bar button')!);
  const saved = JSON.parse(screen.getByTestId('saved').textContent!);
  expect(saved.buttonStyles).toMatchObject({
    buy: 'success',
    trial: 'yellow',
    invoices: 'pink',
    subs: 'purple',
    account: 'cyan',
    support: 'danger',
    tutorial: 'orange',
  });
  expect(saved).not.toHaveProperty('miniAppURL');
});
