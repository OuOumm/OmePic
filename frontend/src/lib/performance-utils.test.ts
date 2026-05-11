import { describe, expect, it } from 'vitest';
import { getInitialThemeScriptTheme, markdownSummaryText } from './utils';

describe('getInitialThemeScriptTheme', () => {
  it('defaults missing and invalid stored themes to light', () => {
    expect(getInitialThemeScriptTheme(null, false)).toBe('light');
    expect(getInitialThemeScriptTheme('{"theme":"unknown"}', true)).toBe('light');
    expect(getInitialThemeScriptTheme('not-json', true)).toBe('light');
  });

  it('resolves system theme from the current media query', () => {
    expect(getInitialThemeScriptTheme('{"theme":"system"}', true)).toBe('dark');
    expect(getInitialThemeScriptTheme('{"theme":"system"}', false)).toBe('light');
  });
});

describe('markdownSummaryText', () => {
  it('strips markdown markers and collapses whitespace for summaries', () => {
    expect(markdownSummaryText('## Hello **world**\n\n- one\n- two')).toBe('Hello world one two');
  });

  it('limits long summaries without splitting surrogate pairs', () => {
    expect(markdownSummaryText('abcdef', 4)).toBe('abcd…');
  });
});
