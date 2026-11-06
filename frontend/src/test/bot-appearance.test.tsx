import { useState } from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { expect, test } from 'vitest';
import { ThemeProvider } from '@/hooks/useTheme';
import BotAppearanceEditor, { emptyPresentation } from '@/pages/telegrambot/BotAppearanceEditor';

function Editor() {
  const [value, setValue] = useState(emptyPresentation);
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
  fireEvent.click(container.querySelector('.bot-save-bar button')!);
  const saved = JSON.parse(screen.getByTestId('saved').textContent!);
  expect(saved.brandName).toBe('Salemi shop');
  expect(saved.labels['fa:buy']).toBe('خرید سریع');
});
