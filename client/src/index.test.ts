import { describe as group, expect, it } from 'vitest';
import { VERSION, describe } from './index';

group('describe', () => {
  it('keeps the role it is given', () => {
    expect(describe('viewer')).toBe(`viewer/${VERSION}`);
  });

  it('reports a blank role instead of dropping it', () => {
    expect(describe('   ')).toBe(`unknown/${VERSION}`);
  });

  it('reports an empty role instead of dropping it', () => {
    expect(describe('')).toBe(`unknown/${VERSION}`);
  });
});
