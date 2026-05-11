export async function runWithConcurrency<T>(tasks: Array<() => Promise<T>>, concurrency: number): Promise<T[]> {
  const limit = Math.max(1, Math.trunc(Number.isFinite(concurrency) ? concurrency : 1));
  const results: T[] = new Array(tasks.length);
  let nextIndex = 0;

  async function worker() {
    while (nextIndex < tasks.length) {
      const currentIndex = nextIndex;
      nextIndex += 1;
      results[currentIndex] = await tasks[currentIndex]();
    }
  }

  const workers = Array.from({ length: Math.min(limit, tasks.length) }, () => worker());
  await Promise.all(workers);
  return results;
}

export function createProgressReporter(onProgress: (progress: number) => void): (progress: number) => void {
  let lastProgress = -1;
  return (progress: number) => {
    const normalized = Math.max(0, Math.min(100, Math.round(progress)));
    if (normalized === lastProgress) return;
    lastProgress = normalized;
    onProgress(normalized);
  };
}
