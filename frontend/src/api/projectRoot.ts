export const DEFAULT_ALPHA_PROJECT_ROOT = 'examples/alpha-project'

export function normalizeProjectRoot(root?: string): string {
  const trimmed = root?.trim()
  return trimmed || DEFAULT_ALPHA_PROJECT_ROOT
}
