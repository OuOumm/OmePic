import { describe, expect, it } from 'vitest';
import { getInitialThemeScriptTheme, initialThemeScript, markdownSummaryText } from './utils';

describe('getInitialThemeScriptTheme', () => {
  it('defaults missing, invalid, and corrupt stored themes to the current system theme', () => {
    expect(getInitialThemeScriptTheme(null, true)).toBe('dark');
    expect(getInitialThemeScriptTheme(null, false)).toBe('light');
    expect(getInitialThemeScriptTheme('{"theme":"unknown"}', true)).toBe('dark');
    expect(getInitialThemeScriptTheme('not-json', false)).toBe('light');
  });

  it('resolves system theme from the current media query', () => {
    expect(getInitialThemeScriptTheme('{"theme":"system"}', true)).toBe('dark');
    expect(getInitialThemeScriptTheme('{"theme":"system"}', false)).toBe('light');
  });

  it('keeps explicit light and dark themes independent of the system theme', () => {
    expect(getInitialThemeScriptTheme('{"theme":"light"}', true)).toBe('light');
    expect(getInitialThemeScriptTheme('{"theme":"dark"}', false)).toBe('dark');
  });
});

describe('initialThemeScript', () => {
  it('uses the shared preference storage key and dark class toggle', () => {
    const script = initialThemeScript();

    expect(script).toContain('omepic-ui-preferences');
    expect(script).toContain("prefers-color-scheme: dark");
    expect(script).toContain("classList.toggle('dark'");
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
