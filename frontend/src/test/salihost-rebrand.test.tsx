import { render, screen } from '@testing-library/react';
import { expect, test } from 'vitest';

import SaliHostSectionHeader from '@/components/ui/SaliHostSectionHeader';

test('each SaliHost section has a distinct product identity', () => {
  const { rerender } = render(<SaliHostSectionHeader section="clients" />);
  expect(screen.getByText('SALIHOST.IR / CUSTOMER HUB')).not.toBeNull();
  expect(screen.getByText('01')).not.toBeNull();
  rerender(<SaliHostSectionHeader section="nodes" />);
  expect(screen.getByText('SALIHOST.IR / CLOUD NODES')).not.toBeNull();
  expect(screen.getByText('03')).not.toBeNull();
});
