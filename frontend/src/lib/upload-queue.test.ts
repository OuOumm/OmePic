import { describe, expect, it } from 'vitest';
import { createProgressReporter, runWithConcurrency } from './upload-queue';

async function flushPromises() {
  await Promise.resolve();
  await Promise.resolve();
}

describe('runWithConcurrency', () => {
  it('limits concurrent tasks while preserving result order', async () => {
    let active = 0;
    let maxActive = 0;
    const release: Array<() => void> = [];

    const tasks = [1, 2, 3, 4].map((value) => async () => {
      active += 1;
      maxActive = Math.max(maxActive, active);
      await new Promise<void>((resolve) => release.push(resolve));
      active -= 1;
      return value * 10;
    });

    const promise = runWithConcurrency(tasks, 2);
    await flushPromises();
    expect(maxActive).toBe(2);
    expect(active).toBe(2);

    release.shift()?.();
    await flushPromises();
    expect(active).toBe(2);

    while (release.length) {
      release.shift()?.();
      await flushPromises();
    }

    await expect(promise).resolves.toEqual([10, 20, 30, 40]);
    expect(maxActive).toBe(2);
  });

  it('runs at least one task even with an invalid concurrency value', async () => {
    await expect(runWithConcurrency([async () => 'done'], 0)).resolves.toEqual(['done']);
  });
});

describe('createProgressReporter', () => {
  it('emits only changed percentages and always allows completion', () => {
    const values: number[] = [];
    const report = createProgressReporter((value) => values.push(value));

    report(1);
    report(1);
    report(2);
    report(100);
    report(100);

    expect(values).toEqual([1, 2, 100]);
  });
});
